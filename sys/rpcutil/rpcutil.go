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
	"encoding/json"
	"io"

	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

type Util struct {
	// No members
}

type util struct {
	// No members
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open a client
func (config Util) Open(log gopi.Logger) (gopi.Driver, error) {
	return new(util), nil
}

func (this *util) Close() error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// READ AND WRITE

// Writer writes an array of service records to a io.Writer object
func (this *util) Writer(fh io.Writer, records []rpc.ServiceRecord, indent bool) error {
	enc := json.NewEncoder(fh)
	if indent {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(records); err != nil {
		return err
	}
	// Success
	return nil
}

// Reader reads the configuration from an io.Reader object
func (this *util) Reader(fh io.Reader) ([]rpc.ServiceRecord, error) {
	dec := json.NewDecoder(fh)
	var records []*record
	if err := dec.Decode(&records); err != nil {
		return nil, err
	} else {
		services := make([]rpc.ServiceRecord, 0, len(records))
		for _, record := range records {
			if record.Expired() == false {
				services = append(services, record)
			}
		}
		return services, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *util) String() string {
	return "<rpc.util>{}"
}
