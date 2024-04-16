package attributes

import (
	"fmt"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type Bool struct {
	id       string
	perms    clientsapi.Permissions
	optional OptionalType
	value    bool
	isSet    bool
	push     func(*clientsapi.Attribute)
	onAction func(val bool)
}

func NewBool(id string, perms clientsapi.Permissions, optional OptionalType) *Bool {
	return &Bool{
		id:       id,
		perms:    perms,
		optional: optional,
	}
}

func (attr *Bool) Get() bool {
	return attr.value
}

func (attr *Bool) Set(value bool) bool {
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
func (attr *Bool) HandleAction(val *clientsapi.BoolValue) {
	if attr.onAction != nil {
		attr.onAction(val.GetValue())
	}
}

// OnAction sets the handler which is called when an action is received for this
// attribute. This allows the implementer to forward the request to the end
// device.
func (attr *Bool) OnAction(handler func(val bool)) {
	attr.onAction = handler
}

// Static assert that Bool implements the Attribute interface.
var _ Attribute = (*Bool)(nil)

func (attr *Bool) ID() string                            { return attr.id }
func (attr *Bool) Perms() clientsapi.Permissions         { return attr.perms }
func (attr *Bool) Optional() OptionalType                { return attr.optional }
func (attr *Bool) IsSet() bool                           { return attr.isSet }
func (attr *Bool) Push(push func(*clientsapi.Attribute)) { attr.push = push }
func (attr *Bool) Pb() *clientsapi.Attribute {
	return &clientsapi.Attribute{
		Id: attr.id,
		Bool: &clientsapi.BoolAttribute{
			Value: attr.value,
			Perms: attr.perms,
		},
	}
}
