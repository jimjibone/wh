package services

import (
	"github.com/jimjibone/wh/v1/devices/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

// Static assert that Contact implements the Service interface.
var _ Service = (*Contact)(nil)

type Contact struct {
	*Generic
	Closed *attributes.Bool // required
}

func init() {
	registerDefaultServiceID(clientsapi.Service_CONTACT, "contact")
}

// New Contact service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "contact").
func NewContact(id string) *Contact {
	if id == "" {
		id = DefaultServiceID(clientsapi.Service_CONTACT)
	}
	srv := &Contact{
		Generic: newGeneric(id, clientsapi.Service_CONTACT),
		Closed:  attributes.NewBool("closed", clientsapi.Permissions_PERM_READONLY, attributes.Required),
	}
	srv.AddAttribute(
		srv.Closed,
	)
	return srv
}
