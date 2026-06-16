package services

import (
	"github.com/jimjibone/wh/v1/devices/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

// Static assert that Environment implements the Service interface.
var _ Service = (*Environment)(nil)

type Environment struct {
	*Generic
	Temperature *attributes.Float // optional
	Humidity    *attributes.Float // optional
	Pressure    *attributes.Float // optional
}

func init() {
	registerDefaultServiceID(clientsapi.Service_ENVIRONMENT, "environment")
}

// New Environment service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "environment").
func NewEnvironment(id string) *Environment {
	if id == "" {
		id = DefaultServiceID(clientsapi.Service_ENVIRONMENT)
	}
	srv := &Environment{
		Generic:     newGeneric(id, clientsapi.Service_ENVIRONMENT),
		Temperature: attributes.NewFloat("temperature", clientsapi.Permissions_PERM_READONLY, attributes.Optional, 0, 0, 0, clientsapi.Unit_UNIT_CELSIUS),
		Humidity:    attributes.NewFloat("humidity", clientsapi.Permissions_PERM_READONLY, attributes.Optional, 0, 0, 0, clientsapi.Unit_UNIT_PERCENTAGE),
		Pressure:    attributes.NewFloat("pressure", clientsapi.Permissions_PERM_READONLY, attributes.Optional, 0, 0, 0, clientsapi.Unit_UNIT_HECTOPASCAL),
	}
	srv.AddAttribute(
		srv.Temperature,
		srv.Humidity,
		srv.Pressure,
	)
	return srv
}
