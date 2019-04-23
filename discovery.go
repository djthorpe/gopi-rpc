package discovery

import (
	"context"

	// Frameworks
	"github.com/djthorpe/gopi"
)

type Discovery interface {
	gopi.Driver
	gopi.Publisher

	// Enumerate service names and generate
	EnumerateServiceNames(ctx context.Context) error

	// Enumerate service instances
	//EnumerateServiceInstances(ctx context.Context, service string) error
}
