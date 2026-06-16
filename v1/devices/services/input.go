package services

import (
	"github.com/jimjibone/wh/v1/devices/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

// Static assert that Input implements the Service interface.
var _ Service = (*Input)(nil)

type Input struct {
	*Generic
	On *attributes.Bool // required
}

func init() {
	registerDefaultServiceID(clientsapi.Service_INPUT, "input")
}

// New Input service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "input").
func NewInput(id string) *Input {
	if id == "" {
		id = DefaultServiceID(clientsapi.Service_INPUT)
	}
	srv := &Input{
		Generic: newGeneric(id, clientsapi.Service_INPUT),
		On:      attributes.NewBool("on", clientsapi.Permissions_PERM_READONLY, attributes.Required),
	}
	srv.AddAttribute(
		srv.On,
	)
	return srv
}
