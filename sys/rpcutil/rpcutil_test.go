package rpcutil_test

import (
	"strconv"
	"strings"
	"testing"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"

	// Modules
	util "github.com/djthorpe/gopi-rpc/sys/rpcutil"
	_ "github.com/djthorpe/gopi/sys/logger"
)

func Test_Util_001(t *testing.T) {
	if driver, err := gopi.Open(util.Util{}, nil); err != nil {
		t.Error(err)
	} else {
		defer driver.Close()
		t.Log(driver)
	}
}

func Test_Util_002(t *testing.T) {
	if driver, err := gopi.Open(util.Util{}, nil); err != nil {
		t.Error(err)
	} else {
		defer driver.Close()
		if rpcutil, ok := driver.(rpc.Util); ok == false {
			t.Fatal("Cannot cast into rpc.Util object")
		} else {
			buf := new(strings.Builder)
			if err := rpcutil.Writer(buf, []rpc.ServiceRecord{}, true); err != nil {
				t.Fatal(err)
			} else {
				t.Log(buf)
			}
		}
	}
}

func Test_Util_003(t *testing.T) {
	if driver, err := gopi.Open(util.Util{}, nil); err != nil {
		t.Error(err)
	} else {
		defer driver.Close()
		if rpcutil, ok := driver.(rpc.Util); ok == false {
			t.Fatal("Cannot cast into rpc.Util object")
		} else {
			buf := new(strings.Builder)
			if err := rpcutil.Writer(buf, []rpc.ServiceRecord{}, true); err != nil {
				t.Fatal(err)
			} else if r := strings.NewReader(buf.String()); r == nil {
				t.Fatal("strings.NewReader failed")
			} else if arr, err := rpcutil.Reader(r); err != nil {
				t.Fatal(err)
			} else {
				t.Log(buf, "=>", arr)
			}
		}
	}
}

func Test_Util_004(t *testing.T) {
	if driver, err := gopi.Open(util.Util{}, nil); err != nil {
		t.Error(err)
	} else {
		defer driver.Close()
		if rpcutil, ok := driver.(rpc.Util); ok == false {
			t.Fatal("Cannot cast into rpc.Util object")
		} else if record := rpcutil.NewServiceRecord(rpc.DISCOVERY_TYPE_DB); record == nil {
			t.Fatal("NewServiceRecord failed")
		} else {
			buf := new(strings.Builder)
			if err := rpcutil.Writer(buf, []rpc.ServiceRecord{record}, true); err != nil {
				t.Fatal(err)
			} else if r := strings.NewReader(buf.String()); r == nil {
				t.Fatal("strings.NewReader failed")
			} else if arr, err := rpcutil.Reader(r); err != nil {
				t.Fatal(err)
			} else if len(arr) != 1 {
				t.Fatal("Expected to read a single record")
			} else if arr[0].Key() != record.Key() {
				t.Fatalf("Keys don't match, expected %v got %v", strconv.Quote(record.Key()), strconv.Quote(arr[0].Key()))
			} else {
				t.Log(buf, "=>", arr)
			}
		}
	}
}
