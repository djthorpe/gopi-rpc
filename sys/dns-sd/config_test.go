package discovery_test

import (
	"strings"
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

func Test_Config_002(t *testing.T) {
	config := new(discovery.Config)
	if util, err := gopi.Open(rpcutil.Util{}, nil); err != nil {
		t.Fatal(err)
	} else if err := config.Init(discovery.Discovery{Util: util.(rpc.Util)}, nil, nil); err != nil {
		t.Error(err)
	} else {
		defer config.Destroy()
		var b strings.Builder
		if err := config.Writer(&b, true); err != nil {
			t.Fatal(err)
		}
		t.Log(b.String())
	}

}

func Test_Config_003(t *testing.T) {
	config := new(discovery.Config)
	if util, err := gopi.Open(rpcutil.Util{}, nil); err != nil {
		t.Fatal(err)
	} else if err := config.Init(discovery.Discovery{Util: util.(rpc.Util)}, nil, nil); err != nil {
		t.Error(err)
	} else {
		defer config.Destroy()
		b := new(strings.Builder)
		if err := config.Writer(b, true); err != nil {
			t.Fatal(err)
		}
		r := strings.NewReader(b.String())
		if _, err := config.Reader(r); err != nil {
			t.Fatal(err)
		}
	}
}
