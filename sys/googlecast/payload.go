/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package googlecast

// Ref: https://github.com/vishen/go-chromecast/

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Payload interface {
	WithId(id int) Payload
}

type PayloadHeader struct {
	Type      string `json:"type"`
	RequestId int    `json:"requestId,omitempty"`
}

type MediaHeader struct {
	PayloadHeader
	MediaSessionId int     `json:"mediaSessionId"`
	CurrentTime    float32 `json:"currentTime"`
	RelativeTime   float32 `json:"relativeTime,omitempty"`
	ResumeState    string  `json:"resumeState"`
}

type VolumeHeader struct {
	PayloadHeader
	Volume volume `json:"volume"`
}

type ReceiverStatusResponse struct {
	PayloadHeader
	Status struct {
		Applications []application `json:"applications"`
		Volume       volume        `json:"volume"`
	} `json:"status"`
}

type DeviceUpdatedResponse struct {
	PayloadHeader
	Device struct {
		DeviceId string `json:"deviceId"`
		Name     string `json:"name"`
		Volume   volume `json:"volume"`
	} `json:"device"`
}

type MediaStatusResponse struct {
	PayloadHeader
	Status []media `json:"status"`
}

////////////////////////////////////////////////////////////////////////////////
// GLOBAL VARIABLES

var (
	// Known Payload headers
	LaunchHeader      = PayloadHeader{Type: "LAUNCH"}       // Launches a new chromecast app
	SeekHeader        = PayloadHeader{Type: "SEEK"}         // Seek into the running app
	LoadHeader        = PayloadHeader{Type: "LOAD"}         // Loads an application onto the chromecast
	QueueLoadHeader   = PayloadHeader{Type: "QUEUE_LOAD"}   // Loads an application onto the chromecast
	QueueUpdateHeader = PayloadHeader{Type: "QUEUE_UPDATE"} // Loads an application onto the chromecast
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

func (this *PayloadHeader) WithId(id int) Payload {
	this.RequestId = id
	return this
}

func (this *MediaHeader) WithId(id int) Payload {
	this.PayloadHeader.RequestId = id
	return this
}

func (this *VolumeHeader) WithId(id int) Payload {
	this.PayloadHeader.RequestId = id
	return this
}
