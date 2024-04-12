package services

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
	"github.com/jimjibone/woodhouse-4/wh/v1/devices/attributes"
)

// Static assert that Switch implements the Service interface.
var _ Service = (*Switch)(nil)

type Switch struct {
	*Generic
	On *attributes.Bool // required
}

func NewSwitch() *Switch {
	srv := &Switch{
		Generic: newGeneric("switch", clientsapi.Service_SWITCH),
		On:      attributes.NewBool("on", clientsapi.Permissions_PERM_READONLY, attributes.Required),
	}
	srv.AddAttribute(
		srv.On,
	)
	return srv
}
