/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"context"
	"fmt"
	"os"

	// Frameworks
	app "github.com/djthorpe/gopi-rpc/v2/app"
	gopi "github.com/djthorpe/gopi/v2"

	// Units
	_ "github.com/djthorpe/gopi-rpc/v2/grpc/gaffer"
	_ "github.com/djthorpe/gopi-rpc/v2/unit/gaffer"
	_ "github.com/djthorpe/gopi-rpc/v2/unit/grpc"
	_ "github.com/djthorpe/gopi/v2/unit/bus"
	_ "github.com/djthorpe/gopi/v2/unit/logger"
)

/*
////////////////////////////////////////////////////////////////////////////////
// LIST PROCESSES

func StatusToString(status rpc.GafferStatus) string {
	return strings.ToLower(strings.TrimPrefix(fmt.Sprint(status), "GAFFER_STATUS_"))
}

func SidToString(sid uint32) string {
	if sid == 0 {
		return ""
	} else {
		return fmt.Sprint(sid)
	}
}

func UserGroupToString(user, group string) string {
	if user == group || group == "" {
		return user
	} else {
		return user + ":" + group
	}
}

func Processes(app gopi.App, stub rpc.GafferKernelStub) error {
	if processes, err := stub.Processes(context.Background(), 0, 0); err != nil {
		return err
	} else if len(processes) == 0 {
		return fmt.Errorf("No processes returned")
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"id", "sid", "status", "service", "user"})
		for _, process := range processes {
			table.Append([]string{
				fmt.Sprint(process.Id()),
				SidToString(process.Service().Sid),
				StatusToString(process.Status()),
				process.Service().Path,
				UserGroupToString(process.Service().User, process.Service().Group),
			})
		}
		table.Render()
	}

	// Success
	return nil
}

func Run(app gopi.App, stub rpc.GafferKernelStub, path string, args []string) error {
	if id, err := stub.CreateProcess(context.Background(), rpc.GafferService{
		Path:    path,
		Args:    args,
		Wd:      app.Flags().GetString("wd", gopi.FLAG_NS_DEFAULT),
		User:    app.Flags().GetString("user", gopi.FLAG_NS_DEFAULT),
		Group:   app.Flags().GetString("group", gopi.FLAG_NS_DEFAULT),
		Timeout: app.Flags().GetDuration("timeout", gopi.FLAG_NS_DEFAULT),
		Sid:     uint32(app.Flags().GetUint("sid", gopi.FLAG_NS_DEFAULT)),
	}); err != nil {
		return err
	} else if err := stub.RunProcess(context.Background(), id); err != nil {
		return err
	} else {
		fmt.Println("id=", id)
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// MAIN

func Main(app gopi.App, args []string) error {
	clientpool := app.UnitInstance("clientpool").(gopi.RPCClientPool)

	if fifo := app.Flags().GetString("kernel.fifo", gopi.FLAG_NS_DEFAULT); fifo == "" {
		return fmt.Errorf("%w: Missing -kernel.fifo flag", gopi.ErrBadParameter)
	} else if conn, err := clientpool.ConnectFifo(fifo); err != nil {
		return err
	} else if stub, ok := clientpool.CreateStub("gaffer.Kernel", conn).(rpc.GafferKernelStub); ok == false {
		return gopi.ErrInternalAppError
	} else if err := stub.Ping(context.Background()); err != nil {
		return err
	} else {
		// Run commands
		if len(args) == 0 {
			return Processes(app, stub)
		} else {
			return Run(app, stub, args[0], args[1:])
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// BOOTSTRAP

func main() {
	if app, err := app.NewCommandLineTool(Main, nil, "clientpool"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		app.Flags().FlagString("kernel.fifo", "", "Gaffer Kernel Fifo")
		app.Flags().FlagString("wd", "", "Working directory")
		app.Flags().FlagString("user", "", "User")
		app.Flags().FlagString("group", "", "Group")
		app.Flags().FlagDuration("timeout", 0, "Process timeout")
		app.Flags().FlagUint("sid", 0, "Service ID")

		// Run and exit
		os.Exit(app.Run())
	}
}
*/

////////////////////////////////////////////////////////////////////////////////
// MAIN

func Main(app gopi.App, args []string) error {
	// Wait until CTRL+C pressed
	fmt.Println("Press CTRL+C to exit")
	app.WaitForSignal(context.Background(), os.Interrupt)

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// BOOTSTRAP

func main() {
	if app, err := app.NewServer(Main, "rpc/gaffer/service"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		// Run and exit
		os.Exit(app.Run())
	}
}
