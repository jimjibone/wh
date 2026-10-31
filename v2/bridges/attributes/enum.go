package attributes

import (
	"slices"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
)

type Enum struct{ base }

// Static assert that Enum implements the Attribute interface.
var _ Attribute = (*Enum)(nil)

// NewEnum creates an enum attribute. With no options any value is accepted
// until SetOptions is called.
func NewEnum(id string, perms clientsapi.Permissions, optional OptionalType, options ...string) *Enum {
	spec := &clientsapi.EnumSpec{Options: slices.Clone(options)}
	return &Enum{newBase(id, perms, optional, &clientsapi.Spec{Kind: &clientsapi.Spec_Enum{Enum: spec}})}
}

func (attr *Enum) Get() string { return attr.value.GetEnum() }

// Options returns a copy of the options.
func (attr *Enum) Options() []string { return slices.Clone(attr.spec.GetEnum().GetOptions()) }

// SetOptions replaces the options. Reports whether they changed.
func (attr *Enum) SetOptions(options []string) bool {
	spec := &clientsapi.EnumSpec{Options: slices.Clone(options)}
	return attr.setSpec(&clientsapi.Spec{Kind: &clientsapi.Spec_Enum{Enum: spec}})
}

// Set sets the value, stamped by the server on receipt. Reports whether it
// changed.
func (attr *Enum) Set(value string) bool { return attr.set(values.Enum(value), time.Time{}) }

// SetAt sets the value as observed by the device at t.
func (attr *Enum) SetAt(value string, t time.Time) bool { return attr.set(values.Enum(value), t) }

// OnWrite sets the handler called with a validated written value. Its error
// becomes the action's error.
func (attr *Enum) OnWrite(handler func(val string) error) {
	attr.onWrite = func(v *clientsapi.Value) error { return handler(v.GetEnum()) }
}
