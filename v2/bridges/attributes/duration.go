package attributes

import (
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
	"google.golang.org/protobuf/proto"
)

type Duration struct{ base }

// Static assert that Duration implements the Attribute interface.
var _ Attribute = (*Duration)(nil)

// NewDuration creates a duration attribute. A nil spec means no limits.
func NewDuration(id string, perms clientsapi.Permissions, optional OptionalType, spec *clientsapi.DurationSpec) *Duration {
	if spec == nil {
		spec = &clientsapi.DurationSpec{}
	}
	return &Duration{newBase(id, perms, optional, &clientsapi.Spec{Kind: &clientsapi.Spec_Duration{Duration: spec}})}
}

func (attr *Duration) Get() time.Duration {
	d, _ := values.AsDuration(attr.load())
	return d
}

// DurationSpec returns a copy of the limits.
func (attr *Duration) DurationSpec() *clientsapi.DurationSpec {
	return proto.Clone(attr.loadSpec().GetDuration()).(*clientsapi.DurationSpec)
}

// SetRange replaces min and max, keeping step.
func (attr *Duration) SetRange(min, max time.Duration) bool {
	spec := attr.DurationSpec()
	lo, hi := min.Milliseconds(), max.Milliseconds()
	spec.Min, spec.Max = &lo, &hi
	return attr.setSpec(&clientsapi.Spec{Kind: &clientsapi.Spec_Duration{Duration: spec}})
}

// Set sets the value (millisecond resolution), stamped by the server on
// receipt. Reports whether it changed.
func (attr *Duration) Set(value time.Duration) bool {
	return attr.set(values.Duration(value), time.Time{})
}

// SetAt sets the value as observed by the device at t.
func (attr *Duration) SetAt(value time.Duration, t time.Time) bool {
	return attr.set(values.Duration(value), t)
}

// OnWrite sets the handler called with a validated written value. Its error
// becomes the action's error.
func (attr *Duration) OnWrite(handler func(val time.Duration) error) {
	attr.setOnWrite(func(v *clientsapi.Value) error {
		d, _ := values.AsDuration(v)
		return handler(d)
	})
}
