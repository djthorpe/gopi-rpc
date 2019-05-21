/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"fmt"
	"strings"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////

func RenderGroupList(groups []string) string {
	groups_ := ""
	for i, group := range groups {
		if i > 0 {
			groups_ += " "
		}
		groups_ += "@" + group
	}
	return groups_
}

func RenderFlags(flags rpc.Tuples) string {
	flags_ := flags.Flags()
	if len(flags_) == 0 {
		return "-"
	} else {
		flags__ := ""
		for i, flag := range flags_ {
			if i > 0 {
				flags__ += "\n"
			}
			flags__ += flag
		}
		return flags__
	}
}

func RenderEnv(env rpc.Tuples) string {
	env_ := env.Env()
	if len(env_) == 0 {
		return "-"
	} else {
		env__ := ""
		for i, e := range env_ {
			if i > 0 {
				env__ += "\n"
			}
			env__ += e
		}
		return env__
	}
}

func RenderMode(service rpc.GafferService) string {
	if service.InstanceCount() == 0 {
		return "disabled"
	}
	if mode := fmt.Sprint(service.Mode()); strings.HasPrefix(mode, "GAFFER_MODE_") {
		return strings.ToLower(strings.TrimPrefix(mode, "GAFFER_MODE_"))
	} else {
		return mode
	}
}

func RenderInstanceStatus(instance rpc.GafferServiceInstance) string {
	if instance.Start().IsZero() && instance.Stop().IsZero() {
		return "Starting"
	} else if instance.Stop().IsZero() == false {
		// Stopped
		return fmt.Sprintf("Exit code %v", instance.ExitCode())
	} else if instance.Start().IsZero() == false {
		dur := time.Now().Sub(instance.Start()).Truncate(time.Minute)
		return fmt.Sprintf("Running %dm", uint(dur.Minutes()))
	}

	// Unhandled status
	return "??"
}

func RenderDuration(duration time.Duration) string {
	if duration == 0 {
		return "-"
	}
	return fmt.Sprint(duration.Truncate(time.Second))
}

func RenderHost(service gopi.RPCServiceRecord) string {
	if service.Port() == 0 {
		return service.Host()
	} else {
		return fmt.Sprintf("%v:%v", service.Host(), service.Port())
	}
}

func RenderService(service gopi.RPCServiceRecord) string {
	if service.Subtype() == "" {
		return service.Service()
	} else {
		return fmt.Sprintf("%v, %v", service.Subtype(), service.Service())
	}
}

func RenderIP(service gopi.RPCServiceRecord) string {
	ips := make([]string, 0)
	for _, ip := range service.IP4() {
		ips = append(ips, fmt.Sprint(ip))
	}
	for _, ip := range service.IP6() {
		ips = append(ips, fmt.Sprint(ip))
	}
	return strings.Join(ips, "\n")
}

func RenderTxt(service gopi.RPCServiceRecord) string {
	return strings.Join(service.Text(), "\n")
}
