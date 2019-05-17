/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package rpc

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Tuples struct {
	tuples []*tuple
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
// PUBLIC METHODS

// Len returns the number of tuples
func (this *Tuples) Len() int {
	return len(this.tuples)
}

// Copy returns a copy of the tuples
func (this Tuples) Copy() Tuples {
	var that Tuples
	that.tuples = make([]*tuple, len(this.tuples))
	for i, t := range this.tuples {
		that.tuples[i] = &tuple{t.key, t.value}
	}
	return that
}

// Keys returns an array of keys
func (this *Tuples) Keys() []string {
	keys := make([]string, len(this.tuples))
	for i, tuple := range this.tuples {
		keys[i] = tuple.key
	}
	return keys
}

// RemoveAll removes all tuples
func (this *Tuples) RemoveAll() {
	this.tuples = make([]*tuple, 0, cap(this.tuples))
}

// SetStringForKey sets a tuple key-value pair. Returns error
// if a key is invalid
func (this *Tuples) SetStringForKey(k, v string) error {
	// Create tuples if nil
	if this.tuples == nil {
		this.RemoveAll()
	}
	// Replace or append tuples by key
	if pos := this.indexForKey(k); pos == -1 {
		// Check key
		if reTupleKey.MatchString(k) == false {
			return fmt.Errorf("Invalid key: %v", strconv.Quote(k))
		} else {
			this.tuples = append(this.tuples, &tuple{k, v})
		}
	} else {
		this.tuples[pos] = &tuple{k, v}
	}
	// Success
	return nil
}

// StringForKey returns the string value for a key or an
// empty string if a keyed tuple was not found
func (this *Tuples) StringForKey(k string) string {
	if pos := this.indexForKey(k); pos >= 0 {
		return this.tuples[pos].value
	} else {
		return ""
	}
}

// ExistsForKey returns true if a key is present
func (this *Tuples) ExistsForKey(k string) bool {
	if pos := this.indexForKey(k); pos >= 0 {
		return true
	} else {
		return false
	}
}

// String returns the string representation of the tuples
func (this Tuples) String() string {
	str := ""
	if len(this.tuples) == 0 {
		str = "<nil>"
	} else {
		for i, t := range this.tuples {
			if i > 0 {
				str += ","
			}
			str += t.String()
		}
	}
	return fmt.Sprintf("<Tuples>{ %v }", str)
}

// Flags returns tuples as a set of flags, including the initial '-' character
func (this Tuples) Flags() []string {
	strs := make([]string, len(this.tuples))
	for i, t := range this.tuples {
		strs[i] = t.Flag()
	}
	return strs
}

// Env returns tuples as a set of key value parameters
func (this Tuples) Env() []string {
	strs := make([]string, len(this.tuples))
	for i, t := range this.tuples {
		strs[i] = t.String()
	}
	return strs
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *Tuples) indexForKey(k string) int {
	for i, tuple := range this.tuples {
		if tuple.key == k {
			return i
		}
	}
	return -1
}

func (this *tuple) String() string {
	return fmt.Sprintf("%v=%v", this.key, strconv.Quote(this.value))
}

func (this *tuple) Flag() string {
	if this.value == "" {
		return fmt.Sprintf("-%v", this.key)
	} else if reTupleValueDigits.MatchString(this.value) || reTupleValueIdent.MatchString(this.value) {
		return fmt.Sprintf("-%v=%v", this.key, this.value)
	} else {
		return fmt.Sprintf("%v=%v", this.key, strconv.Quote(this.value))
	}
}

////////////////////////////////////////////////////////////////////////////////
// JSONIFY

func (t Tuples) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Env())
}

func (t *Tuples) UnmarshalJSON(data []byte) error {
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	if len(arr) == 0 {
		t.RemoveAll()
		return nil
	}
	for _, str := range arr {
		key_value := strings.SplitN(str, "=", 2)
		if len(key_value) == 1 {
			if err := t.SetStringForKey(key_value[0], ""); err != nil {
				return err
			}
		} else if len(key_value) == 2 {
			if err := t.SetStringForKey(key_value[0], key_value[1]); err != nil {
				return err
			}
		} else {
			return gopi.ErrBadParameter
		}
	}
	return nil
}
