package attributes

import (
	"fmt"
	"math"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type Float struct {
	id       string
	perms    clientsapi.Permissions
	optional OptionalType
	min      float64
	max      float64
	step     float64
	unit     clientsapi.Unit
	value    float64
	isSet    bool
	push     func(*clientsapi.Attribute)
	onAction func(val float64)
}

func NewFloat(id string, perms clientsapi.Permissions, optional OptionalType, min, max, step float64, unit clientsapi.Unit) *Float {
	return &Float{
		id:       id,
		perms:    perms,
		optional: optional,
		min:      min,
		max:      max,
		step:     step,
		unit:     unit,
	}
}

func (attr *Float) SetOptional(optional OptionalType) {
	attr.optional = optional
}

func (attr *Float) Get() float64 {
	return attr.value
}

func (attr *Float) GetLimits() (min, max, step float64) {
	return attr.min, attr.max, attr.step
}

func (attr *Float) Set(value float64) bool {
	attr.isSet = true
	if math.Abs(attr.value-value) > 0.0001 {
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

func (attr *Float) SetLimits(min, max, step float64) {
	attr.min = min
	attr.max = max
	attr.step = step
	if attr.push != nil {
		attr.push(attr.Pb())
	} else {
		panic(fmt.Sprintf("attribute %q is not registered with a service", attr.id))
	}
}

// HandleAction calls the attribute's OnAction handler if set.
func (attr *Float) HandleAction(val *clientsapi.FloatValue) {
	if attr.onAction != nil {
		attr.onAction(val.GetValue())
	}
}

// OnAction sets the handler which is called when an action is received for this
// attribute. This allows the implementer to forward the request to the end
// device.
func (attr *Float) OnAction(handler func(val float64)) {
	attr.onAction = handler
}

// Static assert that Float implements the Attribute interface.
var _ Attribute = (*Float)(nil)

func (attr *Float) ID() string                            { return attr.id }
func (attr *Float) Perms() clientsapi.Permissions         { return attr.perms }
func (attr *Float) Optional() OptionalType                { return attr.optional }
func (attr *Float) IsSet() bool                           { return attr.isSet }
func (attr *Float) Push(push func(*clientsapi.Attribute)) { attr.push = push }
func (attr *Float) Pb() *clientsapi.Attribute {
	return &clientsapi.Attribute{
		Id: attr.id,
		Float: &clientsapi.FloatAttribute{
			Value: attr.value,
			Min:   attr.min,
			Max:   attr.max,
			Step:  attr.step,
			Unit:  attr.unit,
			Perms: attr.perms,
		},
	}
}
