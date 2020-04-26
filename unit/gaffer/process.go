/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package gaffer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	rpc "github.com/djthorpe/gopi-rpc/v2"
	"github.com/djthorpe/gopi/v2"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Process struct {
	sync.Mutex
	sync.WaitGroup

	id   uint32 // Process ID
	root string // Service root

	service        *service           // Service description
	cmd            *exec.Cmd          // Command object
	cancel         context.CancelFunc // Cancel function
	stdout, stderr io.ReadCloser      // Stdout and Stderr
	stopping       bool               // Stopping flag
	timeout        time.Duration      // Timeout
	ts             time.Time          // Last updated
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	BUF_MAX_SIZE = 1024
)

////////////////////////////////////////////////////////////////////////////////
// NEW

// NewProcess returns a new process object which is used to control processes
func NewProcess(id uint32, service *service, root string, timeout time.Duration) (*Process, error) {
	this := new(Process)
	ctx, cancel := ctxForTimeout(timeout)
	this.service = service
	this.root = root
	this.cmd = exec.CommandContext(ctx, service.path, service.args...)
	this.cancel = cancel
	this.id = id

	// Transform user & group into uid and gid
	uid := uint32(0)
	gid := uint32(0)
	if service.user != "" {
		if uid_, err := strconv.ParseUint(service.user, 10, 32); err != nil {
			return nil, gopi.ErrBadParameter.WithPrefix("user")
		} else {
			uid = uint32(uid_)
		}
	}
	if service.group != "" {
		if gid_, err := strconv.ParseUint(service.group, 10, 32); err != nil {
			return nil, gopi.ErrBadParameter.WithPrefix("group")
		} else {
			gid = uint32(gid_)
		}
	}
	if uid == 0 {
		uid = uint32(syscall.Getuid())
	}
	if gid == 0 {
		gid = uint32(syscall.Getgid())
	}
	this.cmd.SysProcAttr = &syscall.SysProcAttr{}
	this.cmd.SysProcAttr.Credential = &syscall.Credential{
		Uid:         uid,
		Gid:         gid,
		NoSetGroups: true,
	}

	// Set user & group into names
	if u, err := user.LookupId(service.user); err == nil {
		if u.Username == "" {
			this.service.user = u.Uid
		} else {
			this.service.user = u.Username
		}
	}
	if g, err := user.LookupGroupId(service.group); err == nil {
		if g.Name == "" {
			this.service.group = g.Gid
		} else {
			this.service.group = g.Name
		}
	}

	// Set home folder based on user if not explicitly set
	if this.service.cwd == "" && uid != 0 {
		if user, err := user.LookupId(service.user); err == nil {
			if stat, err := os.Stat(user.HomeDir); err == nil && stat.IsDir() {
				this.service.cwd = user.HomeDir
			}
		}
	}

	// Set working directory
	if this.service.cwd != "" {
		if stat, err := os.Stat(this.service.cwd); err != nil {
			return nil, fmt.Errorf("wd: %w", err)
		} else if stat.IsDir() == false {
			return nil, fmt.Errorf("wd: %w", gopi.ErrBadParameter)
		} else {
			this.cmd.Dir = this.service.cwd
		}
	}

	// Set service path
	if this.root != "" {
		this.service.path = relativePath(this.service.path, this.root)
	}

	// Set deadline
	if deadline, ok := ctx.Deadline(); ok {
		this.timeout = deadline.Sub(time.Now())
	}

	// Set stdout,stderr
	if stdout, err := this.cmd.StdoutPipe(); err != nil {
		return nil, err
	} else {
		this.stdout = stdout
	}
	if stderr, err := this.cmd.StderrPipe(); err != nil {
		return nil, err
	} else {
		this.stderr = stderr
	}

	// Set timestamp
	this.ts = time.Now()

	// Success
	return this, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *Process) Start(out chan<- *Event) error {
	this.Lock()
	defer this.Unlock()

	// Start logging to channels,
	go this.processLogger(this.stdout, out, rpc.GAFFER_STATE_STDOUT)
	go this.processLogger(this.stderr, out, rpc.GAFFER_STATE_STDERR)

	// Start but don't wait
	if err := this.cmd.Start(); err != nil {
		return err
	}

	// Set timestamp
	this.ts = time.Now()

	// Wait for process to end
	go this.processWait(out)

	// Success
	return nil
}

func (this *Process) Stop() error {
	this.Lock()
	defer this.Unlock()

	// Set state to stopping
	this.stopping = true
	this.ts = time.Now()

	if this.cmd.Process != nil {
		this.cancel()
	}

	// Wait for all goroutines to have ended
	this.WaitGroup.Wait()

	// Success
	return nil
}

func (this *Process) Pid() uint32 {
	this.Lock()
	defer this.Unlock()

	if this.cmd != nil && this.cmd.Process != nil {
		return uint32(this.cmd.Process.Pid)
	} else {
		return 0
	}
}

func (this *Process) Id() uint32 {
	return this.id
}

func (this *Process) Timestamp() time.Time {
	return this.ts
}

func (this *Process) ExitCode() int64 {
	this.Lock()
	defer this.Unlock()

	if this.cmd != nil && this.cmd.ProcessState != nil {
		return int64(this.cmd.ProcessState.ExitCode())
	} else {
		return 0
	}
}

func (this *Process) Service() rpc.GafferService {
	return this.service
}

func (this *Process) State() rpc.GafferState {
	this.Lock()
	defer this.Unlock()

	if this.cmd.Process == nil && this.cmd.ProcessState == nil {
		return rpc.GAFFER_STATE_NEW
	} else if this.stopping {
		return rpc.GAFFER_STATE_STOPPING
	} else if this.cmd.ProcessState != nil {
		return rpc.GAFFER_STATE_STOPPED
	} else if this.cmd.Process != nil {
		return rpc.GAFFER_STATE_RUNNING
	} else {
		return rpc.GAFFER_STATE_NONE
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Process) String() string {
	str := "<gaffer.Process"
	str += " id=" + fmt.Sprint(this.id)
	str += " service=" + fmt.Sprint(this.service)
	if this.root != "" {
		str += " root=" + strconv.Quote(this.root)
	}
	str += " exec=" + strconv.Quote(relativePath(this.cmd.Path, this.root))
	str += " state=" + fmt.Sprint(this.State())
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// BACKGROUND TASKS

func (this *Process) processLogger(fh io.Reader, out chan<- *Event, t rpc.GafferState) {
	this.WaitGroup.Add(1)
	defer this.WaitGroup.Done()

	buf := bufio.NewReader(fh)
	bytes := make([]byte, BUF_MAX_SIZE)
FOR_LOOP:
	for {
		if n, err := buf.Read(bytes); err == io.EOF {
			break FOR_LOOP
		} else if n > 0 {
			out <- NewBufferEvent(this, bytes[:n], t)
		}
	}
}

func (this *Process) processWait(out chan<- *Event) {
	this.WaitGroup.Add(1)
	defer this.WaitGroup.Done()

	err := this.cmd.Wait()

	// Flag state change
	this.stopping = false
	this.ts = time.Now()

	// Output stop event
	out <- NewStoppedEvent(this, err)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func ctxForTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		return context.WithCancel(context.Background())
	} else {
		return context.WithTimeout(context.Background(), timeout)
	}
}

func relativePath(abs, root string) string {
	innerRelpath := func(abs, root string) string {
		if root == "" {
			return abs
		} else if rel, err := filepath.Rel(root, abs); err == nil {
			return rel
		} else {
			return abs
		}
	}
	// Ensure it starts with a "/"
	return filepath.Clean(filepath.Join("/", innerRelpath(abs, root)))
}
