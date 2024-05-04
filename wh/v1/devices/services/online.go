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

func NewOnline() *Online {
	return NewOnlineID("online")
}

func NewOnlineID(id string) *Online {
	srv := &Online{
		Generic:  newGeneric(id, clientsapi.Service_ONLINE),
		Online:   attributes.NewBool("online", clientsapi.Permissions_PERM_READONLY, attributes.Required),
		LastSeen: attributes.NewTime("last_seen", clientsapi.Permissions_PERM_READONLY, attributes.Required),
	}
	srv.AddAttribute(
		srv.Online,
		srv.LastSeen,
	)
	return srv
}
