package attributes

import (
	"math"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
)

type Color struct{ base }

// Static assert that Color implements the Attribute interface.
var _ Attribute = (*Color)(nil)

// NewColor creates a color attribute accepting the given encodings.
func NewColor(id string, perms clientsapi.Permissions, optional OptionalType, hueSat, xy bool) *Color {
	spec := &clientsapi.ColorSpec{HueSat: hueSat, Xy: xy}
	return &Color{newBase(id, perms, optional, &clientsapi.Spec{Kind: &clientsapi.Spec_Color{Color: spec}})}
}

// Get returns the color. Either encoding may be nil.
func (attr *Color) Get() (hueSat *clientsapi.ColorHueSat, xy *clientsapi.ColorXY) {
	c := attr.value.GetColor()
	return c.GetHueSat(), c.GetXy()
}

// SetHueSat sets a hue/saturation color, stamped by the server on receipt.
// Changes of 0.001 or less are ignored. Reports whether it changed.
func (attr *Color) SetHueSat(hue, sat float64) bool {
	return attr.SetAt(&clientsapi.ColorHueSat{Hue: hue, Sat: sat}, nil, time.Time{})
}

// SetXY sets a CIE 1931 xy color, stamped by the server on receipt.
func (attr *Color) SetXY(x, y float64) bool {
	return attr.SetAt(nil, &clientsapi.ColorXY{X: x, Y: y}, time.Time{})
}

// SetAt sets the color as observed by the device at t (zero: the server
// stamps it). Either encoding may be nil.
func (attr *Color) SetAt(hueSat *clientsapi.ColorHueSat, xy *clientsapi.ColorXY, t time.Time) bool {
	oldHS, oldXY := attr.Get()
	if attr.IsSet() && closeHueSat(oldHS, hueSat) && closeXY(oldXY, xy) {
		return false
	}
	return attr.set(values.Color(hueSat, xy), t)
}

// OnWrite sets the handler called with a validated written color. Either
// encoding may be nil, never both. Its error becomes the action's error.
func (attr *Color) OnWrite(handler func(hueSat *clientsapi.ColorHueSat, xy *clientsapi.ColorXY) error) {
	attr.onWrite = func(v *clientsapi.Value) error {
		return handler(v.GetColor().GetHueSat(), v.GetColor().GetXy())
	}
}

func closeHueSat(a, b *clientsapi.ColorHueSat) bool {
	if a == nil || b == nil {
		return a == b
	}
	return math.Abs(a.Hue-b.Hue) <= 0.001 && math.Abs(a.Sat-b.Sat) <= 0.001
}

func closeXY(a, b *clientsapi.ColorXY) bool {
	if a == nil || b == nil {
		return a == b
	}
	return math.Abs(a.X-b.X) <= 0.001 && math.Abs(a.Y-b.Y) <= 0.001
}
