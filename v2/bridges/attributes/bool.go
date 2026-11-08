package attributes

import (
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
)

type Bool struct{ base }

// Static assert that Bool implements the Attribute interface.
var _ Attribute = (*Bool)(nil)

func NewBool(id string, perms clientsapi.Permissions, optional OptionalType) *Bool {
	return &Bool{newBase(id, perms, optional, &clientsapi.Spec{Kind: &clientsapi.Spec_Bool{Bool: &clientsapi.BoolSpec{}}})}
}

func (attr *Bool) Get() bool { return attr.load().GetBool() }

// Set sets the value, stamped by the server on receipt. Reports whether it
// changed.
func (attr *Bool) Set(value bool) bool { return attr.set(values.Bool(value), time.Time{}) }

// SetAt sets the value as observed by the device at t.
func (attr *Bool) SetAt(value bool, t time.Time) bool { return attr.set(values.Bool(value), t) }

// OnWrite sets the handler called with a validated written value. Its error
// becomes the action's error.
func (attr *Bool) OnWrite(handler func(val bool) error) {
	attr.setOnWrite(func(v *clientsapi.Value) error { return handler(v.GetBool()) })
}
