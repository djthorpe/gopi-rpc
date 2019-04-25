package rpc

import (
	// Frameworks
	"github.com/djthorpe/gopi"
)

// RPCFlags defines various flags which can be used for RPC
type RPCFlags uint64

type Discovery interface {
	gopi.Driver
	gopi.Publisher

	// Enumerate service names and generate
	//EnumerateServiceNames(ctx context.Context) error

	// Enumerate service instances
	//EnumerateServiceInstances(ctx context.Context, service string) error
}

const (
	RPC_FLAG_NONE RPCFlags = 0
	RPC_FLAG_IPV4 RPCFlags = (1 << iota)
	RPC_FLAG_IPV6 RPCFlags = (1 << iota)
)
