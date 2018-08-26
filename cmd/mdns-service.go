package main

import (
	"os"

	// Frameworks
	"github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/gopi/sys/logger"
)

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	if listener, err := gopi.Open(Listener{}, app.Logger); err != nil {
		return err
	} else {
		app.Logger.Info("Waiting for CTRL+C")
		app.WaitForSignal()
		return listener.Close()
	}
}

func main() {
	config := gopi.NewAppConfig()
	os.Exit(gopi.CommandLineTool(config, Main))
}
