package attributes

import (
	"fmt"
	"slices"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

type Enum struct {
	id       string
	perms    clientsapi.Permissions
	optional OptionalType
	options  []string
	value    string
	isSet    bool
	push     func(*clientsapi.Attribute)
	onAction func(val string)
}

func NewEnum(id string, perms clientsapi.Permissions, optional OptionalType) *Enum {
	return &Enum{
		id:       id,
		perms:    perms,
		optional: optional,
	}
}

func (attr *Enum) SetOptional(optional OptionalType) {
	attr.optional = optional
}

func (attr *Enum) Get() string {
	return attr.value
}

func (attr *Enum) SetOptions(options []string) bool {
	attr.isSet = true
	if !slices.Equal(attr.options, options) {
		attr.options = options
		if attr.push != nil {
			attr.push(attr.Pb())
		} else {
			panic(fmt.Sprintf("attribute %q is not registered with a service", attr.id))
		}
		return true
	}
	return false
}

func (attr *Enum) Set(value string) bool {
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
func (attr *Enum) HandleAction(val *clientsapi.EnumValue) {
	if attr.onAction != nil {
		attr.onAction(val.GetValue())
	}
}

// OnAction sets the handler which is called when an action is received for this
// attribute. This allows the implementer to forward the request to the end
// device.
func (attr *Enum) OnAction(handler func(val string)) {
	attr.onAction = handler
}

// Static assert that Enum implements the Attribute interface.
var _ Attribute = (*Enum)(nil)

func (attr *Enum) ID() string                            { return attr.id }
func (attr *Enum) Perms() clientsapi.Permissions         { return attr.perms }
func (attr *Enum) Optional() OptionalType                { return attr.optional }
func (attr *Enum) IsSet() bool                           { return attr.isSet }
func (attr *Enum) Push(push func(*clientsapi.Attribute)) { attr.push = push }
func (attr *Enum) Pb() *clientsapi.Attribute {
	return &clientsapi.Attribute{
		Id: attr.id,
		Enum: &clientsapi.EnumAttribute{
			Options: attr.options,
			Value:   attr.value,
			Perms:   attr.perms,
		},
	}
}
