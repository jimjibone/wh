package attributes

import (
	"slices"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
)

// List holds a list of values. Build items with the values package.
type List struct{ base }

// Static assert that List implements the Attribute interface.
var _ Attribute = (*List)(nil)

func NewList(id string, perms clientsapi.Permissions, optional OptionalType, spec *clientsapi.ListSpec) *List {
	return &List{newBase(id, perms, optional, &clientsapi.Spec{Kind: &clientsapi.Spec_List{List: spec}})}
}

// Get returns the items. Do not modify them.
func (attr *List) Get() []*clientsapi.Value { return attr.value.GetList().GetItems() }

// Set sets the items, stamped by the server on receipt. Reports whether they
// changed. The items are not validated against the spec.
func (attr *List) Set(items []*clientsapi.Value) bool {
	return attr.set(values.List(slices.Clone(items)...), time.Time{})
}

// SetAt sets the items as observed by the device at t.
func (attr *List) SetAt(items []*clientsapi.Value, t time.Time) bool {
	return attr.set(values.List(slices.Clone(items)...), t)
}

// OnWrite sets the handler called with validated written items. Its error
// becomes the action's error.
func (attr *List) OnWrite(handler func(items []*clientsapi.Value) error) {
	attr.onWrite = func(v *clientsapi.Value) error { return handler(v.GetList().GetItems()) }
}
