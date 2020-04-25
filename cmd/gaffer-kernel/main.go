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
	"os/user"
	"strconv"
	"strings"

	// Frameworks

	app "github.com/djthorpe/gopi-rpc/v2/app"
	gaffer "github.com/djthorpe/gopi-rpc/v2/unit/gaffer"
	gopi "github.com/djthorpe/gopi/v2"

	// Units
	_ "github.com/djthorpe/gopi-rpc/v2/grpc/gaffer"
	_ "github.com/djthorpe/gopi-rpc/v2/unit/gaffer"
	_ "github.com/djthorpe/gopi-rpc/v2/unit/grpc"
	_ "github.com/djthorpe/gopi/v2/unit/bus"
	_ "github.com/djthorpe/gopi/v2/unit/logger"
)

////////////////////////////////////////////////////////////////////////////////
// BOOTSTRAP SERVICE

func StartService(app gopi.App) error {
	args := make([]string, 0)
	user := app.Flags().GetString("gaffer.user", gopi.FLAG_NS_DEFAULT)
	group := app.Flags().GetString("gaffer.group", gopi.FLAG_NS_DEFAULT)

	if kernel := app.UnitInstance("gaffer/kernel").(gaffer.GafferKernel); kernel == nil {
		return gopi.ErrInternalAppError.WithPrefix("StartService")
	} else if service := app.Flags().GetString("gaffer.service", gopi.FLAG_NS_DEFAULT); service == "" {
		// No service to start
		return nil
	} else {
		if port := app.Flags().GetUint("gaffer.port", gopi.FLAG_NS_DEFAULT); port != 0 {
			args = append(args, "-rpc.port", fmt.Sprint(port))
		}
		if fifo := app.Flags().GetString("rpc.sock", gopi.FLAG_NS_DEFAULT); fifo != "" {
			args = append(args, "-kernel.sock", fifo)
		}
		if state := app.Flags().GetString("gaffer.state", gopi.FLAG_NS_DEFAULT); state != "" {
			// Change ownership of folder
			if err := SetStateOwnership(state, user, group); err != nil {
				return err
			} else {
				args = append(args, "-gaffer.state", state)
			}
		}
		if app.Log().IsDebug() {
			args = append(args, "-debug")
		}
		if id, err := kernel.CreateProcessEx(0, gaffer.NewServiceEx(service, user, group, args), 0); err != nil {
			return err
		} else if err := kernel.RunProcess(id); err != nil {
			return err
		} else {
			app.Log().Info("Running", service, strings.Join(args, " "))
		}
	}

	// Return success
	return nil
}

func SetStateOwnership(folder, user, group string) error {

	// Check to make sure folder exists
	if stat, err := os.Stat(folder); err != nil {
		return gopi.ErrNotFound.WithPrefix("-gaffer.state")
	} else if stat.IsDir() == false {
		return gopi.ErrBadParameter.WithPrefix("-gaffer.state")
	}

	// Set uid, gid and mode
	uid := -1
	gid := -1
	mode := os.FileMode(0)
	if user != "" {
		if user_, err := LookupUser(user); err != nil {
			return gopi.ErrBadParameter.WithPrefix("-gaffer.user")
		} else if uid_, err := strconv.ParseUint(user_.Uid, 10, 32); err != nil {
			return gopi.ErrBadParameter.WithPrefix("-gaffer.user")
		} else if gid_, err := strconv.ParseUint(user_.Gid, 10, 32); err != nil {
			return gopi.ErrBadParameter.WithPrefix("-gaffer.user")
		} else {
			uid = int(uid_)
			gid = int(gid_)
			mode = os.FileMode(0755)
		}
	}
	if group != "" {
		if group_, err := LookupGroup(group); err != nil {
			return gopi.ErrBadParameter.WithPrefix("-gaffer.group")
		} else if gid_, err := strconv.ParseUint(group_.Gid, 10, 32); err != nil {
			return gopi.ErrBadParameter.WithPrefix("-gaffer.group")
		} else {
			gid = int(gid_)
			mode = os.FileMode(0775)
		}
	}

	// Set ownership and permissions
	if err := os.Chown(folder, uid, gid); err != nil {
		return err
	} else if err := os.Chmod(folder, mode); err != nil {
		return err
	}

	// Success
	return nil
}

func LookupUser(u string) (*user.User, error) {
	// Find user/group from u
	if user_, err := user.Lookup(u); err == nil {
		return user_, nil
	} else if user_, err := user.LookupId(u); err == nil {
		return user_, nil
	} else {
		return nil, gopi.ErrNotFound.WithPrefix("user")
	}
}

func LookupGroup(g string) (*user.Group, error) {
	// Find group from g
	if group_, err := user.LookupGroup(g); err == nil {
		return group_, nil
	} else if group_, err := user.LookupGroupId(g); err == nil {
		return group_, nil
	} else {
		return nil, gopi.ErrNotFound.WithPrefix("group")
	}
}

////////////////////////////////////////////////////////////////////////////////
// EVENT HANDLER

func RPCEventHandler(_ context.Context, app gopi.App, evt gopi.Event) {
	rpcEvent := evt.(gopi.RPCEvent)
	switch rpcEvent.Type() {
	case gopi.RPC_EVENT_SERVER_STARTED:
		server := rpcEvent.Source().(gopi.RPCServer)
		app.Log().Info("Server started", server.Addr())
		if err := StartService(app); err != nil {
			app.Log().Error(err)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// MAIN

func Main(app gopi.App, args []string) error {
	// We require an rpc.sock argument. Remove any old socket which exists
	if fifo := app.Flags().GetString("rpc.sock", gopi.FLAG_NS_DEFAULT); fifo == "" {
		return fmt.Errorf("Missing required flag -rpc.sock")
	}

	// Add handler for server start and stop
	if err := app.Bus().NewHandler(gopi.EventHandler{
		Name:    "gopi.RPCEvent",
		Handler: RPCEventHandler,
		EventNS: gopi.EVENT_NS_DEFAULT,
	}); err != nil {
		return err
	}

	// Wait until CTRL+C pressed
	fmt.Println("Press CTRL+C to exit")
	app.WaitForSignal(context.Background(), os.Interrupt)
	fmt.Println("Received interrupt signal, exiting")

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// BOOTSTRAP

func main() {
	if app, err := app.NewServer(Main, "rpc/gaffer/kernel"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		// Add bootstrap arguments
		app.Flags().FlagString("gaffer.service", "", "Gaffer service binary")
		app.Flags().FlagString("gaffer.state", "", "Gaffer state folder")
		app.Flags().FlagUint("gaffer.port", 0, "Gaffer service port")
		app.Flags().FlagString("gaffer.user", "", "Gaffer service user")
		app.Flags().FlagString("gaffer.group", "", "Gaffer service group")

		// Run and exit
		os.Exit(app.Run())
	}
}
