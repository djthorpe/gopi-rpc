/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package discovery

import (
	"fmt"
	"net"
	"sync"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Discovery struct {
	Interface *net.Interface
	Domain    string
}

type discovery struct {
	sync.Mutex
	Config
	Listener

	log gopi.Logger
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Discovery) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<rpc.discovery.Open>{ interface=%v domain='%v' }", config.Interface, config.Domain)

	this := new(discovery)
	if err := this.Config.Init(); err != nil {
		return nil, err
	} else if err := this.Listener.Init(config); err != nil {
		return nil, err
	} else {
		this.log = logger
	}

	return this, nil
}

func (this *discovery) Close() error {
	this.log.Debug("<rpc.discovery.Close>{ config=%v listener=%v }", this.Config, this.Listener)

	// Release resources, etc
	this.Config.Deinit()

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *discovery) String() string {
	return fmt.Sprintf("<rpc.discovery>{ config=%v listener=%v }", this.Config, this.Listener)
}
