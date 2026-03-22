package services

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
	"github.com/jimjibone/woodhouse-4/wh/v1/devices/attributes"
)

// Static assert that Battery implements the Service interface.
var _ Service = (*Battery)(nil)

type Battery struct {
	*Generic
	Level   *attributes.Int   // required
	Voltage *attributes.Float // optional
}

func init() {
	registerDefaultServiceID(clientsapi.Service_BATTERY, "battery")
}

// New Battery service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "battery").
func NewBattery(id string) *Battery {
	if id == "" {
		id = DefaultServiceID(clientsapi.Service_BATTERY)
	}
	srv := &Battery{
		Generic: newGeneric(id, clientsapi.Service_BATTERY),
		Level:   attributes.NewInt("level", clientsapi.Permissions_PERM_READONLY, attributes.Required, 0, 100, 1, clientsapi.Unit_UNIT_PERCENTAGE),
		Voltage: attributes.NewFloat("voltage", clientsapi.Permissions_PERM_READONLY, attributes.Optional, 0, 0, 0, clientsapi.Unit_UNIT_VOLTS),
	}
	srv.AddAttribute(
		srv.Level,
		srv.Voltage,
	)
	return srv
}
