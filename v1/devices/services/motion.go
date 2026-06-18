package services

import (
	"github.com/jimjibone/wh/v1/devices/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

// Static assert that Motion implements the Service interface.
var _ Service = (*Motion)(nil)

type Motion struct {
	*Generic
	Motion *attributes.Bool // required
}

// New Motion service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "motion").
func NewMotion(id string) *Motion {
	if id == "" {
		id = DefaultServiceID(clientsapi.Service_MOTION)
	}
	srv := &Motion{
		Generic: newGeneric(id, clientsapi.Service_MOTION),
		Motion:  attributes.NewBool("motion", clientsapi.Permissions_PERM_READONLY, attributes.Required),
	}
	srv.AddAttribute(
		srv.Motion,
	)
	return srv
}
