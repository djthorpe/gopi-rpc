/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package rpc

import (
	// Frameworks
	"github.com/djthorpe/gopi"
)

type Gaffer interface {
	gopi.Driver

	// Return list of executables
	Executables(recursive bool) []string

	// Return a new service
	AddService(executable string) (GafferService, error)
}

type GafferService interface {
	Name() string
}
