package services

import (
	"errors"
	"strings"

	"github.com/jimjibone/wh/v1/bridges/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
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
		Generic:         newGeneric(DefaultServiceID(clientsapi.Service_INFO), clientsapi.Service_INFO),
		Name:            attributes.NewText("name", clientsapi.Permissions_PERM_READONLY, attributes.Required),
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

// EnableRename marks the info name attribute as writable and installs the
// handler called when a user requests a device rename. Call this during
// device construction, before the device is added to the bridge. The handler
// may block while the rename is forwarded to the end device; return nil on
// success or an error which will be reported to the user.
func (srv *Info) EnableRename(handler func(newName string) error) {
	srv.Name.SetPerms(clientsapi.Permissions_PERM_READWRITE)
	srv.OnAction(func(request *clientsapi.ActionRequest, feedback func(*clientsapi.ActionResponse)) error {
		for _, val := range request.GetValues() {
			if val.GetId() != srv.Name.ID() {
				// Every other info attribute is read-only.
				return ErrReadOnly
			}
			if val.GetText() == nil {
				return ErrIncorrectTypeFor(srv.Name)
			}
			newName := strings.TrimSpace(val.GetText().GetValue())
			if newName == "" {
				return errors.New("name cannot be empty")
			}
			if err := handler(newName); err != nil {
				return err
			}
		}
		return nil
	})
}
