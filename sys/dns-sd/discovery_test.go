package discovery_test

import (
	"testing"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"

	// Modules
	discovery "github.com/djthorpe/gopi-rpc/sys/dns-sd"
	rpcutil "github.com/djthorpe/gopi-rpc/sys/rpcutil"
	_ "github.com/djthorpe/gopi/sys/logger"
)

func Test_Discovery_001(t *testing.T) {
	if util, err := gopi.Open(rpcutil.Util{}, nil); err != nil {
		t.Fatal(err)
	} else if driver, err := gopi.Open(discovery.Discovery{
		Flags: gopi.RPC_FLAG_INET_V4,
		Util:  util.(rpc.Util),
	}, nil); err != nil {
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
		time.Sleep(time.Second * 5)
	}
}
