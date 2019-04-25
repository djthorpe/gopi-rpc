package discovery_test

import (
	"testing"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"

	// Modules
	discovery "github.com/djthorpe/gopi-rpc/sys/discovery"
	_ "github.com/djthorpe/gopi/sys/logger"
)

func Test_Discovery_001(t *testing.T) {
	if driver, err := gopi.Open(discovery.Discovery{}, nil); err != nil {
		t.Error(err)
	} else {
		defer driver.Close()
		t.Log(driver)
	}
}

func Test_Discovery_002(t *testing.T) {
	config := gopi.NewAppConfig("rpc/discovery")
	config.Debug = true
	config.Verbose = true
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Error(err)
	} else {
		defer app.Close()
		t.Log(app)
		time.Sleep(time.Second * 30)
	}
}
