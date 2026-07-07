package services

import (
	"github.com/jimjibone/wh/v1/bridges/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

// Static assert that Enum implements the Service interface.
var _ Service = (*Enum)(nil)

type Enum struct {
	*Generic
	Value *attributes.Enum // required
}

// New Enum service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "enum").
func NewEnum(id string) *Enum {
	if id == "" {
		id = DefaultServiceID(clientsapi.Service_ENUM)
	}
	srv := &Enum{
		Generic: newGeneric(id, clientsapi.Service_ENUM),
		Value:   attributes.NewEnum("value", clientsapi.Permissions_PERM_READWRITE, attributes.Required),
	}
	srv.AddAttribute(
		srv.Value,
	)
	return srv
}
