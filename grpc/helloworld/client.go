/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package helloworld

import (

	// Frameworks

	gopi "github.com/djthorpe/gopi/v2"
	base "github.com/djthorpe/gopi/v2/base"
	// Protocol buffers
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
}

type client struct {
	base.Unit
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Unit

func (Client) Name() string { return "rpc/helloworld/client" }

func (config Client) New(log gopi.Logger) (gopi.Unit, error) {
	this := new(client)
	if err := this.Unit.Init(log); err != nil {
		return nil, err
	} else if err := this.Init(config); err != nil {
		return nil, err
	}

	// Success
	return this, nil
}

func (this *client) Init(config Client) error {
	// Success
	return nil
}

func (this *client) Close() error {
	return this.Unit.Close()
}
