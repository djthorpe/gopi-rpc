/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package googlecast

import (
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type volume struct {
	Level_ float32 `json:"level,omitempty"`
	Muted_ bool    `json:"muted"`
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION

func (this *volume) Level() float32 {
	return this.Level_
}

func (this *volume) Muted() bool {
	return this.Muted_
}

func (this *volume) String() string {
	return fmt.Sprintf("<volume>{ level=%.2f muted=%v }", this.Level_, this.Muted_)
}

func (this *volume) Equals(other *volume) bool {
	if this.Level_ != other.Level_ {
		return false
	} else if this.Muted_ != other.Muted_ {
		return false
	} else {
		return true
	}
}
