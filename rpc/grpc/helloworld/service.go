/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package helloworld

import (
	// Frameworks

	rpc "github.com/djthorpe/gopi-rpc/v2"
	gopi "github.com/djthorpe/gopi/v2"
	base "github.com/djthorpe/gopi/v2/base"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Service struct {
	Server rpc.Server
}

type service struct {
	base.Unit
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Unit

func (Service) Name() string { return "rpc/helloworld/service" }

func (config Service) New(log gopi.Logger) (gopi.Unit, error) {
	this := new(service)
	if err := this.Unit.Init(log); err != nil {
		return nil, err
	} else if err := this.Init(config); err != nil {
		return nil, err
	}

	// Success
	return this, nil
}

func (this *service) Init(config Service) error {
	return nil
}

func (this *service) Close() error {
	return this.Unit.Close()
}
