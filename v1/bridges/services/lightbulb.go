package services

import (
	"time"

	"github.com/jimjibone/wh/v1/bridges/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

// Static assert that Lightbulb implements the Service interface.
var _ Service = (*Lightbulb)(nil)

// Values of the Lightbulb ColorMode attribute. ColorMode is read-only and
// follows whichever of ColorTemp or Color the device applied last. Clients
// switch mode by writing color_temp or color. Bridges call ColorMode.SetOptions
// with the modes the bulb supports and ColorMode.Set on state updates.
const (
	LightbulbColorModeColorTemp = "color_temp" // ColorTemp attribute is authoritative
	LightbulbColorModeColor     = "color"      // Color attribute (hue/sat or xy) is authoritative
)

type Lightbulb struct {
	*Generic
	On         *attributes.Bool     // required
	Brightness *attributes.Int      // optional
	ColorTemp  *attributes.Int      // optional
	Color      *attributes.Color    // optional
	ColorMode  *attributes.Enum     // optional
	Transition *attributes.Duration // optional
}

// New Lightbulb service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "lightbulb").
func NewLightbulb(id string) *Lightbulb {
	if id == "" {
		id = DefaultServiceID(clientsapi.Service_LIGHTBULB)
	}
	srv := &Lightbulb{
		Generic:    newGeneric(id, clientsapi.Service_LIGHTBULB),
		On:         attributes.NewBool("on", clientsapi.Permissions_PERM_READWRITE, attributes.Required),
		Brightness: attributes.NewInt("brightness", clientsapi.Permissions_PERM_READWRITE, attributes.Optional, 0, 100, 1, clientsapi.Unit_UNIT_PERCENTAGE),
		ColorTemp:  attributes.NewInt("color_temp", clientsapi.Permissions_PERM_READWRITE, attributes.Optional, 153, 555, 1, clientsapi.Unit_UNIT_MIREDS),
		Color:      attributes.NewColor("color", clientsapi.Permissions_PERM_READWRITE, attributes.Optional),
		ColorMode:  attributes.NewEnum("color_mode", clientsapi.Permissions_PERM_READONLY, attributes.Optional),
		Transition: attributes.NewDuration("transition", clientsapi.Permissions_PERM_WRITEONLY, attributes.Optional, 0, 300*time.Second, time.Second),
	}
	srv.AddAttribute(
		srv.On,
		srv.Brightness,
		srv.ColorTemp,
		srv.Color,
		srv.ColorMode,
		srv.Transition,
	)
	return srv
}
