package services

import (
	"github.com/jimjibone/wh/v1/bridges/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

// Static assert that Climate implements the Service interface.
var _ Service = (*Climate)(nil)

type Climate struct {
	*Generic
	HeatingSetpoint  *attributes.Float // required
	LocalTemperature *attributes.Float // required
	HeatingDemand    *attributes.Bool  // optional
	PIHeatingDemand  *attributes.Int   // optional
	ValvePosition    *attributes.Int   // optional
}

// New Climate service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "climate").
func NewClimate(id string) *Climate {
	if id == "" {
		id = DefaultServiceID(clientsapi.Service_CLIMATE)
	}
	srv := &Climate{
		Generic:          newGeneric(id, clientsapi.Service_CLIMATE),
		HeatingSetpoint:  attributes.NewFloat("heating_setpoint", clientsapi.Permissions_PERM_READWRITE, attributes.Required, 5, 30, 0.5, clientsapi.Unit_UNIT_CELSIUS),
		LocalTemperature: attributes.NewFloat("local_temperature", clientsapi.Permissions_PERM_READONLY, attributes.Required, 0, 0, 0, clientsapi.Unit_UNIT_CELSIUS),
		HeatingDemand:    attributes.NewBool("heating_demand", clientsapi.Permissions_PERM_READONLY, attributes.Optional),
		PIHeatingDemand:  attributes.NewInt("pi_heating_demand", clientsapi.Permissions_PERM_READONLY, attributes.Optional, 0, 100, 1, clientsapi.Unit_UNIT_PERCENTAGE),
		ValvePosition:    attributes.NewInt("valve_position", clientsapi.Permissions_PERM_READWRITE, attributes.Optional, 0, 100, 1, clientsapi.Unit_UNIT_PERCENTAGE),
	}
	srv.AddAttribute(
		srv.HeatingSetpoint,
		srv.LocalTemperature,
		srv.HeatingDemand,
		srv.PIHeatingDemand,
		srv.ValvePosition,
	)
	return srv
}
