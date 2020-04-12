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

	id, sid        uint32
	timeout        time.Duration
	cmd            *exec.Cmd
	cancel         context.CancelFunc
	stdout, stderr io.ReadCloser
	user, group    string
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	BUF_MAX_SIZE = 1024 * 64
)

////////////////////////////////////////////////////////////////////////////////
// NEW

// NewProcess returns a new process object which is used to control processes
func NewProcess(id, sid uint32, path string, wd string, args []string, uid, gid uint32, timeout time.Duration) (*Process, error) {
	this := new(Process)
	ctx, cancel := ctxForTimeout(timeout)
	this.cmd = exec.CommandContext(ctx, path, args...)
	this.cancel = cancel
	this.id = id
	this.sid = sid

	// Set user and group
	if uid != 0 || gid != 0 {
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
	}

	// Set user & group
	if u, err := user.LookupId(fmt.Sprint(uid)); err == nil {
		this.user = u.Username
		if this.user == "" {
			this.user = u.Uid
		}
	}
	if g, err := user.LookupGroupId(fmt.Sprint(uid)); err == nil {
		this.group = g.Name
		if this.group == "" {
			this.group = g.Gid
		}
	}

	// Set home folder based on user if not explicitly set
	if wd == "" && uid != 0 {
		if user, err := user.LookupId(fmt.Sprint(uid)); err == nil {
			if stat, err := os.Stat(user.HomeDir); err == nil && stat.IsDir() {
				wd = user.HomeDir
			}
		}
	}

	// Set working directory
	if wd != "" {
		if stat, err := os.Stat(wd); err != nil {
			return nil, fmt.Errorf("wd: %w", err)
		} else if stat.IsDir() == false {
			return nil, fmt.Errorf("wd: %w", gopi.ErrBadParameter)
		} else {
			this.cmd.Dir = wd
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

	// Success
	return this, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *Process) Start(out chan<- *Event) error {
	this.Lock()
	defer this.Unlock()

	// Start but don't wait
	if err := this.cmd.Start(); err != nil {
		return err
	}

	// Start logging to channels, and wait for process to end
	go this.processLogger(this.stdout, out, EVENT_TYPE_STDOUT)
	go this.processLogger(this.stderr, out, EVENT_TYPE_STDERR)
	go this.processWait(out)

	// Success
	return nil
}

func (this *Process) Stop() error {
	this.Lock()
	defer this.Unlock()

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
			Path:    this.cmd.Path,
			Wd:      this.cmd.Dir,
			Args:    this.cmd.Args,
			Timeout: this.timeout,
			Sid:     this.sid,
			User:    this.user,
			Group:   this.group,
		}
	}
}

func (this *Process) Status() rpc.GafferStatus {
	this.Lock()
	defer this.Unlock()

	if this.cmd.Process == nil && this.cmd.ProcessState == nil {
		return rpc.GAFFER_STATUS_STARTING
	} else if this.cmd.ProcessState != nil {
		return rpc.GAFFER_STATUS_STOPPED
	} else if this.cmd.Process != nil {
		return rpc.GAFFER_STATUS_RUNNING
	} else {
		return rpc.GAFFER_STATUS_NONE
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Process) String() string {
	str := "<gaffer.Process"
	str += " id=" + fmt.Sprint(this.id)
	str += " exec=" + strconv.Quote(this.cmd.Path)
	str += " status=" + fmt.Sprint(this.Status())

	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// BACKGROUND TASKS

func (this *Process) processLogger(fh io.Reader, out chan<- *Event, t EventType) {
	this.WaitGroup.Add(1)
	defer this.WaitGroup.Done()

	buf := bufio.NewReader(fh)
	bytes := make([]byte, BUF_MAX_SIZE)
	for {
		if n, err := buf.Read(bytes); err == io.EOF {
			break
		} else if n > 0 {
			out <- NewBufferEvent(bytes[:n], t)
		}
	}
}

func (this *Process) processWait(out chan<- *Event) {
	this.WaitGroup.Add(1)
	defer this.WaitGroup.Done()

	err := this.cmd.Wait()
	out <- NewStoppedEvent(err)
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
