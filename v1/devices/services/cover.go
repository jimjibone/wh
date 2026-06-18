package services

import (
	"github.com/jimjibone/wh/v1/devices/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

// Static assert that Cover implements the Service interface.
var _ Service = (*Cover)(nil)

type Cover struct {
	*Generic
	Position *attributes.Int  // required
	State    *attributes.Enum // required
}

// New Cover service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "cover").
func NewCover(id string) *Cover {
	if id == "" {
		id = DefaultServiceID(clientsapi.Service_COVER)
	}
	srv := &Cover{
		Generic:  newGeneric(id, clientsapi.Service_COVER),
		Position: attributes.NewInt("position", clientsapi.Permissions_PERM_READWRITE, attributes.Required, 0, 100, 1, clientsapi.Unit_UNIT_PERCENTAGE),
		State:    attributes.NewEnum("state", clientsapi.Permissions_PERM_READWRITE, attributes.Required),
	}
	srv.AddAttribute(
		srv.Position,
		srv.State,
	)
	return srv
}
