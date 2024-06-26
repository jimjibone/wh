package services

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
	"github.com/jimjibone/woodhouse-4/wh/v1/devices/attributes"
)

// Static assert that Update implements the Service interface.
var _ Service = (*Update)(nil)

type Update struct {
	*Generic
	Available      *attributes.Bool // required
	CurrentVersion *attributes.Text // optional
	UpdateVersion  *attributes.Text // optional
}

// New Update service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "update").
func NewUpdate(id string) *Update {
	if id == "" {
		id = "update"
	}
	srv := &Update{
		Generic:        newGeneric(id, clientsapi.Service_UPDATE),
		Available:      attributes.NewBool("available", clientsapi.Permissions_PERM_READONLY, attributes.Required),
		CurrentVersion: attributes.NewText("current_version", clientsapi.Permissions_PERM_READONLY, attributes.Optional),
		UpdateVersion:  attributes.NewText("update_version", clientsapi.Permissions_PERM_READONLY, attributes.Optional),
	}
	srv.AddAttribute(
		srv.Available,
		srv.CurrentVersion,
		srv.UpdateVersion,
	)
	return srv
}
