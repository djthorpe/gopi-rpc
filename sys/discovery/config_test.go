package discovery_test

import (
	"strings"
	"testing"

	// Frameworks
	"github.com/djthorpe/gopi-rpc/sys/discovery"
)

func Test_Config_001(t *testing.T) {
	config := new(discovery.Config)
	if err := config.Init(); err != nil {
		t.Error(err)
	} else {
		defer config.Deinit()
		t.Log(config)
	}

}

func Test_Config_002(t *testing.T) {
	config := new(discovery.Config)
	if err := config.Init(); err != nil {
		t.Error(err)
	} else {
		defer config.Deinit()
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
		defer config.Deinit()
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
