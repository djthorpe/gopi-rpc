/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package googlecast

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	event "github.com/djthorpe/gopi/util/event"
	proto "github.com/gogo/protobuf/proto"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/rpc/protobuf/googlecast"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type CastConn struct {
	Addr    string
	Port    uint16
	Timeout time.Duration
}

type castconn struct {
	conn      *tls.Conn
	timeout   time.Duration
	log       gopi.Logger
	messageid int

	// The current status
	applications  []application
	volume_status *volume
	current       *application
	media         *media

	// Tasks, pubsub and locking
	event.Tasks
	event.Publisher
	sync.Mutex
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	DEFAULT_TIMEOUT       = 30 * time.Second
	READ_TIMEOUT          = 500 * time.Millisecond
	STATUS_INTERVAL       = 10 * time.Second
	CAST_DEFAULT_SENDER   = "sender-0"
	CAST_DEFAULT_RECEIVER = "receiver-0"
	CAST_NS_CONN          = "urn:x-cast:com.google.cast.tp.connection"
	CAST_NS_RECV          = "urn:x-cast:com.google.cast.receiver"
	CAST_NS_MEDIA         = "urn:x-cast:com.google.cast.media"
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config CastConn) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<googlecast.conn.Open>{ %+v }", config)

	this := new(castconn)
	this.log = log
	if config.Timeout == 0 {
		this.timeout = DEFAULT_TIMEOUT
	} else {
		this.timeout = config.Timeout
	}

	addrport := fmt.Sprintf("%s:%d", config.Addr, config.Port)
	if conn, err := tls.DialWithDialer(&net.Dialer{
		Timeout:   this.timeout,
		KeepAlive: this.timeout,
	}, "tcp", addrport, &tls.Config{
		InsecureSkipVerify: true,
	}); err != nil {
		return nil, fmt.Errorf("%s: %w", addrport, err)
	} else {
		this.conn = conn
	}

	// Task to receive messages
	this.Tasks.Start(this.ReceiveTask)

	// Success
	return this, nil
}

func (this *castconn) Close() error {
	this.log.Debug("<googlecast.conn.Close>{ remote_addr=%v }", strconv.Quote(this.RemoteAddr()))

	// Kill background tasks
	if err := this.Tasks.Close(); err != nil {
		return err
	}

	// Unsubscribe
	this.Publisher.Close()

	// Close connection
	if this.conn != nil {
		if err := this.conn.Close(); err != nil {
			return err
		}
	}

	// Release resoruces
	this.conn = nil
	this.current = nil
	this.media = nil
	this.applications = nil
	this.volume_status = nil

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *castconn) String() string {
	return fmt.Sprintf("<googlecast.conn>{ remote_addr=%v }", strconv.Quote(this.RemoteAddr()))
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *castconn) LocalAddr() string {
	if this.conn != nil {
		return this.conn.LocalAddr().String()
	} else {
		return "<nil>"
	}
}

func (this *castconn) RemoteAddr() string {
	if this.conn != nil {
		return this.conn.RemoteAddr().String()
	} else {
		return "<nil>"
	}
}

func (this *castconn) Applications() []rpc.GoogleCastApplication {
	if this.conn == nil || this.applications == nil {
		return nil
	}
	apps := make([]rpc.GoogleCastApplication, len(this.applications))
	for i, app := range this.applications {
		apps[i] = &app
	}
	return apps
}

func (this *castconn) Volume() rpc.GoogleCastVolume {
	if this.conn == nil || this.volume_status == nil {
		return nil
	} else {
		return this.volume_status
	}
}

func (this *castconn) Media() rpc.GoogleCastMedia {
	if this.conn == nil || this.media == nil {
		return nil
	} else {
		return this.media
	}
}

////////////////////////////////////////////////////////////////////////////////
// SEND CONNECT AND DISCONNECT MESSAGES

func (this *castconn) Connect() error {
	payload := &PayloadHeader{Type: "CONNECT"}
	// Connect to "conn"
	if err := this.send(CAST_DEFAULT_SENDER, CAST_DEFAULT_RECEIVER, CAST_NS_CONN, payload.WithId(this.nextMessageId())); err != nil {
		return err
	} else {
		go func() {
			this.Emit(NewChannelEvent(this, rpc.GOOGLE_CAST_EVENT_CONNECT, this))
		}()
	}
	// Release resources
	this.media = nil
	this.current = nil
	this.applications = nil
	this.volume_status = nil
	// Success
	return nil
}

func (this *castconn) Disconnect() error {
	payload := &PayloadHeader{Type: "CLOSE"}

	// Disconnect from "conn"
	if err := this.send(CAST_DEFAULT_SENDER, CAST_DEFAULT_RECEIVER, CAST_NS_CONN, payload.WithId(this.nextMessageId())); err != nil {
		return err
	} else {
		go func() {
			this.Emit(NewChannelEvent(this, rpc.GOOGLE_CAST_EVENT_DISCONNECT, this))
		}()
	}
	// Release resources
	this.media = nil
	this.current = nil
	this.applications = nil
	this.volume_status = nil
	// Success
	return nil
}

func (this *castconn) ConnectMedia(app rpc.GoogleCastApplication) error {
	payload := &PayloadHeader{Type: "CONNECT"}

	if app_, ok := app.(*application); app_ == nil || ok == false {
		return gopi.ErrBadParameter
	} else if err := this.send(CAST_DEFAULT_SENDER, app_.TransportId, CAST_NS_CONN, payload.WithId(this.nextMessageId())); err != nil {
		return err
	}
	// success
	return nil
}

func (this *castconn) DisconnectMedia(app rpc.GoogleCastApplication) error {
	payload := &PayloadHeader{Type: "CLOSE"}

	if app_, ok := app.(*application); app_ == nil || ok == false {
		return gopi.ErrBadParameter
	} else if err := this.send(CAST_DEFAULT_SENDER, app_.TransportId, CAST_NS_CONN, payload.WithId(this.nextMessageId())); err != nil {
		return err
	}
	// success
	return nil
}

func (this *castconn) GetMediaStatus(app rpc.GoogleCastApplication) (int, error) {
	payload := &PayloadHeader{Type: "GET_STATUS"}

	if app_, ok := app.(*application); app_ == nil || ok == false {
		return 0, gopi.ErrBadParameter
	} else if err := this.send(CAST_DEFAULT_SENDER, app_.TransportId, CAST_NS_MEDIA, payload.WithId(this.nextMessageId())); err != nil {
		return 0, err
	} else {
		return payload.RequestId, nil
	}
}

func (this *castconn) GetReceiverStatus() (int, error) {
	payload := &PayloadHeader{Type: "GET_STATUS"}
	if err := this.send(CAST_DEFAULT_SENDER, CAST_DEFAULT_RECEIVER, CAST_NS_RECV, payload.WithId(this.nextMessageId())); err != nil {
		return 0, err
	} else {
		return payload.RequestId, nil
	}
}

func (this *castconn) SetApplication(app rpc.GoogleCastApplication) error {
	this.log.Debug2("<googlecast.conn.SetApplication>{ app=%v }", app)

	if app == nil && this.current != nil {
		// Disconnect
		err := this.DisconnectMedia(app)
		this.current = nil
		return err
	} else if app_ := this.appWithId(app.ID()); app_ == nil {
		return gopi.ErrNotFound
	} else if this.current != nil && app_.ID() == this.current.ID() {
		// Nothing to do
		return nil
	} else if err := this.ConnectMedia(app_); err != nil {
		return err
	} else if _, err := this.GetMediaStatus(app_); err != nil {
		return err
	} else {
		if this.current != nil {
			if err := this.DisconnectMedia(this.current); err != nil {
				this.log.Warn("DisconnectMedia: %v", err)
			}
		}
		this.current = app_
	}

	// Return success
	return nil
}

func (this *castconn) SetPause(state bool) (int, error) {
	this.log.Debug2("<googlecast.conn.SetPause>{ state=%v }", state)

	if this.media == nil || this.current == nil {
		return 0, gopi.ErrOutOfOrder
	}
	payload := MediaHeader{
		PayloadHeader:  PayloadHeader{Type: "PAUSE"},
		MediaSessionId: this.media.MediaSessionId,
	}
	if state == false {
		payload.PayloadHeader.Type = "PLAY"
	}
	if err := this.send(CAST_DEFAULT_SENDER, this.current.TransportId, CAST_NS_MEDIA, payload.WithId(this.nextMessageId())); err != nil {
		return 0, err
	} else {
		return payload.PayloadHeader.RequestId, nil
	}
}

func (this *castconn) SetPlay(state bool) (int, error) {
	this.log.Debug2("<googlecast.conn.SetPlay>{ state=%v }", state)

	if this.media != nil && this.current != nil {
		payload := MediaHeader{
			PayloadHeader:  PayloadHeader{Type: "PLAY"},
			MediaSessionId: this.media.MediaSessionId,
		}
		if state == false {
			payload.PayloadHeader.Type = "STOP"
		}
		if err := this.send(CAST_DEFAULT_SENDER, this.current.TransportId, CAST_NS_MEDIA, payload.WithId(this.nextMessageId())); err != nil {
			return 0, err
		} else {
			return payload.PayloadHeader.RequestId, nil
		}
	} else if state == false {
		payload := PayloadHeader{Type: "STOP"}
		if err := this.send(CAST_DEFAULT_SENDER, CAST_DEFAULT_RECEIVER, CAST_NS_RECV, payload.WithId(this.nextMessageId())); err != nil {
			return 0, err
		} else {
			return payload.RequestId, nil
		}
	} else {
		return 0, gopi.ErrOutOfOrder
	}
}

func (this *castconn) SetVolume(level float32) (int, error) {
	this.log.Debug2("<googlecast.conn.SetVolume>{ level=%v }", level)

	if level > 1.0 || level < 0 {
		return 0, gopi.ErrBadParameter
	}

	payload := &VolumeHeader{
		PayloadHeader: PayloadHeader{Type: "SET_VOLUME"},
		Volume: volume{
			Level_: level,
		},
	}
	if err := this.send(CAST_DEFAULT_SENDER, CAST_DEFAULT_RECEIVER, CAST_NS_RECV, payload.WithId(this.nextMessageId())); err != nil {
		return 0, err
	} else {
		return payload.PayloadHeader.RequestId, nil
	}
}

func (this *castconn) SetMuted(state bool) (int, error) {
	this.log.Debug2("<googlecast.conn.SetMuted>{ state=%v }", state)

	payload := &VolumeHeader{
		PayloadHeader: PayloadHeader{Type: "SET_VOLUME"},
		Volume: volume{
			Muted_: state,
		},
	}
	if err := this.send(CAST_DEFAULT_SENDER, CAST_DEFAULT_RECEIVER, CAST_NS_RECV, payload.WithId(this.nextMessageId())); err != nil {
		return 0, err
	} else {
		return payload.PayloadHeader.RequestId, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// SEND MESSAGES

func (this *castconn) send(source, dest, ns string, payload Payload) error {
	this.log.Debug2("<googlecast.conn.send>{ source=%v dest=%v ns=%v payload=%v }", strconv.Quote(source), strconv.Quote(dest), strconv.Quote(ns), payload)

	if json, err := json.Marshal(payload); err != nil {
		return err
	} else {
		payload_str := string(json)
		message := &pb.CastMessage{
			ProtocolVersion: pb.CastMessage_CASTV2_1_0.Enum(),
			SourceId:        &source,
			DestinationId:   &dest,
			Namespace:       &ns,
			PayloadType:     pb.CastMessage_STRING.Enum(),
			PayloadUtf8:     &payload_str,
		}
		proto.SetDefaults(message)
		if data, err := proto.Marshal(message); err != nil {
			return err
		} else if err := binary.Write(this.conn, binary.BigEndian, uint32(len(data))); err != nil {
			return err
		} else if _, err := this.conn.Write(data); err != nil {
			return err
		}
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// BACKGROUND TASKS

func (this *castconn) ReceiveTask(start chan<- event.Signal, stop <-chan event.Signal) error {
	status := time.NewTimer(500 * time.Millisecond)
	start <- gopi.DONE
FOR_LOOP:
	for {
		select {
		case <-status.C:
			// Update receiver status if empty
			if this.applications == nil || this.volume_status == nil {
				if _, err := this.GetReceiverStatus(); err != nil {
					this.log.Error("ReceiveTask: %v", err)
				}
			} else if this.media == nil && this.current != nil {
				if _, err := this.GetMediaStatus(this.current); err != nil {
					this.log.Error("ReceiveTask: %v", err)
				}
			}
			status.Reset(STATUS_INTERVAL)
		case <-stop:
			status.Stop()
			break FOR_LOOP
		default:
			var length uint32
			if err := this.conn.SetReadDeadline(time.Now().Add(READ_TIMEOUT)); err != nil {
				this.log.Error("ReceiveTask: %v", err)
			} else if err := binary.Read(this.conn, binary.BigEndian, &length); err != nil {
				if err == io.EOF || os.IsTimeout(err) {
					// Ignore error
				} else {
					this.log.Error("ReceiveTask: %v", err)
				}
			} else if length == 0 {
				this.log.Warn("ReceiveTask: Received zero-sized data")
			} else {
				payload := make([]byte, length)
				if bytes_read, err := io.ReadFull(this.conn, payload); err != nil {
					this.log.Warn("ReceiveTask: %v", err)
				} else if bytes_read != int(length) {
					this.log.Warn("ReceiveTask: Received different number of bytes %v read, expected %v", bytes_read, length)
				} else if err := this.emit(payload); err != nil {
					this.log.Warn("ReceiveTask: %v", payload)
				}
			}
		}
	}

	this.log.Debug("ReceiveTask: Stopped")

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *castconn) setVolume(volume *volume) {
	if this.volume_status == nil || this.volume_status.Equals(volume) == false {
		this.volume_status = volume
		this.Emit(NewChannelEvent(this, rpc.GOOGLE_CAST_EVENT_VOLUME, this))
	}
}

func (this *castconn) setApplications(apps []application) {
	if len(apps) > 0 {
		this.applications = apps
	} else {
		this.applications = make([]application, 0)
	}
	// Emit change of applications message
	this.Emit(NewChannelEvent(this, rpc.GOOGLE_CAST_EVENT_APPLICATION, this))
}

func (this *castconn) setMedia(media *media) {
	// Set volume first
	this.setVolume(&media.Volume)
	// Set media
	this.media = media
	// Emit media update
	this.Emit(NewChannelEvent(this, rpc.GOOGLE_CAST_EVENT_MEDIA, this))
}

func (this *castconn) nextMessageId() int {
	this.Lock()
	defer this.Unlock()

	// Cycle messages from 1 to 99999
	this.messageid = (this.messageid + 1) % 100000
	return this.messageid
}

func (this *castconn) appWithId(appId string) *application {
	for _, app := range this.applications {
		if app.AppId == appId {
			return &app
		}
	}
	// Not found
	return nil
}

func (this *castconn) emit(data []byte) error {
	var header PayloadHeader
	var receiver_status ReceiverStatusResponse
	var device_updated DeviceUpdatedResponse
	var media_status MediaStatusResponse

	message := &pb.CastMessage{}
	if err := proto.Unmarshal(data, message); err != nil {
		return err
	} else if err := json.Unmarshal([]byte(*message.PayloadUtf8), &header); err != nil {
		return err
	} else if header.Type == "PING" {
		payload := &PayloadHeader{Type: "PONG", RequestId: -1}
		if err := this.send(message.GetDestinationId(), message.GetSourceId(), message.GetNamespace(), payload); err != nil {
			return fmt.Errorf("Ping error: %w", err)
		}
	} else if header.Type == "RECEIVER_STATUS" {
		if err := json.Unmarshal([]byte(message.GetPayloadUtf8()), &receiver_status); err != nil {
			return fmt.Errorf("RECEIVER_STATUS: %w", err)
		} else {
			// Set applications
			this.setApplications(receiver_status.Status.Applications)
			// Set the volume
			this.setVolume(&receiver_status.Status.Volume)
		}
	} else if header.Type == "DEVICE_UPDATED" {
		if err := json.Unmarshal([]byte(message.GetPayloadUtf8()), &device_updated); err != nil {
			return fmt.Errorf("DEVICE_UPDATED: %w", err)
		} else {
			// Set the volume and emit if it has changed
			this.setVolume(&device_updated.Device.Volume)
		}
	} else if header.Type == "MEDIA_STATUS" {
		if err := json.Unmarshal([]byte(message.GetPayloadUtf8()), &media_status); err != nil {
			return fmt.Errorf("MEDIA_STATUS: %w", err)
		} else {
			// Update media
			for _, media := range media_status.Status {
				this.setMedia(&media)
			}
		}
	} else {
		this.log.Debug("ReceiveTask: Ignoring: %v: %v", header.Type, message)
	}

	// Success
	return nil
}
