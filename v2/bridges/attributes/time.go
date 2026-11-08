package attributes

import (
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
)

type Time struct{ base }

// Static assert that Time implements the Attribute interface.
var _ Attribute = (*Time)(nil)

func NewTime(id string, perms clientsapi.Permissions, optional OptionalType) *Time {
	return &Time{newBase(id, perms, optional, &clientsapi.Spec{Kind: &clientsapi.Spec_Time{Time: &clientsapi.TimeSpec{}}})}
}

func (attr *Time) Get() time.Time {
	t, _ := values.AsTime(attr.load())
	return t
}

// Set sets the value, stamped by the server on receipt. Reports whether it
// changed.
func (attr *Time) Set(value time.Time) bool { return attr.set(values.Time(value), time.Time{}) }

// SetAt sets the value as observed by the device at t.
func (attr *Time) SetAt(value time.Time, t time.Time) bool { return attr.set(values.Time(value), t) }

// OnWrite sets the handler called with a validated written value. Its error
// becomes the action's error.
func (attr *Time) OnWrite(handler func(val time.Time) error) {
	attr.setOnWrite(func(v *clientsapi.Value) error {
		t, _ := values.AsTime(v)
		return handler(t)
	})
}
