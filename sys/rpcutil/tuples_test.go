package rpcutil_test

import (
	"strings"
	"testing"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"

	// Modules
	util "github.com/djthorpe/gopi-rpc/sys/rpcutil"
	_ "github.com/djthorpe/gopi/sys/logger"
)

func Test_Tuples_001(t *testing.T) {
	if driver, err := gopi.Open(util.Util{}, nil); err != nil {
		t.Error(err)
	} else if util, ok := driver.(rpc.Util); ok == false {
		t.Error("unable to cast tp rpc.Util")
		_ = driver.(rpc.Util)
	} else if err := util.Close(); err != nil {
		t.Error(err)
	}
}

func Test_Tuples_002(t *testing.T) {
	if driver, err := gopi.Open(util.Util{}, nil); err != nil {
		t.Error(err)
	} else if util, ok := driver.(rpc.Util); ok == false {
		t.Error("unable to cast tp rpc.Util")
		_ = driver.(rpc.Util)
	} else {
		defer util.Close()

		tuples := util.NewTuples()
		if tuples == nil {
			t.FailNow()
		}
		if strings := tuples.Strings(); strings == nil {
			t.FailNow()
		} else if len(strings) != 0 {
			t.FailNow()
		}
	}
}

func Test_Tuples_003(t *testing.T) {
	if driver, err := gopi.Open(util.Util{}, nil); err != nil {
		t.Error(err)
	} else if util, ok := driver.(rpc.Util); ok == false {
		t.Error("unable to cast tp rpc.Util")
		_ = driver.(rpc.Util)
	} else {
		defer util.Close()

		tuples := util.NewTuples()
		tuples.AddString("a", "b")
		tuples.AddString("c", "d")
		tuples.AddString("e", "f")
		if strings := tuples.Strings(); strings == nil {
			t.FailNow()
		} else if len(strings) != 3 {
			t.FailNow()
		} else if strings[0] != "a=b" {
			t.FailNow()
		} else if strings[1] != "c=d" {
			t.FailNow()
		} else if strings[2] != "e=f" {
			t.FailNow()
		}
	}
}

func Test_Tuples_004(t *testing.T) {
	if driver, err := gopi.Open(util.Util{}, nil); err != nil {
		t.Error(err)
	} else if util, ok := driver.(rpc.Util); ok == false {
		t.Error("unable to cast tp rpc.Util")
		_ = driver.(rpc.Util)
	} else {
		defer util.Close()

		tuples := util.NewTuples()
		tuples.AddString("a", "b")
		tuples.AddString("c", "d")
		tuples.AddString("a", "f")
		if strings := tuples.Strings(); strings == nil {
			t.FailNow()
		} else if len(strings) != 2 {
			t.FailNow()
		} else if strings[0] != "c=d" {
			t.FailNow()
		} else if strings[1] != "a=f" {
			t.FailNow()
		}
	}
}

func Test_Tuples_005(t *testing.T) {
	if driver, err := gopi.Open(util.Util{}, nil); err != nil {
		t.Error(err)
	} else if util, ok := driver.(rpc.Util); ok == false {
		t.Error("unable to cast tp rpc.Util")
		_ = driver.(rpc.Util)
	} else {
		defer util.Close()

		tuples := util.NewTuples()
		tuples.AddString("test1", "100")
		tuples.AddString("test2", "200")
		tuples.AddString("test3", "-100")
		if strs := tuples.Strings(); strs == nil {
			t.FailNow()
		} else if len(strs) != 3 {
			t.FailNow()
		} else if str := strings.Join(strs, " "); str != "test1=100 test2=200 test3=-100" {
			t.FailNow()
		}
	}
}

func Test_Tuples_006(t *testing.T) {
	if driver, err := gopi.Open(util.Util{}, nil); err != nil {
		t.Error(err)
	} else if util, ok := driver.(rpc.Util); ok == false {
		t.Error("unable to cast tp rpc.Util")
		_ = driver.(rpc.Util)
	} else {
		defer util.Close()

		tuples := util.NewTuples()
		tuples.AddString("test1", "")
		tuples.AddString("test2", "test quotes")
		if strs := tuples.Strings(); strs == nil {
			t.FailNow()
		} else if len(strs) != 2 {
			t.FailNow()
		} else if str := strings.Join(strs, " "); str != "test1=\"\" test2=\"test quotes\"" {
			t.FailNow()
		}
	}
}
