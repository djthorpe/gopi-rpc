/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/
package gaffer_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	gaffer "github.com/djthorpe/gopi-rpc/sys/gaffer"
	logger "github.com/djthorpe/gopi/sys/logger"
)

func Test_Instances_001(t *testing.T) {
	config := gaffer.Gaffer{}
	log, _ := gopi.Open(logger.Config{}, nil)
	instances := new(gaffer.Instances)
	if err := instances.Init(config, log.(gopi.Logger)); err != nil {
		t.Fatalf("instances.Init: %v", err)
	} else if err := instances.Destroy(); err != nil {
		t.Fatalf("instances.Destroy: %v", err)
	} else {
		t.Logf("instances=%v", instances.String())
	}
}

func Test_Instances_002(t *testing.T) {
	log, _ := gopi.Open(logger.Config{}, nil)
	config := gaffer.Gaffer{MaxInstances: 10}
	instances := new(gaffer.Instances)
	if err := instances.Init(config, log.(gopi.Logger)); err != nil {
		t.Fatalf("instances: %v", err)
	} else {
		defer instances.Destroy()
		for i := 0; i < 10; i++ {
			if id := instances.GetUnusedIdentifier(); id == 0 {
				t.Error("Unexpected return from GetUnusedIdentifier: id == 0")
			} else {
				t.Logf("GetUnusedIdentifier: id = %v", id)
			}
		}
	}
}

func Test_Instances_003(t *testing.T) {
	log, _ := gopi.Open(logger.Config{}, nil)
	config := gaffer.Gaffer{MaxInstances: 10}
	instances := new(gaffer.Instances)
	if err := instances.Init(config, log.(gopi.Logger)); err != nil {
		t.Fatalf("instances: %v", err)
	} else {
		defer instances.Destroy()
		errs := 0
		for i := 0; i < 20; i++ {
			if id := instances.GetUnusedIdentifier(); id == 0 {
				errs++
			} else {
				t.Logf("GetUnusedIdentifier: id = %v", id)
			}
		}
		if errs != 10 {
			t.Errorf("Expected 10 errors, got %v", errs)
		}
	}
}

func Test_Instances_004(t *testing.T) {
	log, _ := gopi.Open(logger.Config{}, nil)
	config := gaffer.Gaffer{MaxInstances: 100}
	instances := new(gaffer.Instances)
	if err := instances.Init(config, log.(gopi.Logger)); err != nil {
		t.Fatalf("instances: %v", err)
	} else {
		defer instances.Destroy()
		errs := 0
		for i := 0; i < 101; i++ {
			if id := instances.GetUnusedIdentifier(); id == 0 {
				errs++
			} else {
				t.Logf("GetUnusedIdentifier: id = %v", id)
			}
		}
		if errs != 1 {
			t.Errorf("Expected 1 errors, got %v", errs)
		}
	}
}

func Test_Instances_005(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG}, nil)
	config := gaffer.Gaffer{MaxInstances: 100, DeltaCleanup: 1 * time.Second}
	instances := new(gaffer.Instances)
	if err := instances.Init(config, log.(gopi.Logger)); err != nil {
		t.Fatalf("instances: %v", err)
	} else {
		defer instances.Destroy()
		for i := 0; i < 100; i++ {
			if id := instances.GetUnusedIdentifier(); id == 0 {
				t.Errorf("GetUnusedIdentifier: id = %v", id)
			} else {
				t.Logf("GetUnusedIdentifier: id = %v", id)
			}
		}
		time.Sleep(1 * time.Second)
		for i := 0; i < 100; i++ {
			if id := instances.GetUnusedIdentifier(); id == 0 {
				t.Errorf("GetUnusedIdentifier: id = %v", id)
			} else {
				t.Logf("GetUnusedIdentifier: id = %v", id)
			}
		}
	}
}

func Test_Instances_006(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG}, nil)
	config := gaffer.Gaffer{MaxInstances: 100, DeltaCleanup: 1 * time.Second}
	instances := new(gaffer.Instances)
	if err := instances.Init(config, log.(gopi.Logger)); err != nil {
		t.Fatalf("instances: %v", err)
	} else {
		defer instances.Destroy()
		if _, err := instances.NewInstance(0, nil, nil, ""); err == nil {
			t.Error("Expecing err != nil")
		}
		if _, err := instances.NewInstance(1, nil, nil, ""); err == nil {
			t.Error("Expecing err != nil")
		}
		if _, err := instances.NewInstance(1, &gaffer.Service{InstanceCount_: 0}, nil, ""); err == nil {
			t.Error("Expecing err != nil")
		}
	}
}

func Test_Instances_007(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG}, nil)
	config := gaffer.Gaffer{MaxInstances: 100, DeltaCleanup: 1 * time.Second}
	instances := new(gaffer.Instances)
	test_folder, test_exec := "test_folder", "test_file"

	if tmp_folder, err := ioutil.TempDir("", test_folder); err != nil {
		t.Fatalf("instances: %v", err)
	} else if err := instances.Init(config, log.(gopi.Logger)); err != nil {
		t.Fatalf("instances: %v", err)
	} else if err := MakeRegularFile(tmp_folder, test_exec, 0755); err != nil {
		t.Fatalf("instances: %v", err)
	} else {
		defer instances.Destroy()
		service := &gaffer.Service{Name_: test_exec, InstanceCount_: 1, Path_: test_exec, Groups_: []string{}, Flags_: NewTuples(), Env_: NewTuples(), Mode_: rpc.GAFFER_MODE_MANUAL}
		if _, err := instances.NewInstance(1, service, nil, test_folder); err == nil {
			t.Error("Expecting err != nil")
		} else {
			t.Logf("OK, expected error=%v", err)
		}

		if id := instances.GetUnusedIdentifier(); id == 0 {
			t.Error("Expecting id != 0")
		} else if instance, err := instances.NewInstance(id, service, []*gaffer.ServiceGroup{}, tmp_folder); err != nil {
			t.Errorf("Unexpected error=%v", err)
		} else if instance == nil {
			t.Errorf("Unexpected instance == nil")
		} else if arr := instances.GetInstances(); len(arr) != 1 {
			t.Errorf("Unexpected instances.Instances() != 1")
		} else if arr[0] != instance {
			t.Errorf("Unexpected instances.Instances(0) != instance")
		} else {
			t.Log(instance)
		}
	}
}

func Test_Instances_008(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG}, nil)
	config := gaffer.Gaffer{MaxInstances: 100, DeltaCleanup: 1 * time.Second}
	instances := new(gaffer.Instances)
	if err := instances.Init(config, log.(gopi.Logger)); err != nil {
		t.Fatalf("instances: %v", err)
	} else {
		defer instances.Destroy()
		if service := gaffer.NewService("ls", "/bin/ls"); service == nil {
			t.Error("Expecting service != nil")
		} else {
			service.InstanceCount_ = 0
			if id := instances.GetUnusedIdentifier(); id == 0 {
				t.Error("Expecting id != 0")
			} else if _, err := instances.NewInstance(id, service, nil, ""); err == nil {
				t.Errorf("NewInstance: Expecting error")
			} else {
				t.Logf("OK, Got error: %v", err)
			}
		}
	}
}

func Test_Instances_009(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG}, nil)
	config := gaffer.Gaffer{MaxInstances: 100, DeltaCleanup: 1 * time.Second}
	instances := new(gaffer.Instances)
	if err := instances.Init(config, log.(gopi.Logger)); err != nil {
		t.Fatalf("instances: %v", err)
	} else {
		defer instances.Destroy()
		if service := gaffer.NewService("ls", "/bin/ls"); service == nil {
			t.Error("Expecting service != nil")
		} else if id := instances.GetUnusedIdentifier(); id == 0 {
			t.Error("Expecting id != 0")
		} else if instance, err := instances.NewInstance(id, service, []*gaffer.ServiceGroup{}, ""); err != nil {
			t.Errorf("NewInstance: %v", err)
		} else if instance == nil {
			t.Error("instance != nil")
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func MakeRegularFile(tmpfolder, tmpfile string, permissions os.FileMode) error {
	if f, err := os.OpenFile(filepath.Join(tmpfolder, tmpfile), os.O_RDWR|os.O_CREATE, permissions); err != nil {
		return err
	} else if err := f.Close(); err != nil {
		return err
	} else {
		return nil
	}
}

func NewTuples() rpc.Tuples {
	return rpc.Tuples{}
}
