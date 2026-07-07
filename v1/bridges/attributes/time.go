package attributes

import (
	"fmt"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

type Time struct {
	id       string
	perms    clientsapi.Permissions
	optional OptionalType
	value    time.Time
	isSet    bool
	push     func(*clientsapi.Attribute)
	onAction func(val time.Time)
}

func timeFromPb(t *clientsapi.TimeValue) time.Time {
	if t == nil {
		return time.Time{}
	}
	if t.Seconds == 0 && t.Nanos == 0 {
		return time.Time{}
	}
	return time.Unix(t.Seconds, int64(t.Nanos))
}

func timeToPb(t time.Time, p clientsapi.Permissions) *clientsapi.TimeAttribute {
	secs, nanos := t.Unix(), int32(t.Nanosecond())
	if t.IsZero() {
		secs, nanos = 0, 0
	}
	return &clientsapi.TimeAttribute{
		Seconds: secs,
		Nanos:   nanos,
		Perms:   p,
	}
}

func NewTime(id string, perms clientsapi.Permissions, optional OptionalType) *Time {
	return &Time{
		id:       id,
		perms:    perms,
		optional: optional,
	}
}

func (attr *Time) SetOptional(optional OptionalType) {
	attr.optional = optional
}

func (attr *Time) Get() time.Time {
	return attr.value
}

func (attr *Time) Set(value time.Time) bool {
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
func (attr *Time) HandleAction(val *clientsapi.TimeValue) {
	if attr.onAction != nil {
		attr.onAction(timeFromPb(val))
	}
}

// OnAction sets the handler which is called when an action is received for this
// attribute. This allows the implementer to forward the request to the end
// device.
func (attr *Time) OnAction(handler func(val time.Time)) {
	attr.onAction = handler
}

// Static assert that Time implements the Attribute interface.
var _ Attribute = (*Time)(nil)

func (attr *Time) ID() string                            { return attr.id }
func (attr *Time) Perms() clientsapi.Permissions         { return attr.perms }
func (attr *Time) Optional() OptionalType                { return attr.optional }
func (attr *Time) IsSet() bool                           { return attr.isSet }
func (attr *Time) Push(push func(*clientsapi.Attribute)) { attr.push = push }
func (attr *Time) Pb() *clientsapi.Attribute {
	return &clientsapi.Attribute{
		Id:   attr.id,
		Time: timeToPb(attr.value, attr.perms),
	}
}
