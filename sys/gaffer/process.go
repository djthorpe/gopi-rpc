/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package gaffer

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Process struct {
	sync.Mutex

	cmd            *exec.Cmd
	cancel         context.CancelFunc
	stdout, stderr io.ReadCloser
	start, stop    time.Time
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

var (
	ErrSuccess = errors.New("No Error")
)

////////////////////////////////////////////////////////////////////////////////
// NEW

// Return a new process object which is used to control processes
func NewProcess(instance *ServiceInstance) (*Process, error) {
	this := new(Process)
	ctx, cancel := ctxForTimeout(instance.RunTime())
	this.cmd = exec.CommandContext(ctx, instance.Path(), instance.Flags()...)
	this.cancel = cancel

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

	// Set environment
	this.cmd.Env = instance.Env()

	// Success
	return this, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *Process) Start(stdout, stderr chan<- []byte, stop chan<- error) error {
	this.Lock()
	defer this.Unlock()

	// Start but don't wait
	this.start = time.Now()
	this.stop = time.Time{}
	if err := this.cmd.Start(); err != nil {
		return err
	}

	// Call wait in the background, which then returns the error
	go func() {
		if err := this.cmd.Wait(); err != nil {
			stop <- err
		} else {
			stop <- ErrSuccess
		}
		close(stop)
	}()

	// Start logging to channels
	go ProcessLogger(this.stdout, stdout)
	go ProcessLogger(this.stderr, stderr)

	// Success
	return nil
}

func (this *Process) Stop() error {
	this.Lock()
	defer this.Unlock()

	if this.cmd.Process != nil {
		this.stop = time.Now()
		this.cancel()
	}

	// TODO: WAIT UNTIL PROCESS IS STOPPED OR TIMEOUT

	// Success
	return nil
}

func (this *Process) IsRunning() bool {
	if this.cmd == nil || this.cmd.ProcessState == nil {
		return false
	} else {
		return this.cmd.ProcessState.Exited() == false
	}
}

func (this *Process) Id() uint32 {
	if this.cmd != nil && this.cmd.Process != nil {
		return uint32(this.cmd.Process.Pid)
	} else {
		return 0
	}
}

func (this *Process) ExitCode() int64 {
	if this.cmd != nil && this.cmd.ProcessState != nil {
		return int64(this.cmd.ProcessState.ExitCode())
	} else {
		return 0
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Process) String() string {
	return fmt.Sprintf("<gaffer.Process>{ %v }", this.cmd.ProcessState)
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

////////////////////////////////////////////////////////////////////////////////
// PROCESS LOG FILES

func ProcessLogger(fh io.Reader, c chan<- []byte) error {
	buf := bufio.NewReader(fh)
	for {
		if line, err := buf.ReadBytes('\n'); err == io.EOF {
			break
		} else {
			c <- line
		}
	}
	// Close channel to quit parent goroutine
	close(c)
	// Return after channel closed
	return nil
}
