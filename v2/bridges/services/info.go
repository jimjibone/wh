package services

import (
	"errors"
	"strings"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
)

// setup runs at the end of NewInfo. The schema allows writing info.name, but
// only bridges that can rename their devices enable it.
func (srv *Info) setup() {
	srv.Name.SetPerms(clientsapi.Permissions_PERM_READONLY)
}

// EnableRename marks the name attribute as writable and installs the handler
// called when a user renames the device. Call this during device
// construction, before the device is added to the bridge. The handler gets
// the trimmed, non-empty new name and may block while the rename is forwarded
// to the device; return nil on success or an error to report to the user.
func (srv *Info) EnableRename(handler func(newName string) error) {
	srv.Name.SetPerms(clientsapi.Permissions_PERM_READWRITE)
	srv.Name.OnWrite(func(val string) error {
		newName := strings.TrimSpace(val)
		if newName == "" {
			return errors.New("name cannot be empty")
		}
		return handler(newName)
	})
}
