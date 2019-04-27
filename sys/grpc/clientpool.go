/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016-2019
	All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package grpc

import (
	"fmt"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ClientPool struct {
	Discovery gopi.RPCServiceDiscovery
}

type clientpool struct {
	log       gopi.Logger
	discovery gopi.RPCServiceDiscovery
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config ClientPool) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<grpc.clientpool>Open{ discovery=%v }", config.Discovery)

	this := new(clientpool)
	this.log = log
	this.discovery = config.Discovery

	// Success
	return this, nil
}

func (this *clientpool) Close() error {
	this.log.Debug("<grpc.clientpool>Close{ discovery=%v }", this.discovery)

	// Release resources
	this.discovery = nil

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *clientpool) String() string {
	return fmt.Sprintf("<grpc.clientpool>{ discovery=%v }", this.discovery)
}
