package attributes

import (
	"fmt"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type Text struct {
	id       string
	perms    clientsapi.Permissions
	optional OptionalType
	value    string
	isSet    bool
	push     func(*clientsapi.Attribute)
	onAction func(val string)
}

func NewText(id string, perms clientsapi.Permissions, optional OptionalType) *Text {
	return &Text{
		id:       id,
		perms:    perms,
		optional: optional,
	}
}

func (attr *Text) Get() string {
	return attr.value
}

func (attr *Text) Set(value string) bool {
	attr.isSet = true
	if attr.value != value {
		attr.value = value
		if attr.push != nil {
			attr.push(attr.Pb())
		} else {
			panic(fmt.Sprintf("attribute %q is not registered with a service", attr.id))
		}
		return true
	}
	return false
}

// HandleAction calls the attribute's OnAction handler if set.
func (attr *Text) HandleAction(val *clientsapi.TextValue) {
	if attr.onAction != nil {
		attr.onAction(val.GetValue())
	}
}

// OnAction sets the handler which is called when an action is received for this
// attribute. This allows the implementer to forward the request to the end
// device.
func (attr *Text) OnAction(handler func(val string)) {
	attr.onAction = handler
}

// Static assert that Text implements the Attribute interface.
var _ Attribute = (*Text)(nil)

func (attr *Text) ID() string                            { return attr.id }
func (attr *Text) Perms() clientsapi.Permissions         { return attr.perms }
func (attr *Text) Optional() OptionalType                { return attr.optional }
func (attr *Text) IsSet() bool                           { return attr.isSet }
func (attr *Text) Push(push func(*clientsapi.Attribute)) { attr.push = push }
func (attr *Text) Pb() *clientsapi.Attribute {
	return &clientsapi.Attribute{
		Id: attr.id,
		Attr: &clientsapi.Attribute_Text{
			Text: &clientsapi.TextAttribute{
				Value: attr.value,
				Perms: attr.perms,
			},
		},
	}
}
