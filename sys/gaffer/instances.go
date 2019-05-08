/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package gaffer

import (
	"fmt"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type instances struct {
	// Private Members
	log       gopi.Logger
	instances map[*Service]*ServiceInstance
}

////////////////////////////////////////////////////////////////////////////////
// INIT / DESTROY

func (this *instances) Init(config Gaffer, logger gopi.Logger) error {
	logger.Debug("<gaffer.instances.Init>{ config=%+v }", config)

	this.log = logger
	this.instances = make(map[*Service]*ServiceInstance, 0)

	// Success
	return nil
}

func (this *instances) Destroy() error {
	this.log.Debug("<gaffer.instances.Destroy>{ }")

	// TODO: Stop instances

	// Release resources
	this.instances = nil

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *instances) String() string {
	return fmt.Sprintf("<instances>{ }")
}
