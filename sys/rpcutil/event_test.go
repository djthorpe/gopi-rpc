package rpcutil_test

import (
	"testing"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"

	// Modules
	util "github.com/djthorpe/gopi-rpc/sys/rpcutil"
	_ "github.com/djthorpe/gopi/sys/logger"
)

func Test_Event_001(t *testing.T) {
	if driver, err := gopi.Open(util.Util{}, nil); err != nil {
		t.Error(err)
	} else {
		defer driver.Close()
		t.Log(driver)
	}
}

func Test_Event_002(t *testing.T) {
	config := gopi.NewAppConfig("rpc/util")
	config.Debug = true
	config.Verbose = true
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Error(err)
	} else {
		defer app.Close()
		t.Log(app)
	}
}

func Test_Event_003(t *testing.T) {
	config := gopi.NewAppConfig("rpc/util")
	config.Debug = true
	config.Verbose = true
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Error(err)
	} else {
		defer app.Close()
		if util, ok := app.ModuleInstance("rpc/util").(rpc.Util); ok == false {
			t.Fatal("Cannot cast to rpc/util", util)
		} else {
			t.Log(util)
		}
	}
}
func Test_Event_004(t *testing.T) {
	config := gopi.NewAppConfig("rpc/util")
	config.Debug = true
	config.Verbose = true
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Error(err)
	} else {
		defer app.Close()
		if util, ok := app.ModuleInstance("rpc/util").(rpc.Util); ok == false {
			t.Fatal("Cannot cast to rpc/util", util)
		} else if evt := util.NewEvent(nil, gopi.RPC_EVENT_NONE, nil); evt == nil {
			t.Error("Unexpected evt == nil")
		} else if evt.Source() != nil {
			t.Error("Unexpected source != nil")
		} else if evt.Type() != gopi.RPC_EVENT_NONE {
			t.Error("Unexpected type != 0")
		} else if evt.ServiceRecord() != nil {
			t.Error("Unexpected service record != nil")
		} else {
			t.Log(evt)
		}
	}
}
