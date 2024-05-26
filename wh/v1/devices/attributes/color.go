package attributes

import (
	"fmt"
	"math"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type Color struct {
	id       string
	perms    clientsapi.Permissions
	optional OptionalType
	hue      float64
	sat      float64
	x        float64
	y        float64
	isSet    bool
	push     func(*clientsapi.Attribute)
	onAction func(huesat *clientsapi.ColorHueSat, xy *clientsapi.ColorXY)
}

func NewColor(id string, perms clientsapi.Permissions, optional OptionalType) *Color {
	return &Color{
		id:       id,
		perms:    perms,
		optional: optional,
	}
}

func (attr *Color) SetOptional(optional OptionalType) {
	attr.optional = optional
}

func (attr *Color) Get() (hue, sat, x, y float64) {
	return attr.hue, attr.sat, attr.x, attr.y
}

func (attr *Color) SetHueSat(hue, sat float64) bool {
	attr.isSet = true
	if math.Abs(attr.hue-hue) > 0.001 || math.Abs(attr.sat-sat) > 0.001 {
		attr.hue = hue
		attr.sat = sat
		if attr.push != nil {
			attr.push(attr.Pb())
		} else {
			panic(fmt.Sprintf("attribute %q is not registered with a service", attr.id))
		}
		return true
	}
	return false
}

func (attr *Color) SetXY(x, y float64) bool {
	attr.isSet = true
	if math.Abs(attr.x-x) > 0.001 || math.Abs(attr.y-y) > 0.001 {
		attr.x = x
		attr.y = y
		if attr.push != nil {
			attr.push(attr.Pb())
		} else {
			panic(fmt.Sprintf("attribute %q is not registered with a service", attr.id))
		}
		return true
	}
	return false
}

// HandleAction calls the attribute's OnAction handler if set.
func (attr *Color) HandleAction(val *clientsapi.ColorValue) {
	if attr.onAction != nil {
		attr.onAction(val.GetHueSat(), val.GetXy())
	}
}

// OnAction sets the handler which is called when an action is received for this
// attribute. This allows the implementer to forward the request to the end
// device.
func (attr *Color) OnAction(handler func(huesat *clientsapi.ColorHueSat, xy *clientsapi.ColorXY)) {
	attr.onAction = handler
}

// Static assert that Color implements the Attribute interface.
var _ Attribute = (*Color)(nil)

func (attr *Color) ID() string                            { return attr.id }
func (attr *Color) Perms() clientsapi.Permissions         { return attr.perms }
func (attr *Color) Optional() OptionalType                { return attr.optional }
func (attr *Color) IsSet() bool                           { return attr.isSet }
func (attr *Color) Push(push func(*clientsapi.Attribute)) { attr.push = push }
func (attr *Color) Pb() *clientsapi.Attribute {
	return &clientsapi.Attribute{
		Id: attr.id,
		Color: &clientsapi.ColorAttribute{
			HueSat: &clientsapi.ColorHueSat{Hue: attr.hue, Sat: attr.sat},
			Xy:     &clientsapi.ColorXY{X: attr.x, Y: attr.y},
			Perms:  attr.perms,
		},
	}
}
