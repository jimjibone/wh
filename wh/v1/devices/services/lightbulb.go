package services

import (
	"time"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
	"github.com/jimjibone/woodhouse-4/wh/v1/devices/attributes"
)

// Static assert that Lightbulb implements the Service interface.
var _ Service = (*Lightbulb)(nil)

type Lightbulb struct {
	*Generic
	On         *attributes.Bool     // required
	Brightness *attributes.Int      // optional
	Saturation *attributes.Int      // optional
	Hue        *attributes.Float    // optional
	Transition *attributes.Duration // optional
}

// New Lightbulb service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "lightbulb").
func NewLightbulb(id string) *Lightbulb {
	srv := &Lightbulb{
		Generic:    newGeneric(id, clientsapi.Service_LIGHTBULB),
		On:         attributes.NewBool("on", clientsapi.Permissions_PERM_READWRITE, attributes.Required),
		Brightness: attributes.NewInt("brightness", clientsapi.Permissions_PERM_READWRITE, attributes.Optional, 0, 100, 1, clientsapi.Unit_UNIT_PERCENTAGE),
		Saturation: attributes.NewInt("saturation", clientsapi.Permissions_PERM_READWRITE, attributes.Optional, 0, 100, 1, clientsapi.Unit_UNIT_PERCENTAGE),
		Hue:        attributes.NewFloat("hue", clientsapi.Permissions_PERM_READWRITE, attributes.Optional, 0.0, 360.0, 0.0, clientsapi.Unit_UNIT_ARC_DEGREES),
		Transition: attributes.NewDuration("transition", clientsapi.Permissions_PERM_WRITEONLY, attributes.Optional, 0, 300*time.Second, time.Second),
	}
	srv.AddAttribute(
		srv.On,
		srv.Brightness,
		srv.Saturation,
		srv.Hue,
		srv.Transition,
	)
	return srv
}
