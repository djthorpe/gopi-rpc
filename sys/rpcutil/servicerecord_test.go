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

func Test_ServiceRecord_001(t *testing.T) {
	if driver, err := gopi.Open(util.Util{}, nil); err != nil {
		t.Error(err)
	} else {
		defer driver.Close()
		t.Log(driver)
	}
}

func Test_ServiceRecord_002(t *testing.T) {
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

func Test_ServiceRecord_003(t *testing.T) {
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
func Test_ServiceRecord_004(t *testing.T) {
	config := gopi.NewAppConfig("rpc/util")
	config.Debug = true
	config.Verbose = true
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Error(err)
	} else {
		defer app.Close()
		if util, ok := app.ModuleInstance("rpc/util").(rpc.Util); ok == false {
			t.Fatal("Cannot cast to rpc/util", util)
		} else if record := util.NewServiceRecord(rpc.DISCOVERY_TYPE_DB); record == nil {
			t.Error("Unexpected record == nil")
		} else {
			t.Log(record)
		}
	}
}

func Test_ServiceRecord_005(t *testing.T) {
	config := gopi.NewAppConfig("rpc/util")
	config.Debug = true
	config.Verbose = true
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Error(err)
	} else {
		defer app.Close()
		if util, ok := app.ModuleInstance("rpc/util").(rpc.Util); ok == false {
			t.Fatal("Cannot cast to rpc/util", util)
		} else if record := util.NewServiceRecord(rpc.DISCOVERY_TYPE_DB); record == nil {
			t.Error("Unexpected record == nil")
		} else if err := record.SetService("_http._udp", "printer"); err != nil {
			t.Error(err)
		} else if record.Service() != "_http._udp" {
			t.Error("Unexpected service type, ", record.Service())
		} else if record.Subtype() != "printer" {
			t.Error("Unexpected service subtype, ", record.Subtype())
		} else {
			t.Log(record)
		}
	}
}
func Test_ServiceRecord_006(t *testing.T) {
	config := gopi.NewAppConfig("rpc/util")
	config.Debug = true
	config.Verbose = true

	tests := []struct {
		service string
		subtype string
		result  string
	}{
		{"_a._tcp", "b", "_a._tcp"},
		{"_a._tcp", "", "_a._tcp"},
		{"_a._udp", "", "_a._udp"},
		{"_http._tcp", "", "_http._tcp"},
		{"_a._udp", "b", "_a._udp"},
		{"_a._tcp", "b", "_a._tcp"},
		{"_a._tcp", "printer", "_a._tcp"},
		{"_a._tcp", "subtype", "_a._tcp"},
	}

	app, _ := gopi.NewAppInstance(config)
	util := app.ModuleInstance("rpc/util").(rpc.Util)
	record := util.NewServiceRecord(rpc.DISCOVERY_TYPE_DB)
	for i, test := range tests {
		if err := record.SetService(test.service, test.subtype); err != nil {
			t.Error(err)
		} else if record.Service() != test.result {
			t.Errorf("Unexpected: service=%v expected=%v (test %v)", record.Service(), test.result, i+1)
		} else if test.subtype != "" && record.Subtype() != test.subtype {
			t.Errorf("Unexpected: subtype=%v expected=%v (test %v)", record.Subtype(), test.subtype, i+1)
		}
	}
}
