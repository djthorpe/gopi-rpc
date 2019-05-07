/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"fmt"
	"os"

	rpc "github.com/djthorpe/gopi-rpc"

	// Frameworks
	gopi "github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/sys/dns-sd"
	_ "github.com/djthorpe/gopi-rpc/sys/googlecast"
	_ "github.com/djthorpe/gopi-rpc/sys/rpcutil"
	_ "github.com/djthorpe/gopi/sys/logger"
)

/*
////////////////////////////////////////////////////////////////////////////////

type GoogleCast struct {
	params map[string]string
}

var (
	casts = make(map[string]*GoogleCast)
)

////////////////////////////////////////////////////////////////////////////////

func NewGoogleCast(service gopi.RPCServiceRecord) *GoogleCast {
	this := new(GoogleCast)
	this.params = make(map[string]string)
	for _, txt := range service.Text() {
		if pair := strings.SplitN(txt, "=", 2); len(pair) == 2 {
			this.params[pair[0]] = pair[1]
		}
	}
	// Validate id exists
	if value, exists := this.params["id"]; exists == false || value == "" {
		return nil
	} else {
		return this
	}
}

func (this *GoogleCast) String() string {
	return fmt.Sprintf("Cast<id=%v name=%v model=%v app=%v state=%v>", this.Id(), strconv.Quote(this.Name()), strconv.Quote(this.Model()), strconv.Quote(this.App()), this.State())
}

func (this *GoogleCast) Id() string {
	return this.params["id"]
}

func (this *GoogleCast) Model() string {
	if value, exists := this.params["md"]; exists == false || value == "" {
		return "Unknown"
	} else {
		return value
	}
	return this.params["id"]
}

func (this *GoogleCast) App() string {
	if value, exists := this.params["rs"]; exists == false || value == "" {
		return ""
	} else {
		return value
	}
}

func (this *GoogleCast) Name() string {
	if value, exists := this.params["fn"]; exists == false || value == "" {
		return "Unknown"
	} else {
		return value
	}
	return this.params["fn"]
}

func (this *GoogleCast) State() uint {
	if value, exists := this.params["st"]; exists == false || value == "" {
		return 0
	} else if value_, err := strconv.ParseUint(value, 10, 32); err != nil {
		return 0
	} else {
		return uint(value_)
	}
}

func (this *GoogleCast) EqualsState(other *GoogleCast) bool {
	if this.Model() != other.Model() {
		return false
	}
	if this.App() != other.App() {
		return false
	}
	if this.Name() != other.Name() {
		return false
	}
	if this.State() != other.State() {
		return false
	}
	return true
}

////////////////////////////////////////////////////////////////////////////////

func RenderCast(cast *GoogleCast) {
	if cast.State() != 0 {
		fmt.Printf("%-20s %-20s Now Playing: %v\n", cast.Name(), cast.Model(), cast.App())
	} else {
		fmt.Printf("%-20s %-20s Idle\n", cast.Name(), cast.Model())
	}
}

func WatchEvent(evt gopi.RPCEvent) {
	service := evt.ServiceRecord()
	if service == nil || service.Service() != SERVICE_TYPE_GOOGLECAST {
		return
	}
	if cast := NewGoogleCast(service); cast == nil {
		return
	} else if other, exists := casts[cast.Id()]; exists == false || cast.EqualsState(other) == false {
		RenderCast(cast)
		casts[cast.Id()] = cast
	} else {
		fmt.Println(cast.params)
		casts[cast.Id()] = cast
	}
}

*/

func Watch(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	googlecast := app.ModuleInstance("googlecast").(rpc.GoogleCast)
	start <- gopi.DONE

	// If there is an argument, then this is the service to lookup
	events := googlecast.Subscribe()
FOR_LOOP:
	for {
		select {
		case evt := <-events:
			if evt_, ok := evt.(rpc.GoogleCastEvent); ok {
				fmt.Println(evt_)
			}
		case <-stop:
			break FOR_LOOP
		}
	}
	googlecast.Unsubscribe(events)
	app.SendSignal()
	return nil
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	app.Logger.Info("Waiting for CTRL+C")
	app.WaitForSignal()
	done <- gopi.DONE

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("googlecast", "discovery")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main, Watch))
}
