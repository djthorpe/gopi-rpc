package discovery_test

import (
	"testing"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	discovery "github.com/djthorpe/gopi-rpc/sys/dns-sd"
	rpcutil "github.com/djthorpe/gopi-rpc/sys/rpcutil"
)

func Test_Config_001(t *testing.T) {
	config := new(discovery.Config)
	if util, err := gopi.Open(rpcutil.Util{}, nil); err != nil {
		t.Fatal(err)
	} else if err := config.Init(discovery.Discovery{Util: util.(rpc.Util)}, nil, nil); err != nil {
		t.Error(err)
	} else {
		defer config.Destroy()
		t.Log(config)
	}

}
