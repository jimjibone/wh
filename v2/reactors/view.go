package reactors

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
	"google.golang.org/protobuf/proto"
)

// ActionError is a failed action: its final status and the bridge's details.
type ActionError struct {
	Status  clientsapi.ActionResponse_ActionStatus
	Details string
}

func (e *ActionError) Error() string {
	if e.Details == "" {
		return e.Status.String()
	}
	return e.Status.String() + ": " + e.Details
}

// ErrNoFinalStatus is returned when an action's response stream ends without
// a final status.
var ErrNoFinalStatus = errors.New("action ended without a final status")

// ServiceView is a reactor's view of one service of a device: its mirrored
// state, writes, commands and events. The typed views (LightView, ...) embed
// it and add typed accessors. It is valid before the service is known; it
// then reads as zero values and Exists is false.
type ServiceView struct {
	rc  *Reactor
	key serviceKey
	typ clientsapi.Service_ServiceType

	onUpdate func(changed bool)
	events   []*eventHandler
}

func (v *ServiceView) DeviceID() string   { return v.key.device }
func (v *ServiceView) EndpointID() string { return v.key.endpoint }
func (v *ServiceView) ServiceID() string  { return v.key.service }

// pb returns the mirrored service, or nil if absent or of another type (a
// view with type UNDEFINED takes any type). The caller holds rc.mu.
func (v *ServiceView) pb() *clientsapi.Service {
	srv := v.rc.devices[v.key.device].service(v.key.endpoint, v.key.service)
	if srv == nil || (v.typ != clientsapi.Service_UNDEFINED && srv.GetTyp() != v.typ) {
		return nil
	}
	return srv
}

// Exists reports whether the device reports this service with the expected
// type. It may be false until the reactor has received the initial state.
func (v *ServiceView) Exists() bool {
	v.rc.mu.RLock()
	defer v.rc.mu.RUnlock()
	return v.pb() != nil
}

// Features returns the service's enabled feature flags.
func (v *ServiceView) Features() []string {
	v.rc.mu.RLock()
	defer v.rc.mu.RUnlock()
	return slices.Clone(v.pb().GetFeatures())
}

// HasFeature reports whether a feature is enabled.
func (v *ServiceView) HasFeature(feature string) bool {
	v.rc.mu.RLock()
	defer v.rc.mu.RUnlock()
	return slices.Contains(v.pb().GetFeatures(), feature)
}

// Attribute returns a copy of an attribute (spec, value and time), or nil.
func (v *ServiceView) Attribute(id string) *clientsapi.Attribute {
	v.rc.mu.RLock()
	defer v.rc.mu.RUnlock()
	a := attribute(v.pb(), id)
	if a == nil {
		return nil
	}
	return proto.Clone(a).(*clientsapi.Attribute)
}

// Has reports whether an attribute has a value.
func (v *ServiceView) Has(id string) bool {
	return values.IsSet(v.value(id))
}

// value returns an attribute's value, or nil.
func (v *ServiceView) value(id string) *clientsapi.Value {
	v.rc.mu.RLock()
	defer v.rc.mu.RUnlock()
	return attribute(v.pb(), id).GetValue()
}

// AttributeTime returns when an attribute's value was observed.
func (v *ServiceView) AttributeTime(id string) time.Time {
	v.rc.mu.RLock()
	defer v.rc.mu.RUnlock()
	return values.TimeFromPb(attribute(v.pb(), id).GetTime())
}

// CommandSpec returns a copy of a command's spec, or nil, e.g. to read the
// options of an enum param.
func (v *ServiceView) CommandSpec(id string) *clientsapi.CommandSpec {
	v.rc.mu.RLock()
	defer v.rc.mu.RUnlock()
	for _, c := range v.pb().GetCommands() {
		if c.GetId() == id {
			return proto.Clone(c).(*clientsapi.CommandSpec)
		}
	}
	return nil
}

// DeviceName returns the name from the device's info service.
func (v *ServiceView) DeviceName() string {
	return v.rc.Device(v.key.device).Name()
}

// Online reports the device's online state from its online service.
func (v *ServiceView) Online() bool {
	return v.rc.Device(v.key.device).Online()
}

// LastSeen returns when the device was last seen, from its online service.
func (v *ServiceView) LastSeen() time.Time {
	return v.rc.Device(v.key.device).LastSeen()
}

// OnUpdate sets the handler called after every update of the service.
// changed reports whether a value, spec, feature or the service's presence
// changed. It replaces an earlier handler.
func (v *ServiceView) OnUpdate(handler func(changed bool)) {
	v.rc.mu.Lock()
	defer v.rc.mu.Unlock()
	v.onUpdate = handler
}

// OnEvent adds a handler for an event of the service, or every event when id
// is "". Replayed events older than DefaultMaxEventAge are dropped unless
// MaxAge says otherwise.
func (v *ServiceView) OnEvent(id string, handler func(e *clientsapi.Event, info EventInfo), opts ...EventOption) {
	h := &eventHandler{id: id, fn: handler, maxAge: DefaultMaxEventAge}
	for _, opt := range opts {
		opt(h)
	}
	v.rc.mu.Lock()
	defer v.rc.mu.Unlock()
	v.events = append(v.events, h)
}

// Write writes attribute values and waits for the action to finish.
func (v *ServiceView) Write(ctx context.Context, fields ...*clientsapi.FieldValue) error {
	_, err := v.Do(ctx, &clientsapi.ActionRequest{
		Action: &clientsapi.ActionRequest_Write{Write: &clientsapi.WriteAction{Values: fields}},
	}, nil)
	return err
}

// Command runs a command and waits for its result.
func (v *ServiceView) Command(ctx context.Context, id string, params ...*clientsapi.FieldValue) ([]*clientsapi.FieldValue, error) {
	return v.Do(ctx, &clientsapi.ActionRequest{
		Action: &clientsapi.ActionRequest_Command{Command: &clientsapi.CommandAction{Id: id, Params: params}},
	}, nil)
}

// Do sends an action to the service and waits for its final status. progress,
// if not nil, sees every response. A final status other than COMPLETE is
// returned as an *ActionError.
func (v *ServiceView) Do(ctx context.Context, req *clientsapi.ActionRequest, progress func(*clientsapi.ActionResponse)) ([]*clientsapi.FieldValue, error) {
	req.DeviceId = v.key.device
	req.EndpointId = v.key.endpoint
	req.ServiceId = v.key.service
	var final *clientsapi.ActionResponse
	err := v.rc.requestAction(ctx, req, func(res *clientsapi.ActionResponse) {
		if progress != nil {
			progress(res)
		}
		if res.GetStatus() >= clientsapi.ActionResponse_COMPLETE {
			final = res
		}
	})
	if err != nil {
		return nil, err
	}
	if final == nil {
		return nil, ErrNoFinalStatus
	}
	if final.GetStatus() != clientsapi.ActionResponse_COMPLETE {
		return nil, &ActionError{Status: final.GetStatus(), Details: final.GetDetails()}
	}
	return final.GetResult(), nil
}

func (v *ServiceView) String() string {
	return fmt.Sprintf("%s %s/%s/%s", v.typ, v.key.device, v.key.endpoint, v.key.service)
}
