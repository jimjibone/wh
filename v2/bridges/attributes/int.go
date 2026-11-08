package attributes

import (
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
	"google.golang.org/protobuf/proto"
)

type Int struct{ base }

// Static assert that Int implements the Attribute interface.
var _ Attribute = (*Int)(nil)

// NewInt creates an int attribute. A nil spec means no limits and no unit.
func NewInt(id string, perms clientsapi.Permissions, optional OptionalType, spec *clientsapi.IntSpec) *Int {
	if spec == nil {
		spec = &clientsapi.IntSpec{}
	}
	return &Int{newBase(id, perms, optional, &clientsapi.Spec{Kind: &clientsapi.Spec_Int{Int: spec}})}
}

func (attr *Int) Get() int64 { return attr.load().GetInt() }

// IntSpec returns a copy of the limits and unit.
func (attr *Int) IntSpec() *clientsapi.IntSpec {
	return proto.Clone(attr.loadSpec().GetInt()).(*clientsapi.IntSpec)
}

// SetIntSpec replaces the limits and unit, e.g. with a device's own range.
func (attr *Int) SetIntSpec(spec *clientsapi.IntSpec) bool {
	return attr.setSpec(&clientsapi.Spec{Kind: &clientsapi.Spec_Int{Int: spec}})
}

// SetRange replaces min and max, keeping step and unit.
func (attr *Int) SetRange(min, max int64) bool {
	spec := attr.IntSpec()
	spec.Min, spec.Max = &min, &max
	return attr.SetIntSpec(spec)
}

// Set sets the value, stamped by the server on receipt. Reports whether it
// changed.
func (attr *Int) Set(value int64) bool { return attr.set(values.Int(value), time.Time{}) }

// SetAt sets the value as observed by the device at t.
func (attr *Int) SetAt(value int64, t time.Time) bool { return attr.set(values.Int(value), t) }

// OnWrite sets the handler called with a validated written value. Its error
// becomes the action's error.
func (attr *Int) OnWrite(handler func(val int64) error) {
	attr.setOnWrite(func(v *clientsapi.Value) error { return handler(v.GetInt()) })
}
