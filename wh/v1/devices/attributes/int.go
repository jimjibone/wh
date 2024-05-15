package attributes

import (
	"fmt"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type Int struct {
	id       string
	perms    clientsapi.Permissions
	optional OptionalType
	min      int64
	max      int64
	step     uint64
	unit     clientsapi.Unit
	value    int64
	isSet    bool
	push     func(*clientsapi.Attribute)
	onAction func(val int64)
}

func NewInt(id string, perms clientsapi.Permissions, optional OptionalType, min, max int64, step uint64, unit clientsapi.Unit) *Int {
	return &Int{
		id:       id,
		perms:    perms,
		optional: optional,
		min:      min,
		max:      max,
		step:     step,
		unit:     unit,
	}
}

func (attr *Int) SetOptional(optional OptionalType) {
	attr.optional = optional
}

func (attr *Int) Get() int64 {
	return attr.value
}

func (attr *Int) Set(value int64) bool {
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
func (attr *Int) HandleAction(val *clientsapi.IntValue) {
	if attr.onAction != nil {
		attr.onAction(val.GetValue())
	}
}

// OnAction sets the handler which is called when an action is received for this
// attribute. This allows the implementer to forward the request to the end
// device.
func (attr *Int) OnAction(handler func(val int64)) {
	attr.onAction = handler
}

// Static assert that Int implements the Attribute interface.
var _ Attribute = (*Int)(nil)

func (attr *Int) ID() string                            { return attr.id }
func (attr *Int) Perms() clientsapi.Permissions         { return attr.perms }
func (attr *Int) Optional() OptionalType                { return attr.optional }
func (attr *Int) IsSet() bool                           { return attr.isSet }
func (attr *Int) Push(push func(*clientsapi.Attribute)) { attr.push = push }
func (attr *Int) Pb() *clientsapi.Attribute {
	return &clientsapi.Attribute{
		Id: attr.id,
		Int: &clientsapi.IntAttribute{
			Value: attr.value,
			Min:   attr.min,
			Max:   attr.max,
			Step:  attr.step,
			Unit:  attr.unit,
			Perms: attr.perms,
		},
	}
}
