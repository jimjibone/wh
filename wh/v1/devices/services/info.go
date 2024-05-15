package services

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
	"github.com/jimjibone/woodhouse-4/wh/v1/devices/attributes"
)

// Static assert that Info implements the Service interface.
var _ Service = (*Info)(nil)

type Info struct {
	*Generic
	Name            *attributes.Text // required
	Model           *attributes.Text // optional
	Manufacturer    *attributes.Text // optional
	SerialNumber    *attributes.Text // optional
	FirmwareVersion *attributes.Text // optional
	WebUrl          *attributes.Text // optional
}

// New Info service. Only one of these should exist on a device.
func NewInfo() *Info {
	srv := &Info{
		Generic:         newGeneric("info", clientsapi.Service_INFO),
		Name:            attributes.NewText("name", clientsapi.Permissions_PERM_READWRITE, attributes.Required),
		Model:           attributes.NewText("model", clientsapi.Permissions_PERM_READONLY, attributes.Optional),
		Manufacturer:    attributes.NewText("manufacturer", clientsapi.Permissions_PERM_READONLY, attributes.Optional),
		SerialNumber:    attributes.NewText("serial_number", clientsapi.Permissions_PERM_READONLY, attributes.Optional),
		FirmwareVersion: attributes.NewText("firmware_version", clientsapi.Permissions_PERM_READONLY, attributes.Optional),
		WebUrl:          attributes.NewText("web_url", clientsapi.Permissions_PERM_READONLY, attributes.Optional),
	}
	srv.AddAttribute(
		srv.Name,
		srv.Model,
		srv.Manufacturer,
		srv.SerialNumber,
		srv.FirmwareVersion,
		srv.WebUrl,
	)
	return srv
}
