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

	id, sid        uint32             // Process ID and Service ID
	name           string             // Service name
	cmd            *exec.Cmd          // Command object
	cancel         context.CancelFunc // Cancel function
	stdout, stderr io.ReadCloser      // Stdout and Stderr
	user, group    string             // User and group process run under
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
func NewProcess(id, sid uint32, name, path string, cwd string, args []string, uid, gid uint32, timeout time.Duration) (*Process, error) {
	this := new(Process)
	ctx, cancel := ctxForTimeout(timeout)
	this.name = name
	this.cmd = exec.CommandContext(ctx, path, args...)
	this.cancel = cancel
	this.id = id
	this.sid = sid

	// Set user and group
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

	// Set user & group
	if u, err := user.LookupId(fmt.Sprint(uid)); err == nil {
		this.user = u.Username
		if this.user == "" {
			this.user = u.Uid
		}
	}
	if g, err := user.LookupGroupId(fmt.Sprint(gid)); err == nil {
		this.group = g.Name
		if this.group == "" {
			this.group = g.Gid
		}
	}

	// Set home folder based on user if not explicitly set
	if cwd == "" && uid != 0 {
		if user, err := user.LookupId(fmt.Sprint(uid)); err == nil {
			if stat, err := os.Stat(user.HomeDir); err == nil && stat.IsDir() {
				cwd = user.HomeDir
			}
		}
	}

	// Set working directory
	if cwd != "" {
		if stat, err := os.Stat(cwd); err != nil {
			return nil, fmt.Errorf("wd: %w", err)
		} else if stat.IsDir() == false {
			return nil, fmt.Errorf("wd: %w", gopi.ErrBadParameter)
		} else {
			this.cmd.Dir = cwd
		}
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
	this.Lock()
	defer this.Unlock()

	if this.cmd == nil {
		return rpc.GafferService{}
	} else {
		return rpc.GafferService{
			Name:  this.name,
			Path:  this.cmd.Path, // TODO: relative to root
			Cwd:   this.cmd.Dir,
			Args:  this.cmd.Args,
			Sid:   this.sid,
			User:  this.user,
			Group: this.group,
		}
	}
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
	if this.name != "" {
		str += " name=" + strconv.Quote(this.name)
	}
	str += " exec=" + strconv.Quote(this.cmd.Path)
	str += " state=" + fmt.Sprint(this.State())
	str += " user=" + fmt.Sprint(this.user)
	str += " group=" + fmt.Sprint(this.group)
	if this.sid != 0 {
		str += " sid=" + fmt.Sprint(this.sid)
	}
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
