package rpc_test

import (
	"testing"

	// Frameworks

	rpc "github.com/djthorpe/gopi-rpc"

	// Modules

	_ "github.com/djthorpe/gopi/sys/logger"
)

func Test_Tuples_001(t *testing.T) {
	var tuples rpc.Tuples
	if tuples.Len() != 0 {
		t.FailNow()
	}
	if keys := tuples.Keys(); len(keys) != 0 {
		t.FailNow()
	}
	tuples.RemoveAll()
	if tuples.Len() != 0 {
		t.FailNow()
	}
	if keys := tuples.Keys(); len(keys) != 0 {
		t.FailNow()
	}
}

func Test_Tuples_002(t *testing.T) {
	var tuples rpc.Tuples
	tuples.SetStringForKey("k", "v")
	if tuples.Len() != 1 {
		t.Fatal("Len != 1")
	}
	if keys := tuples.Keys(); len(keys) != 1 {
		t.FailNow()
	} else if keys[0] != "k" {
		t.FailNow()
	} else if tuples.StringForKey("k") != "v" {
		t.FailNow()
	}
	tuples.RemoveAll()
	if tuples.Len() != 0 {
		t.FailNow()
	}
	if keys := tuples.Keys(); len(keys) != 0 {
		t.FailNow()
	}
}

func Test_Tuples_003(t *testing.T) {
	var tuples rpc.Tuples
	tuples.SetStringForKey("k1", "v1")
	tuples.SetStringForKey("k2", "v2")
	if tuples.Len() != 2 {
		t.Fatal("Len != 2")
	}
	if keys := tuples.Keys(); len(keys) != 2 {
		t.FailNow()
	} else if keys[0] != "k1" {
		t.FailNow()
	} else if tuples.StringForKey("k1") != "v1" {
		t.FailNow()
	} else if keys[1] != "k2" {
		t.FailNow()
	} else if tuples.StringForKey("k2") != "v2" {
		t.FailNow()
	}
	tuples.RemoveAll()
	if tuples.Len() != 0 {
		t.FailNow()
	}
	if keys := tuples.Keys(); len(keys) != 0 {
		t.FailNow()
	}
}

func Test_Tuples_004(t *testing.T) {
	var tuples rpc.Tuples
	tuples.SetStringForKey("k1", "v1")
	tuples.SetStringForKey("k2", "v2")
	tuples_copy := tuples.Copy()
	if tuples_copy.Len() != 2 {
		t.Fatal("Len != 2")
	}
	if keys := tuples_copy.Keys(); len(keys) != 2 {
		t.FailNow()
	} else if keys[0] != "k1" {
		t.FailNow()
	} else if tuples_copy.StringForKey("k1") != "v1" {
		t.FailNow()
	} else if keys[1] != "k2" {
		t.FailNow()
	} else if tuples_copy.StringForKey("k2") != "v2" {
		t.FailNow()
	}
	tuples_copy.RemoveAll()
	if tuples_copy.Len() != 0 {
		t.FailNow()
	}
	if keys := tuples_copy.Keys(); len(keys) != 0 {
		t.FailNow()
	}
}
