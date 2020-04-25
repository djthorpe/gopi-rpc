/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package gaffer

import (
	"fmt"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	// Frameworks
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
	base.PubSub
	sync.Mutex
	sync.WaitGroup

	root             string
	process          map[uint32]*Process
	stop             chan struct{} // stop kernel signal
	runproc, endproc chan uint32   // run and end process signal
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	// Time to look to prune new and stopped processes
	DURATION_PRUNE = 10 * time.Second
)

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

	// Lock to release resources
	this.Lock()
	defer this.Unlock()

	// Unsubscribe channels
	this.PubSub.Close()

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

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *kernel) String() string {
	str := "<" + this.Log.Name()
	str += " root=" + strconv.Quote(this.root)
	if len(this.process) > 0 {
		str += " processes=["
		for _, process := range this.process {
			str += " " + process.String()
		}
		str += " ]"
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION GafferKernel

func (this *kernel) CreateProcess(service rpc.GafferService) (uint32, error) {
	// Get ID in critical section
	this.Lock()
	id := this.newId()
	this.Unlock()
	if id == 0 {
		return 0, gopi.ErrInternalAppError.WithPrefix("NewId")
	} else {
		return this.CreateProcessEx(id, service, 0)
	}
}

func (this *kernel) CreateProcessEx(id uint32, service rpc.GafferService, timeout time.Duration) (uint32, error) {
	this.Lock()
	defer this.Unlock()

	if _, exists := this.process[id]; exists {
		return 0, gopi.ErrDuplicateItem.WithPrefix("id")
	} else if path, err := this.pathForExecutable(strings.TrimSpace(service.Path())); err != nil {
		return 0, err
	} else if uid, gid, err := getUserGroup(service.User(), service.Group()); err != nil {
		return 0, err
	} else if service_ := NewService(service, 0); service_ == nil {
		return 0, gopi.ErrBadParameter.WithPrefix("service")
	} else {
		service_.path = path
		service_.user = fmt.Sprint(uid)
		service_.group = fmt.Sprint(gid)
		if process, err := NewProcess(id, service_, this.root, timeout); err != nil {
			return 0, err
		} else if _, exists := this.process[id]; exists {
			return 0, gopi.ErrInternalAppError.WithPrefix("CreateProcessEx")
		} else {
			this.Log.Debug("NewProcess<", process, ">")
			this.process[id] = process
		}
	}

	// Success
	return id, nil
}

// RunProcess starts a process in NEW state
func (this *kernel) RunProcess(id uint32) error {
	this.Lock()
	defer this.Unlock()

	if process, exists := this.process[id]; exists == false {
		return gopi.ErrBadParameter.WithPrefix("id")
	} else if process.State() != rpc.GAFFER_STATE_NEW {
		return gopi.ErrOutOfOrder
	} else {
		// signal the process needs to be run
		this.runproc <- id
	}

	// Success
	return nil
}

// StopProcess kills a process in RUNNING state
func (this *kernel) StopProcess(id uint32) error {
	this.Lock()
	defer this.Unlock()

	if id == 0 {
		return gopi.ErrBadParameter.WithPrefix("id")
	} else if process, exists := this.process[id]; exists == false {
		return gopi.ErrBadParameter.WithPrefix("id")
	} else if process.State() != rpc.GAFFER_STATE_RUNNING {
		return gopi.ErrOutOfOrder
	} else if err := process.Stop(); err != nil {
		return err
	}

	// Success
	return nil
}

func (this *kernel) Processes(id, sid uint32) []rpc.GafferProcess {
	processes := make([]rpc.GafferProcess, 0, len(this.process))
	for _, process := range this.process {
		if sid != 0 && sid != process.service.sid {
			continue
		}
		if id != 0 && id != process.id {
			continue
		}
		processes = append(processes, process)
	}
	return processes
}

func (this *kernel) Executables(recursive bool) []string {
	// Do not list executables if the root is not set
	if this.root == "/" || this.root == "" {
		return nil
	}

	// Walk the root path to find executables
	executables := make([]string, 0)
	if err := filepath.Walk(this.root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && recursive == false {
			return filepath.SkipDir
		}
		if info.Mode().IsRegular() && isExecutableFileAtPath(path) == nil {
			// Trim prefix
			sep := string(filepath.Separator)
			path := strings.TrimPrefix(strings.TrimPrefix(path, this.root), sep)
			// Append
			executables = append(executables, path)
		}
		return nil
	}); err != nil {
		this.Log.Error(err)
		return nil
	}

	return executables
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
					go func() {
						this.EventProcess(id, events)
					}()
				}
			}
		case id := <-this.endproc:
			this.Log.Debug("PROC", id, "HAS ENDED")
		case <-ticker.C:
			// Prune old processes which are stale after 20 seconds
			if modified := this.ProcessPrune(DURATION_PRUNE * 2); modified {
				ticker.Reset(DURATION_PRUNE / 2)
			} else {
				ticker.Reset(DURATION_PRUNE)
			}
		}
	}
}

func (this *kernel) EventProcess(id uint32, out <-chan *Event) {
FOR_LOOP:
	for {
		select {
		case evt := <-out:
			this.Emit(evt)
			if evt.state == rpc.GAFFER_STATE_STOPPED {
				break FOR_LOOP
			}
		}
	}
	// Signal end of process
	this.endproc <- id
}

func (this *kernel) ProcessPrune(delta time.Duration) bool {
	this.Lock()
	defer this.Unlock()

	for k, process := range this.process {
		switch process.State() {
		case rpc.GAFFER_STATE_NEW, rpc.GAFFER_STATE_STOPPED:
			if process.Timestamp().IsZero() == false && time.Now().Sub(process.Timestamp()) > delta {
				// Prune processes which are stale
				this.Log.Debug("Prune:", process)
				delete(this.process, k)
				return true
			}
		}
	}

	// Nothing updated
	return false
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
		// Get user ID
		user_, err := getUser(u)
		if err != nil {
			return 0, 0, err
		} else if uid_, err := strconv.ParseUint(user_.Uid, 10, 32); err != nil {
			return 0, 0, err
		} else {
			uid = uint32(uid_)
		}
		// Get group ID
		if gid_, err := strconv.ParseUint(user_.Gid, 10, 32); err != nil {
			return 0, 0, err
		} else {
			gid = uint32(gid_)
		}
	}

	// Find group from g
	if g != "" {
		group_, err := getGroup(g)
		if err != nil {
			return 0, 0, err
		} else if gid_, err := strconv.ParseUint(group_.Gid, 10, 32); err != nil {
			return 0, 0, err
		} else {
			gid = uint32(gid_)
		}
	}

	// Ensure user and group match up
	if user_, err := getUser(fmt.Sprint(uid)); err != nil {
		return 0, 0, err
	} else if gids_, err := user_.GroupIds(); err != nil {
		return 0, 0, err
	} else if containsString(gids_, fmt.Sprint(gid)) == false {
		return 0, 0, fmt.Errorf("User %v is not a member of group with gid %v", user_.Username, gid)
	}

	// Return success
	return uid, gid, nil
}

func getUser(u string) (*user.User, error) {
	// Find user/group from u
	if user_, err := user.Lookup(u); err == nil {
		return user_, nil
	} else if user_, err := user.LookupId(u); err == nil {
		return user_, nil
	} else {
		return nil, gopi.ErrNotFound.WithPrefix("user")
	}
}

func getGroup(g string) (*user.Group, error) {
	// Find group from g
	if group_, err := user.LookupGroup(g); err == nil {
		return group_, nil
	} else if group_, err := user.LookupGroupId(g); err == nil {
		return group_, nil
	} else {
		return nil, gopi.ErrNotFound.WithPrefix("group")
	}
}

func containsString(arr []string, value string) bool {
	for _, member := range arr {
		if member == value {
			return true
		}
	}
	return false
}
