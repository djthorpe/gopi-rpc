/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package rpcutil

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Record implements a rpc.Tuples
type tuples struct {
}

////////////////////////////////////////////////////////////////////////////////
// NEW

func (this *util) NewTuples() rpc.Tuples {
	return new(tuples)
}

////////////////////////////////////////////////////////////////////////////////
// RETURN TUPLES AS STRING

func (this *tuples) Strings() []string {
	return nil
}

func (this *tuples) AddString(key, value string) error {
	return gopi.ErrNotImplemented
}
