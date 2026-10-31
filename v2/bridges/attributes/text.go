package attributes

import (
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
)

type Text struct{ base }

// Static assert that Text implements the Attribute interface.
var _ Attribute = (*Text)(nil)

func NewText(id string, perms clientsapi.Permissions, optional OptionalType) *Text {
	return &Text{newBase(id, perms, optional, &clientsapi.Spec{Kind: &clientsapi.Spec_Text{Text: &clientsapi.TextSpec{}}})}
}

func (attr *Text) Get() string { return attr.value.GetText() }

// Set sets the value, stamped by the server on receipt. Reports whether it
// changed.
func (attr *Text) Set(value string) bool { return attr.set(values.Text(value), time.Time{}) }

// SetAt sets the value as observed by the device at t.
func (attr *Text) SetAt(value string, t time.Time) bool { return attr.set(values.Text(value), t) }

// OnWrite sets the handler called with a validated written value. Its error
// becomes the action's error.
func (attr *Text) OnWrite(handler func(val string) error) {
	attr.onWrite = func(v *clientsapi.Value) error { return handler(v.GetText()) }
}
