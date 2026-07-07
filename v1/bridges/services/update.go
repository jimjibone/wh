package services

import (
	"time"

	"github.com/jimjibone/wh/v1/bridges/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

// Static assert that Update implements the Service interface.
var _ Service = (*Update)(nil)

type Update struct {
	*Generic
	Available      *attributes.Bool     // required
	CurrentVersion *attributes.Text     // optional
	UpdateVersion  *attributes.Text     // optional
	StartUpdate    *attributes.Bool     // optional
	Updating       *attributes.Bool     // optional
	Progress       *attributes.Int      // optional
	Remaining      *attributes.Duration // optional
}

// New Update service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "update").
func NewUpdate(id string) *Update {
	if id == "" {
		id = DefaultServiceID(clientsapi.Service_UPDATE)
	}
	srv := &Update{
		Generic:        newGeneric(id, clientsapi.Service_UPDATE),
		Available:      attributes.NewBool("available", clientsapi.Permissions_PERM_READONLY, attributes.Required),
		CurrentVersion: attributes.NewText("current_version", clientsapi.Permissions_PERM_READONLY, attributes.Optional),
		UpdateVersion:  attributes.NewText("update_version", clientsapi.Permissions_PERM_READONLY, attributes.Optional),
		StartUpdate:    attributes.NewBool("start_update", clientsapi.Permissions_PERM_WRITEONLY, attributes.Optional),
		Updating:       attributes.NewBool("updating", clientsapi.Permissions_PERM_READONLY, attributes.Optional),
		Progress:       attributes.NewInt("progress", clientsapi.Permissions_PERM_READONLY, attributes.Optional, 0, 100, 1, clientsapi.Unit_UNIT_PERCENTAGE),
		Remaining:      attributes.NewDuration("remaining", clientsapi.Permissions_PERM_READONLY, attributes.Optional, 0, time.Hour, time.Second),
	}
	srv.AddAttribute(
		srv.Available,
		srv.CurrentVersion,
		srv.UpdateVersion,
		srv.StartUpdate,
		srv.Updating,
		srv.Progress,
		srv.Remaining,
	)
	return srv
}
