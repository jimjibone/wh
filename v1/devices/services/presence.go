package services

import (
	"github.com/jimjibone/wh/v1/devices/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

// Static assert that Presence implements the Service interface.
var _ Service = (*Presence)(nil)

type Presence struct {
	*Generic
	Motion   *attributes.Bool  // required
	Presence *attributes.Bool  // required
	Distance *attributes.Float // required
}

func init() {
	registerDefaultServiceID(clientsapi.Service_PRESENCE, "presence")
}

// New Presence service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "presence").
func NewPresence(id string) *Presence {
	if id == "" {
		id = DefaultServiceID(clientsapi.Service_PRESENCE)
	}
	srv := &Presence{
		Generic:  newGeneric(id, clientsapi.Service_PRESENCE),
		Motion:   attributes.NewBool("motion", clientsapi.Permissions_PERM_READONLY, attributes.Required),
		Presence: attributes.NewBool("presence", clientsapi.Permissions_PERM_READONLY, attributes.Required),
		Distance: attributes.NewFloat("distance", clientsapi.Permissions_PERM_READONLY, attributes.Required, 0.0, 100.0, 0.01, clientsapi.Unit_UNIT_METERS),
	}
	srv.AddAttribute(
		srv.Motion,
		srv.Presence,
		srv.Distance,
	)
	return srv
}
