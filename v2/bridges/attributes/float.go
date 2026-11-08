package attributes

import (
	"math"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
	"google.golang.org/protobuf/proto"
)

type Float struct{ base }

// Static assert that Float implements the Attribute interface.
var _ Attribute = (*Float)(nil)

// NewFloat creates a float attribute. A nil spec means no limits and no unit.
func NewFloat(id string, perms clientsapi.Permissions, optional OptionalType, spec *clientsapi.FloatSpec) *Float {
	if spec == nil {
		spec = &clientsapi.FloatSpec{}
	}
	return &Float{newBase(id, perms, optional, &clientsapi.Spec{Kind: &clientsapi.Spec_Float{Float: spec}})}
}

func (attr *Float) Get() float64 { return attr.load().GetFloat() }

// FloatSpec returns a copy of the limits and unit.
func (attr *Float) FloatSpec() *clientsapi.FloatSpec {
	return proto.Clone(attr.loadSpec().GetFloat()).(*clientsapi.FloatSpec)
}

// SetFloatSpec replaces the limits and unit.
func (attr *Float) SetFloatSpec(spec *clientsapi.FloatSpec) bool {
	return attr.setSpec(&clientsapi.Spec{Kind: &clientsapi.Spec_Float{Float: spec}})
}

// SetRange replaces min and max, keeping step and unit.
func (attr *Float) SetRange(min, max float64) bool {
	spec := attr.FloatSpec()
	spec.Min, spec.Max = &min, &max
	return attr.SetFloatSpec(spec)
}

// Set sets the value, stamped by the server on receipt. Changes of 0.0001 or
// less are ignored. Reports whether it changed.
func (attr *Float) Set(value float64) bool { return attr.SetAt(value, time.Time{}) }

// SetAt sets the value as observed by the device at t.
func (attr *Float) SetAt(value float64, t time.Time) bool {
	same := func(old *clientsapi.Value) bool { return math.Abs(old.GetFloat()-value) <= 0.0001 }
	return attr.setUnless(same, values.Float(value), t)
}

// OnWrite sets the handler called with a validated written value. Its error
// becomes the action's error.
func (attr *Float) OnWrite(handler func(val float64) error) {
	attr.setOnWrite(func(v *clientsapi.Value) error { return handler(v.GetFloat()) })
}
