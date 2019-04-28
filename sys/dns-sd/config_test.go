package discovery_test

import (
	"strings"
	"testing"

	// Frameworks
	discovery "github.com/djthorpe/gopi-rpc/sys/dns-sd"
)

func Test_Config_001(t *testing.T) {
	config := new(discovery.Config)
	if err := config.Init(); err != nil {
		t.Error(err)
	} else {
		defer config.Destroy()
		t.Log(config)
	}

}

func Test_Config_002(t *testing.T) {
	config := new(discovery.Config)
	if err := config.Init(); err != nil {
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
	if err := config.Init(); err != nil {
		t.Error(err)
	} else {
		defer config.Destroy()
		b := new(strings.Builder)
		if err := config.Writer(b, true); err != nil {
			t.Fatal(err)
		}
		r := strings.NewReader(b.String())
		if err := config.Reader(r); err != nil {
			t.Fatal(err)
		}
	}
}
