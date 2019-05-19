/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/
package gaffer_test

import (
	"os"
	"testing"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	gaffer "github.com/djthorpe/gopi-rpc/sys/gaffer"
	logger "github.com/djthorpe/gopi/sys/logger"
)

func Test_Process_001(t *testing.T) {
	if instances, err := Process_NewInstances(); err != nil {
		t.Fatal(err)
	} else {
		defer instances.Destroy()
		srv := &gaffer.Service{"ls", "ls", []string{}, rpc.Tuples{}, rpc.Tuples{}, rpc.GAFFER_MODE_MANUAL, 1, 0, 0}
		if id := instances.GetUnusedIdentifier(); id == 0 {
			t.Fatal("GetUnusedIdentifier returns 0")
		} else if instance, err := instances.NewInstance(id, srv, []*gaffer.ServiceGroup{}, "/bin"); err != nil {
			t.Fatal(err)
		} else {
			t.Log(instance)
		}
	}
}

func Test_Process_002(t *testing.T) {
	if instances, err := Process_NewInstances(); err != nil {
		t.Fatal(err)
	} else {
		defer instances.Destroy()
		srv := &gaffer.Service{"ls", "ls", []string{}, rpc.Tuples{}, rpc.Tuples{}, rpc.GAFFER_MODE_MANUAL, 1, 0, 0}
		if id := instances.GetUnusedIdentifier(); id == 0 {
			t.Fatal("GetUnusedIdentifier returns 0")
		} else if instance, err := instances.NewInstance(id, srv, []*gaffer.ServiceGroup{}, "/bin"); err != nil {
			t.Fatal(err)
		} else {
			t.Log(instance)
		}
	}
}

func Test_Process_003(t *testing.T) {
	if instances, err := Process_NewInstances(); err != nil {
		t.Fatal(err)
	} else {
		defer instances.Destroy()
		flags := rpc.Tuples{}
		env := rpc.Tuples{}
		env.SetStringForKey("A", "return_a")
		env.SetStringForKey("B", "return_b")
		env.SetStringForKey("C", "${A} $B $$")
		flags.SetStringForKey("test", "${A} $B $C")
		expected := "return_a return_b return_a return_b $"
		srv := &gaffer.Service{"ls", "ls", []string{}, flags, env, rpc.GAFFER_MODE_MANUAL, 1, 0, 0}
		if id := instances.GetUnusedIdentifier(); id == 0 {
			t.Fatal("GetUnusedIdentifier returns 0")
		} else if instance, err := instances.NewInstance(id, srv, []*gaffer.ServiceGroup{}, "/bin"); err != nil {
			t.Fatal(err)
		} else if test_return := instance.Flags_.StringForKey("test"); test_return != expected {
			t.Fatal("Unexpected flags value", instance.Flags_)
		}
	}
}

func Test_Process_004(t *testing.T) {
	if instances, err := Process_NewInstances(); err != nil {
		t.Fatal(err)
	} else {
		defer instances.Destroy()
		flags := rpc.Tuples{}
		env := rpc.Tuples{}
		flags.SetStringForKey("test", "${A}")
		env.SetStringForKey("A", "${B}")
		env.SetStringForKey("B", "${C}")
		env.SetStringForKey("C", "${A}")
		expected := "${C}"
		srv := &gaffer.Service{"ls", "ls", []string{}, flags, env, rpc.GAFFER_MODE_MANUAL, 1, 0, 0}
		if id := instances.GetUnusedIdentifier(); id == 0 {
			t.Fatal("GetUnusedIdentifier returns 0")
		} else if instance, err := instances.NewInstance(id, srv, []*gaffer.ServiceGroup{}, "/bin"); err != nil {
			t.Fatal(err)
		} else if test_return := instance.Flags_.StringForKey("test"); test_return != expected {
			t.Fatal("Unexpected flags value", instance.Flags_)
		}
	}
}

func Test_Process_005(t *testing.T) {
	if instances, err := Process_NewInstances(); err != nil {
		t.Fatal(err)
	} else {
		defer instances.Destroy()
		flags := rpc.Tuples{}
		env := rpc.Tuples{}
		flags.SetStringForKey("rpc.port", "${rpc.port}")
		srv := &gaffer.Service{"ls", "ls", []string{}, flags, env, rpc.GAFFER_MODE_MANUAL, 1, 0, 0}
		if id := instances.GetUnusedIdentifier(); id == 0 {
			t.Fatal("GetUnusedIdentifier returns 0")
		} else if instance, err := instances.NewInstance(id, srv, []*gaffer.ServiceGroup{}, "/bin"); err != nil {
			t.Fatal(err)
		} else if test_return := instance.Flags_.StringForKey("rpc.port"); test_return == "" {
			t.Fatal("Unexpected flags value", instance.Flags_)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func Process_NewInstances() (*gaffer.Instances, error) {
	config := gaffer.Gaffer{
		AppFlags: gopi.NewFlags("process_test"),
	}
	config.AppFlags.Parse(os.Args)
	instances := new(gaffer.Instances)
	if log, err := gopi.Open(logger.Config{}, nil); err != nil {
		return nil, err
	} else if err := instances.Init(config, log.(gopi.Logger)); err != nil {
		return nil, err
	} else {
		return instances, nil
	}
}
