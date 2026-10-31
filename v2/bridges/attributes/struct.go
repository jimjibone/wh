package attributes

import (
	"slices"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
)

// Struct holds named fields. Build fields with the values package.
type Struct struct{ base }

// Static assert that Struct implements the Attribute interface.
var _ Attribute = (*Struct)(nil)

func NewStruct(id string, perms clientsapi.Permissions, optional OptionalType, spec *clientsapi.StructSpec) *Struct {
	return &Struct{newBase(id, perms, optional, &clientsapi.Spec{Kind: &clientsapi.Spec_Struct{Struct: spec}})}
}

// Get returns the fields. Do not modify them.
func (attr *Struct) Get() []*clientsapi.FieldValue { return attr.value.GetStruct().GetFields() }

// Set sets the fields, stamped by the server on receipt. Reports whether they
// changed. The fields are not validated against the spec.
func (attr *Struct) Set(fields []*clientsapi.FieldValue) bool {
	return attr.set(values.Struct(slices.Clone(fields)...), time.Time{})
}

// SetAt sets the fields as observed by the device at t.
func (attr *Struct) SetAt(fields []*clientsapi.FieldValue, t time.Time) bool {
	return attr.set(values.Struct(slices.Clone(fields)...), t)
}

// OnWrite sets the handler called with validated written fields. Its error
// becomes the action's error.
func (attr *Struct) OnWrite(handler func(fields []*clientsapi.FieldValue) error) {
	attr.onWrite = func(v *clientsapi.Value) error { return handler(v.GetStruct().GetFields()) }
}
