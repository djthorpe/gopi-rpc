/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package gaffer

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Gaffer struct {
	Path        string
	BinRoot     string
	BinOverride bool
}

type gaffer struct {
	log gopi.Logger

	config
}

type Service struct {
	// Name is unique name for the service
	Name string `json:"name"`

	// Path is the path to the executable
	Path string `json:"path"`
}

/*
	// MaxInstances determines maximum number of
	// instances which can be started at once,
	// when 0 means service is off
	MaxInstances uint `json:"max_instances"`

	// Timeout is the length of time a service should run for
	// before cancelling
	Timeout time.Duration `json:"timeout"`

	// Mode is manual or auto, which indicates if instances
	// are automatically created or manually
	Mode ServiceMode `json:"mode"`

	// Flags on the command line
	Flags []*Tuple `json:"flags"`

	// Args on the command line
	Args []string `json:"args"`
}

// Tuple defines a key-value pair for flags or environment vars
type Tuple struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ServiceMode is auto or manual
type ServiceMode uint
*/

////////////////////////////////////////////////////////////////////////////////
// COMSTANTS

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Gaffer) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<gaffer.Open>{ path=%v binroot=%v }", strconv.Quote(config.Path), strconv.Quote(config.DefaultBin()))

	this := new(gaffer)
	this.log = logger

	if err := this.config.Init(config, logger); err != nil {
		logger.Debug2("Config.Init returned nil")
		return nil, err
	}

	// Success
	return this, nil
}

func (this *gaffer) Close() error {
	this.log.Debug("<gaffer.Close>{ }")

	// Release resources, etc
	if err := this.config.Destroy(); err != nil {
		return err
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *gaffer) String() string {
	return fmt.Sprintf("<gaffer>{ %v }", this.config.String())
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *gaffer) Executables(recursive bool) []string {
	this.log.Debug2("<gaffer>Executables{ recursive=%v }", recursive)

	executables := make([]string, 0, 100)

	// Discover executables under a path
	if root, err := this.config.Root(); err != nil {
		this.log.Error("Executables: %v", err)
		return nil
	} else if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && recursive == false {
			return filepath.SkipDir
		}
		if info.Mode().IsRegular() && isExecutableFileAtPath(path) == nil {
			// Trim prefix
			sep := string(filepath.Separator)
			path := strings.TrimPrefix(strings.TrimPrefix(path, root), sep)
			// Append
			executables = append(executables, path)
		}
		return nil
	}); err != nil {
		this.log.Error("Executables: %v", err)
		return nil
	}

	return executables
}
