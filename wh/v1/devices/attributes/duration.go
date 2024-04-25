package attributes

import (
	"fmt"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type Duration struct {
	id       string
	perms    clientsapi.Permissions
	optional OptionalType
	min      time.Duration
	max      time.Duration
	step     time.Duration
	value    time.Duration
	isSet    bool
	push     func(*clientsapi.Attribute)
	onAction func(val time.Duration)
}

func NewDuration(id string, perms clientsapi.Permissions, optional OptionalType, min, max, step time.Duration) *Duration {
	if step < 0 {
		step = -step
	}
	return &Duration{
		id:       id,
		perms:    perms,
		optional: optional,
		min:      min,
		max:      max,
		step:     step,
	}
}

func (attr *Duration) Get() time.Duration {
	return attr.value
}

func (attr *Duration) Set(value time.Duration) bool {
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
func (attr *Duration) HandleAction(val *clientsapi.DurationValue) {
	if attr.onAction != nil {
		attr.onAction(time.Duration(val.GetValue()) * time.Millisecond)
	}
}

// OnAction sets the handler which is called when an action is received for this
// attribute. This allows the implementer to forward the request to the end
// device.
func (attr *Duration) OnAction(handler func(val time.Duration)) {
	attr.onAction = handler
}

// Static assert that Duration implements the Attribute interface.
var _ Attribute = (*Duration)(nil)

func (attr *Duration) ID() string                            { return attr.id }
func (attr *Duration) Perms() clientsapi.Permissions         { return attr.perms }
func (attr *Duration) Optional() OptionalType                { return attr.optional }
func (attr *Duration) IsSet() bool                           { return attr.isSet }
func (attr *Duration) Push(push func(*clientsapi.Attribute)) { attr.push = push }
func (attr *Duration) Pb() *clientsapi.Attribute {
	return &clientsapi.Attribute{
		Id: attr.id,
		Duration: &clientsapi.DurationAttribute{
			Value: int64(attr.value / time.Millisecond),
			Min:   int64(attr.min / time.Millisecond),
			Max:   int64(attr.max / time.Millisecond),
			Step:  uint64(attr.step / time.Millisecond),
			Perms: attr.perms,
		},
	}
}
