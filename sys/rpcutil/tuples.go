/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package rpcutil

import (
	"regexp"
	"strconv"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Record implements a rpc.Tuples
type tuples struct {
	values []*tuple
}

type tuple struct {
	key, value string
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

var (
	reTupleKey         = regexp.MustCompile("^[A-Za-z][A-Za-z0-9\\-\\_\\.]*$")
	reTupleValueIdent  = reTupleKey
	reTupleValueDigits = regexp.MustCompile("^\\-?[0-9]+$")
)

////////////////////////////////////////////////////////////////////////////////
// NEW

func (this *util) NewTuples() rpc.Tuples {
	t := new(tuples)
	t.values = make([]*tuple, 0)
	return t
}

////////////////////////////////////////////////////////////////////////////////
// TUPLES

func (this *tuples) Strings() []string {
	strings := make([]string, len(this.values))
	for i, tuple := range this.values {
		strings[i] = tuple.String()
	}
	return strings
}

func (this *tuples) AddString(key, value string) error {
	if key == "" || reTupleKey.MatchString(key) == false {
		return gopi.ErrBadParameter
	}

	// Remove existing tuple
	if pos := this.IndexForKey(key); pos >= 0 {
		this.values = append(this.values[:pos], this.values[pos+1:]...)
	}

	// Add new tuple
	this.values = append(this.values, &tuple{key, value})

	// Return success
	return nil
}

func (this *tuples) IndexForKey(key string) int {
	for pos, tuple := range this.values {
		if tuple.key == key {
			return pos
		}
	}
	return -1
}

////////////////////////////////////////////////////////////////////////////////
// TUPLE

func (this *tuple) String() string {
	if reTupleValueIdent.MatchString(this.value) == true {
		return this.key + "=" + this.value
	}
	if reTupleValueDigits.MatchString(this.value) == true {
		return this.key + "=" + this.value
	}
	return this.key + "=" + strconv.Quote(this.value)
}
