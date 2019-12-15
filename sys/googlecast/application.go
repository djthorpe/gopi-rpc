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
	"strconv"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type application struct {
	AppId        string `json:"appId"`
	DisplayName  string `json:"displayName"`
	IsIdleScreen bool   `json:"isIdleScreen"`
	SessionId    string `json:"sessionId"`
	StatusText   string `json:"statusText"`
	TransportId  string `json:"transportId"`
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION

func (this *application) ID() string {
	return this.AppId
}

func (this *application) Name() string {
	return this.DisplayName
}

func (this *application) Status() string {
	return this.StatusText
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *application) String() string {
	return fmt.Sprintf("<googlecast.Application>{ id=%v name=%v status=%v session=%v transport=%v idle_screen=%v }",
		strconv.Quote(this.AppId), strconv.Quote(this.DisplayName), strconv.Quote(this.StatusText), strconv.Quote(this.SessionId), strconv.Quote(this.TransportId), this.IsIdleScreen)
}
