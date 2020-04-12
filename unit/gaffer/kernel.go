/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package gaffer

import (
	// Frameworks
	"fmt"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	rpc "github.com/djthorpe/gopi-rpc/v2"
	gopi "github.com/djthorpe/gopi/v2"
	base "github.com/djthorpe/gopi/v2/base"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Kernel struct {
	gopi.Config
	Root string // Root is the path that executables are under
}

type kernel struct {
	base.Unit
	sync.Mutex
	sync.WaitGroup

	root             string
	process          map[uint32]*Process
	stop             chan struct{} // stop kernel signal
	runproc, endproc chan uint32   // run and end process signal
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Unit

func (Kernel) Name() string { return "gaffer/kernel" }

func (config Kernel) New(log gopi.Logger) (gopi.Unit, error) {
	this := new(kernel)
	if err := this.Unit.Init(log); err != nil {
		return nil, err
	}
	if err := this.Init(config); err != nil {
		return nil, err
	}

	// Background orchestrator
	go this.BackgroundProcess()

	return this, nil
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gaffer.Kernel

func (this *kernel) Init(config Kernel) error {
	// Seed random number
	rand.Seed(time.Now().Unix())

	// Make sure the root exists and is a folder
	if config.Root == "" {
		config.Root = "/"
	}
	if stat, err := os.Stat(config.Root); err != nil {
		return fmt.Errorf("-gaffer.root: %w", err)
	} else if stat.IsDir() == false {
		return fmt.Errorf("-gaffer.root: %w", gopi.ErrBadParameter)
	} else {
		this.root = filepath.Clean(config.Root)
	}

	// Set up processes, stop signal
	this.process = make(map[uint32]*Process)
	this.stop = make(chan struct{})
	this.runproc, this.endproc = make(chan uint32), make(chan uint32)

	// Success
	return nil
}

func (this *kernel) Close() error {
	// signal stop and wait for end
	close(this.stop)
	this.WaitGroup.Wait()

	// Close channels
	close(this.runproc)
	close(this.endproc)

	// Release resources
	this.process = nil
	this.runproc = nil
	this.endproc = nil
	this.stop = nil

	// Return success
	return nil
}

func (this *kernel) CreateProcess(service rpc.GafferService) (uint32, error) {
	this.Lock()
	defer this.Unlock()

	if path, err := this.pathForExecutable(strings.TrimSpace(service.Path)); err != nil {
		return 0, err
	} else if uid, gid, err := getUserGroup(service.User, service.Group); err != nil {
		return 0, err
	} else if id := this.newId(); id == 0 {
		return 0, gopi.ErrInternalAppError
	} else if process, err := NewProcess(id, service.Sid, path, service.Wd, service.Args, uid, gid, service.Timeout); err != nil {
		return 0, err
	} else if _, exists := this.process[id]; exists {
		return 0, gopi.ErrInternalAppError
	} else {
		this.process[id] = process
		// signal run
		this.runproc <- id
		// return success
		return id, nil
	}
}

func (this *kernel) Processes(id, sid uint32) []rpc.GafferProcess {
	processes := make([]rpc.GafferProcess, 0, len(this.process))
	for _, process := range this.process {
		if sid != 0 && sid != process.sid {
			continue
		}
		if id != 0 && id != process.id {
			continue
		}
		processes = append(processes, process)
	}
	return processes
}

////////////////////////////////////////////////////////////////////////////////
// BACKGROUND PROCESS

func (this *kernel) BackgroundProcess() {
	this.WaitGroup.Add(1)
	defer this.WaitGroup.Done()

	// ticker every 10 secs
	ticker := time.NewTimer(100 * time.Millisecond)

FOR_LOOP:
	for {
		select {
		case <-this.stop:
			ticker.Stop()
			break FOR_LOOP
		case id := <-this.runproc:
			// create a channel for events
			events := make(chan *Event)
			if process, exists := this.process[id]; exists {
				if err := process.Start(events); err != nil {
					// TODO: Emit error event
					this.Log.Error(err)
					close(events)
				} else {
					// Start routine to receive events
					go this.EventProcess(id, events)
				}
			}
		case id := <-this.endproc:
			this.Log.Debug("PROC", id, "HAS ENDED")
		case <-ticker.C:
			this.Log.Debug("TICK")
			ticker.Reset(10 * time.Second)
		}
	}
}

func (this *kernel) EventProcess(id uint32, out <-chan *Event) {
FOR_LOOP:
	for {
		select {
		case evt := <-out:
			fmt.Println("ID=", id, "EVT=", evt)
			if evt.Type == EVENT_TYPE_STOPPED {
				break FOR_LOOP
			}
		}
	}
	// Signal end of process
	this.endproc <- id
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *kernel) pathForExecutable(path string) (string, error) {
	// Set path to be under root
	if path == "" {
		return "", gopi.ErrBadParameter
	} else {
		// Join path and root
		path = filepath.Clean(filepath.Join(this.root, path))
	}

	// Ensure path is under root
	if strings.HasPrefix(path, this.root) == false {
		return "", gopi.ErrBadParameter
	}

	// Ensure exists and regular file
	if stat, err := os.Stat(path); err != nil {
		return "", err
	} else if stat.Mode().IsRegular() == false {
		return "", fmt.Errorf("%w: Not a regular file", gopi.ErrBadParameter)
	}

	return path, nil
}

func (this *kernel) newId() uint32 {
	// Try to get a unique id 25 times before failing
	// the first id's would be between 1 and 63 and the second
	// between 1 and 127 and so forth.
	mod := uint32(64)
	for i := 0; i < 25; i++ {
		rand := uint32(rand.Int31()) % mod
		if _, exists := this.process[rand]; exists == false && rand > 0 {
			return rand
		} else {
			mod <<= 1
		}
	}
	return 0
}

func getUserGroup(u, g string) (uint32, uint32, error) {
	uid, gid := uint32(0), uint32(0)

	// Find user/group from u
	if u != "" {
		if user_, err := user.Lookup(u); err == nil {
			if uid_, err := strconv.ParseUint(user_.Uid, 10, 32); err != nil {
				return 0, 0, err
			} else {
				uid = uint32(uid_)
			}
			if gid_, err := strconv.ParseUint(user_.Gid, 10, 32); err != nil {
				return 0, 0, err
			} else {
				gid = uint32(gid_)
			}
		} else if user_, err := user.LookupId(u); err == nil {
			if uid_, err := strconv.ParseUint(user_.Uid, 10, 32); err != nil {
				return 0, 0, err
			} else {
				uid = uint32(uid_)
			}
			if gid_, err := strconv.ParseUint(user_.Gid, 10, 32); err != nil {
				return 0, 0, err
			} else {
				gid = uint32(gid_)
			}
		} else {
			return 0, 0, fmt.Errorf("%w: Invalid user", gopi.ErrBadParameter)
		}
	}

	// Find group from g
	if g != "" {
		if group_, err := user.LookupGroup(g); err == nil {
			if gid_, err := strconv.ParseUint(group_.Gid, 10, 32); err != nil {
				return 0, 0, err
			} else {
				gid = uint32(gid_)
			}
		} else if group_, err := user.LookupGroupId(g); err == nil {
			if gid_, err := strconv.ParseUint(group_.Gid, 10, 32); err != nil {
				return 0, 0, err
			} else {
				gid = uint32(gid_)
			}
		} else {
			return 0, 0, fmt.Errorf("%w: Invalid group", gopi.ErrBadParameter)
		}
	}

	// Return success
	return uid, gid, nil
}
