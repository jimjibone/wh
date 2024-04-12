package services

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
	"github.com/jimjibone/woodhouse-4/wh/v1/devices/attributes"
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

func NewRelay() *Relay {
	srv := &Relay{
		Generic:     newGeneric("relay", clientsapi.Service_RELAY),
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
