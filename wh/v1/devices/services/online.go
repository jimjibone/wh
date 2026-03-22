package services

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
	"github.com/jimjibone/woodhouse-4/wh/v1/devices/attributes"
)

// Static assert that Online implements the Service interface.
var _ Service = (*Online)(nil)

type Online struct {
	*Generic
	Online   *attributes.Bool // required
	LastSeen *attributes.Time // required
}

func init() {
	registerDefaultServiceID(clientsapi.Service_ONLINE, "online")
}

// New Online service. Only one of these should exist on a device.
func NewOnline() *Online {
	srv := &Online{
		Generic:  newGeneric(DefaultServiceID(clientsapi.Service_ONLINE), clientsapi.Service_ONLINE),
		Online:   attributes.NewBool("online", clientsapi.Permissions_PERM_READONLY, attributes.Required),
		LastSeen: attributes.NewTime("last_seen", clientsapi.Permissions_PERM_READONLY, attributes.Required),
	}
	srv.AddAttribute(
		srv.Online,
		srv.LastSeen,
	)
	return srv
}
