package gaffer_test

import (
	"strconv"
	"testing"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"

	// Modules
	gaffer "github.com/djthorpe/gopi-rpc/sys/gaffer"
	logger "github.com/djthorpe/gopi/sys/logger"
)

func Test_Gaffer_001(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG2}, nil)
	if gaffer, err := gopi.Open(gaffer.Gaffer{}, log.(gopi.Logger)); err != nil {
		t.Fatal(err)
	} else if err := gaffer.Close(); err != nil {
		t.Fatal(err)
	}
}

func Test_Gaffer_002(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG2}, nil)
	if gaffer_, err := gopi.Open(gaffer.Gaffer{}, log.(gopi.Logger)); err != nil {
		t.Fatal(err)
	} else {
		defer gaffer_.Close()
		if gaffer, ok := gaffer_.(rpc.Gaffer); ok == false {
			t.Error("Cannot cast to rpc.Gaffer interface")
			_ = gaffer_.(rpc.Gaffer)
		} else {
			t.Log(gaffer)
		}
	}
}

func Test_Gaffer_003(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG2}, nil)
	if gaffer_, err := gopi.Open(gaffer.Gaffer{}, log.(gopi.Logger)); err != nil {
		t.Fatal(err)
	} else {
		defer gaffer_.Close()
		if gaffer, ok := gaffer_.(rpc.Gaffer); ok == false {
			t.Error("Cannot cast to rpc.Gaffer interface")
			_ = gaffer_.(rpc.Gaffer)
		} else if executables := gaffer.GetExecutables(false); executables == nil {
			t.Error("Expected executables != nil")
		} else if len(executables) == 0 {
			t.Error("Expected executables != 0")
		} else {
			t.Log(executables)
		}
	}
}

func Test_Gaffer_004(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG2}, nil)
	if gaffer_, err := gopi.Open(gaffer.Gaffer{}, log.(gopi.Logger)); err != nil {
		t.Fatal(err)
	} else {
		defer gaffer_.Close()
		if gaffer, ok := gaffer_.(rpc.Gaffer); ok == false {
			t.Error("Cannot cast to rpc.Gaffer interface")
			_ = gaffer_.(rpc.Gaffer)
		} else if executables := gaffer.GetExecutables(false); executables == nil {
			t.Error("Expected executables != nil")
		} else if executables_recursive := gaffer.GetExecutables(true); executables_recursive == nil {
			t.Error("Expected executables_recursive != nil")
		} else if len(executables) > len(executables_recursive) {
			t.Error("Expected len(executables) <= len(executables_recursive)")
		}
	}
}

func Test_Gaffer_006(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG2}, nil)
	if gaffer_, err := gopi.Open(gaffer.Gaffer{}, log.(gopi.Logger)); err != nil {
		t.Fatal(err)
	} else {
		defer gaffer_.Close()
		if gaffer, ok := gaffer_.(rpc.Gaffer); ok == false {
			t.Error("Cannot cast to rpc.Gaffer interface")
			_ = gaffer_.(rpc.Gaffer)
		} else if executables := gaffer.GetExecutables(false); executables == nil {
			t.Error("Expected executables != nil")
		} else {
			for _, path := range executables {
				if service, err := gaffer.AddServiceForPath(path); err != nil {
					t.Errorf("AddServiceForPath: Path %v: %v", strconv.Quote(path), err)
				} else if service == nil {
					t.Errorf("AddServiceForPath: Path %v: service == nil", strconv.Quote(path))
				}
			}
		}
	}
}

func Test_Gaffer_007(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG2}, nil)
	if gaffer_, err := gopi.Open(gaffer.Gaffer{}, log.(gopi.Logger)); err != nil {
		t.Fatal(err)
	} else {
		defer gaffer_.Close()
		if gaffer, ok := gaffer_.(rpc.Gaffer); ok == false {
			t.Error("Cannot cast to rpc.Gaffer interface")
			_ = gaffer_.(rpc.Gaffer)
		} else if executables := gaffer.GetExecutables(true); executables == nil {
			t.Error("Expected executables != nil")
		} else {
			for _, path := range executables {
				if service, err := gaffer.AddServiceForPath(path); err != nil {
					t.Errorf("AddServiceForPath: Path %v: %v", strconv.Quote(path), err)
				} else if service == nil {
					t.Errorf("AddServiceForPath: Path %v: service == nil", strconv.Quote(path))
				}
			}
		}
	}
}

func Test_Gaffer_008(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG2}, nil)
	if gaffer_, err := gopi.Open(gaffer.Gaffer{}, log.(gopi.Logger)); err != nil {
		t.Fatal(err)
	} else {
		defer gaffer_.Close()
		if gaffer, ok := gaffer_.(rpc.Gaffer); ok == false {
			t.Error("Cannot cast to rpc.Gaffer interface")
			_ = gaffer_.(rpc.Gaffer)
		} else if executables := gaffer.GetExecutables(true); executables == nil {
			t.Error("Expected executables != nil")
		} else if len(executables) == 0 {
			t.Error("Expected len(executables) != 0")
		} else if service, err := gaffer.AddServiceForPath(executables[0]); err != nil {
			t.Errorf("AddServiceForPath: Path %v: %v", strconv.Quote(executables[0]), err)
		} else if services := gaffer.GetServices(); len(services) != 1 {
			t.Error("Expected len(GetServices) == 1")
		} else if services[0] != service {
			t.Error("Expected services[0] == service")
		}
	}
}

func Test_Gaffer_009(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG2}, nil)
	if gaffer_, err := gopi.Open(gaffer.Gaffer{}, log.(gopi.Logger)); err != nil {
		t.Fatal(err)
	} else {
		defer gaffer_.Close()
		if gaffer, ok := gaffer_.(rpc.Gaffer); ok == false {
			t.Error("Cannot cast to rpc.Gaffer interface")
			_ = gaffer_.(rpc.Gaffer)
		} else if executables := gaffer.GetExecutables(true); executables == nil {
			t.Error("Expected executables != nil")
		} else if len(executables) == 0 {
			t.Error("Expected len(executables) != 0")
		} else if service, err := gaffer.AddServiceForPath(executables[0]); err != nil {
			t.Errorf("AddServiceForPath: Path %v: %v", strconv.Quote(executables[0]), err)
		} else if err := gaffer.RemoveServiceForName(service.Name()); err != nil {
			t.Errorf("RemoveServiceForName: %v", err)
		} else if services := gaffer.GetServices(); len(services) != 0 {
			t.Error("Expected len(GetServices) == 0")
		}
	}
}

func Test_Gaffer_Groups_010(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG2}, nil)
	if gaffer_, err := gopi.Open(gaffer.Gaffer{}, log.(gopi.Logger)); err != nil {
		t.Fatal(err)
	} else {
		defer gaffer_.Close()
		if gaffer, ok := gaffer_.(rpc.Gaffer); ok == false {
			t.Error("Cannot cast to rpc.Gaffer interface")
			_ = gaffer_.(rpc.Gaffer)
		} else if _, err := gaffer.AddGroupForName(""); err == nil {
			t.Error("Expected err != nil")
		} else if _, err := gaffer.AddGroupForName("test"); err != nil {
			t.Errorf("AddGroupForName: %v", err)
		} else if groups := gaffer.GetGroups(); len(groups) != 1 {
			t.Error("Expected GetGroups == 1")
		} else if groups_ := gaffer.GetGroupsForNames([]string{"test"}); len(groups_) != 1 {
			t.Errorf("Expected GetGroupsForNames == 1, got %v", groups_)
		} else if groups[0] != groups_[0] {
			t.Error("Expected GetGroupsForNames == GetGroups")
		}
	}
}

func Test_Gaffer_Groups_011(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG2}, nil)
	if gaffer_, err := gopi.Open(gaffer.Gaffer{}, log.(gopi.Logger)); err != nil {
		t.Fatal(err)
	} else {
		defer gaffer_.Close()
		if gaffer, ok := gaffer_.(rpc.Gaffer); ok == false {
			t.Error("Cannot cast to rpc.Gaffer interface")
			_ = gaffer_.(rpc.Gaffer)
		} else if _, err := gaffer.AddGroupForName("test"); err != nil {
			t.Errorf("AddGroupForName: %v", err)
		} else if groups := gaffer.GetGroupsForNames([]string{"test"}); len(groups) != 1 {
			t.Errorf("Expected GetGroupsForNames == 1, got %v", groups)
		} else if groups[0].Name() != "test" {
			t.Error("Expected groups[0].Name() == test")
		}
	}
}

func Test_Gaffer_Groups_012(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG2}, nil)
	if gaffer_, err := gopi.Open(gaffer.Gaffer{}, log.(gopi.Logger)); err != nil {
		t.Fatal(err)
	} else {
		defer gaffer_.Close()
		if gaffer, ok := gaffer_.(rpc.Gaffer); ok == false {
			t.Error("Cannot cast to rpc.Gaffer interface")
			_ = gaffer_.(rpc.Gaffer)
		} else if _, err := gaffer.AddGroupForName("test"); err != nil {
			t.Errorf("AddGroupForName: %v", err)
		} else if groups := gaffer.GetGroupsForNames([]string{"test"}); len(groups) != 1 {
			t.Errorf("Expected GetGroupsForNames == 1, got %v", groups)
		} else if groups[0].Name() != "test" {
			t.Error("Expected groups[0].Name() == test")
		} else if err := gaffer.RemoveGroupForName("test"); err != nil {
			t.Errorf("RemoveGroupForName: %v", err)
		} else if groups := gaffer.GetGroups(); len(groups) != 0 {
			t.Error("Expected groups == 0")
		}
	}
}

func Test_Gaffer_Groups_013(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG2}, nil)
	if gaffer_, err := gopi.Open(gaffer.Gaffer{}, log.(gopi.Logger)); err != nil {
		t.Fatal(err)
	} else {
		defer gaffer_.Close()
		if gaffer, ok := gaffer_.(rpc.Gaffer); ok == false {
			t.Error("Cannot cast to rpc.Gaffer interface")
			_ = gaffer_.(rpc.Gaffer)
		} else if _, err := gaffer.AddGroupForName("test"); err != nil {
			t.Errorf("AddGroupForName: %v", err)
		} else if _, err := gaffer.AddGroupForName("second_test"); err != nil {
			t.Errorf("AddGroupForName: %v", err)
		} else if groups := gaffer.GetGroupsForNames([]string{"second_test", "test"}); len(groups) != 2 {
			t.Errorf("Expected GetGroupsForNames == 2, got %v", groups)
		} else if groups[0].Name() != "second_test" {
			t.Error("Expected groups[0].Name() == second_test")
		} else if groups[1].Name() != "test" {
			t.Error("Expected groups[1].Name() == test")
		} else if err := gaffer.RemoveGroupForName("test"); err != nil {
			t.Errorf("RemoveGroupForName: %v", err)
		} else if groups := gaffer.GetGroups(); len(groups) != 1 {
			t.Error("Expected groups == 1")
		} else if err := gaffer.RemoveGroupForName("second_test"); err != nil {
			t.Errorf("RemoveGroupForName: %v", err)
		}
	}
}

func Test_Gaffer_Groups_014(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG2}, nil)
	if gaffer_, err := gopi.Open(gaffer.Gaffer{}, log.(gopi.Logger)); err != nil {
		t.Fatal(err)
	} else {
		defer gaffer_.Close()
		if gaffer, ok := gaffer_.(rpc.Gaffer); ok == false {
			t.Error("Cannot cast to rpc.Gaffer interface")
			_ = gaffer_.(rpc.Gaffer)
		} else if group, err := gaffer.AddGroupForName("test"); err != nil {
			t.Errorf("AddGroupForName: %v", err)
		} else if len(group.Flags()) != 0 {
			t.Errorf("Expected len(flags) == 0")
		} else if err := group.SetFlag("key", "value"); err != nil {
			t.Errorf("SetFlag: %v", err)
		} else if flags := group.Flags(); len(flags) != 1 {
			t.Errorf("Expected len(flags) == 1")
		} else if flags[0] != "key=\"value\"" {
			t.Errorf("Unexpected flags, %v", flags)
		} else if err := group.SetFlag("key", "value2"); err != nil {
			t.Errorf("SetFlag: %v", err)
		} else if flags := group.Flags(); len(flags) != 1 {
			t.Errorf("Expected len(flags) == 1")
		} else if flags[0] != "key=\"value2\"" {
			t.Errorf("Unexpected flags, %v", flags)
		} else {
			t.Log(group)
		}
	}
}

func Test_Gaffer_Groups_014(t *testing.T) {
	log, _ := gopi.Open(logger.Config{Level: logger.LOG_DEBUG2}, nil)
	if gaffer_, err := gopi.Open(gaffer.Gaffer{}, log.(gopi.Logger)); err != nil {
		t.Fatal(err)
	} else {
		defer gaffer_.Close()
		if gaffer, ok := gaffer_.(rpc.Gaffer); ok == false {
			t.Error("Cannot cast to rpc.Gaffer interface")
			_ = gaffer_.(rpc.Gaffer)
		} else if group, err := gaffer.AddGroupForName("test"); err != nil {
			t.Errorf("AddGroupForName: %v", err)
		} else if len(group.Flags()) != 0 {
			t.Errorf("Expected len(flags) == 0")
		} else if err := group.SetFlag("key", "value"); err != nil {
			t.Errorf("SetFlag: %v", err)
		} else if flags := group.Flags(); len(flags) != 1 {
			t.Errorf("Expected len(flags) == 1")
		} else if flags[0] != "key=\"value\"" {
			t.Errorf("Unexpected flags, %v", flags)
		} else if err := group.SetFlag("key", "value2"); err != nil {
			t.Errorf("SetFlag: %v", err)
		} else if flags := group.Flags(); len(flags) != 1 {
			t.Errorf("Expected len(flags) == 1")
		} else if flags[0] != "key=\"value2\"" {
			t.Errorf("Unexpected flags, %v", flags)
		} else {
			t.Log(group)
		}
	}
}
