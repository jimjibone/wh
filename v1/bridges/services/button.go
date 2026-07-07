package services

import (
	"github.com/jimjibone/wh/v1/bridges/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

// Static assert that Button implements the Service interface.
var _ Service = (*Button)(nil)

type Button struct {
	*Generic
	State    *attributes.Enum     // required
	Duration *attributes.Duration // optional
}

// New Button service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "button").
func NewButton(id string) *Button {
	if id == "" {
		id = DefaultServiceID(clientsapi.Service_BUTTON)
	}
	srv := &Button{
		Generic:  newGeneric(id, clientsapi.Service_BUTTON),
		State:    attributes.NewEnum("state", clientsapi.Permissions_PERM_READONLY, attributes.Required),
		Duration: attributes.NewDuration("duration", clientsapi.Permissions_PERM_READONLY, attributes.Optional, 0, 0, 0),
	}
	srv.AddAttribute(
		srv.State,
		srv.Duration,
	)
	return srv
}
