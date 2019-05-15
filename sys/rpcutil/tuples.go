/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package rpcutil

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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

// Copy tuples
func (this *tuples) Copy() rpc.Tuples {
	that := new(tuples)
	that.values = make([]*tuple, len(this.values))
	for i, value := range this.values {
		that.values[i] = &tuple{value.key, value.value}
	}
	return that
}

// Merge tuples in from another tuple set
func (this *tuples) Merge(that rpc.Tuples) error {
	for _, value := range that.(*tuples).values {
		if this.IndexForKey(value.key) >= 0 {
			// Do not add
		} else if err := this.AddString(value.key, value.value); err != nil {
			return err
		}
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// TUPLE

func NewTuple(keyvalue string) *tuple {
	if keyvalue == "" {
		return nil
	} else if arr := strings.SplitN(keyvalue, "=", 2); len(arr) == 0 {
		return nil
	} else if reTupleKey.MatchString(arr[0]) == false {
		return nil
	} else if len(arr) == 1 {
		return &tuple{arr[0], ""}
	} else if len(arr) > 2 {
		return nil
	} else {
		return &tuple{arr[0], arr[1]}
	}
}

func (this *tuple) String() string {
	if reTupleValueIdent.MatchString(this.value) == true {
		return this.key + "=" + this.value
	}
	if reTupleValueDigits.MatchString(this.value) == true {
		return this.key + "=" + this.value
	}
	return this.key + "=" + strconv.Quote(this.value)
}

////////////////////////////////////////////////////////////////////////////////
// JSON ENCODE AND DECODE

func (this *tuples) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.Strings())
}

func (this *tuples) UnmarshalJSON(data []byte) error {
	var strs []string
	if err := json.Unmarshal(data, &strs); err != nil {
		return err
	} else {
		this.values = make([]*tuple, len(strs))
		for i, str := range strs {
			if tuple := NewTuple(str); tuple == nil {
				return fmt.Errorf("Syntax error: %v", strconv.Quote(str))
			} else if this.IndexForKey(tuple.key) >= 0 {
				return fmt.Errorf("Duplicate key: %v", strconv.Quote(tuple.key))
			} else {
				this.values[i] = tuple
			}
		}
		return nil
	}
}
