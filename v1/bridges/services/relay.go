package services

import (
	"github.com/jimjibone/wh/v1/bridges/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

// Static assert that Relay implements the Service interface.
var _ Service = (*Relay)(nil)

type Relay struct {
	*Generic
	On          *attributes.Bool  // required
	Voltage     *attributes.Float // optional
	Current     *attributes.Float // optional
	Power       *attributes.Float // optional
	Temperature *attributes.Float // optional
}

// New Relay service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "relay").
func NewRelay(id string) *Relay {
	if id == "" {
		id = DefaultServiceID(clientsapi.Service_RELAY)
	}
	srv := &Relay{
		Generic:     newGeneric(id, clientsapi.Service_RELAY),
		On:          attributes.NewBool("on", clientsapi.Permissions_PERM_READWRITE, attributes.Required),
		Voltage:     attributes.NewFloat("voltage", clientsapi.Permissions_PERM_READONLY, attributes.Optional, 0, 0, 0, clientsapi.Unit_UNIT_VOLTS),
		Current:     attributes.NewFloat("current", clientsapi.Permissions_PERM_READONLY, attributes.Optional, 0, 0, 0, clientsapi.Unit_UNIT_AMPS),
		Power:       attributes.NewFloat("power", clientsapi.Permissions_PERM_READONLY, attributes.Optional, 0, 0, 0, clientsapi.Unit_UNIT_WATTS),
		Temperature: attributes.NewFloat("temperature", clientsapi.Permissions_PERM_READONLY, attributes.Optional, 0, 0, 0, clientsapi.Unit_UNIT_CELSIUS),
	}
	srv.AddAttribute(
		srv.On,
		srv.Voltage,
		srv.Current,
		srv.Power,
		srv.Temperature,
	)
	return srv
}
