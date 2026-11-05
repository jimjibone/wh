package services

import (
	"fmt"
	"slices"
	"time"

	"github.com/jimjibone/wh/v2/bridges/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
)

// Static assert that Generic implements the Service interface.
var _ Service = (*Generic)(nil)

// Generic is the base of every service: a set of attributes, commands and
// events. The typed services (Light, OnOff, ...) embed it. Used alone it is the
// "generic" service type, which may hold anything.
type Generic struct {
	id       string
	typ      clientsapi.Service_ServiceType
	alias    string
	features []string

	attrs    []attributes.Attribute // in the order added, so updates are stable
	commands []*command
	events   []*clientsapi.EventSpec

	push     func(*clientsapi.Service)
	emit     func(*clientsapi.Event)
	onAction ActionHandler
	onImage  ImageHandler
}

type command struct {
	spec    *clientsapi.CommandSpec
	handler CommandHandler
}

// NewGeneric creates a generic service. The service ID must be unique within
// the endpoint and is normally the service name in lowercase (e.g. "generic").
func NewGeneric(id string) *Generic {
	return newGeneric(id, clientsapi.Service_GENERIC, nil)
}

func newGeneric(id string, typ clientsapi.Service_ServiceType, features []string) *Generic {
	if id == "" {
		id = DefaultServiceID(typ)
	}
	return &Generic{id: id, typ: typ, features: slices.Clone(features)}
}

// SetAlias sets the alias for this service. Sometimes used to differentiate
// between multiple services of the same type on a single endpoint.
func (srv *Generic) SetAlias(alias string) {
	srv.alias = alias
}

// HasFeature reports whether a feature is enabled.
func (srv *Generic) HasFeature(feature string) bool {
	return slices.Contains(srv.features, feature)
}

// AddAttribute adds attributes to the service. Adding an ID twice replaces
// the earlier attribute.
func (srv *Generic) AddAttribute(attrs ...attributes.Attribute) {
	for _, attr := range attrs {
		attr.Push(srv.pusher)
		srv.attrs = slices.DeleteFunc(srv.attrs, func(a attributes.Attribute) bool { return a.ID() == attr.ID() })
		srv.attrs = append(srv.attrs, attr)
	}
}

// Attributes returns the service's attributes in the order added.
func (srv *Generic) Attributes() []attributes.Attribute {
	return slices.Clone(srv.attrs)
}

// Attribute returns the attribute with the ID, or nil.
func (srv *Generic) Attribute(id string) attributes.Attribute {
	for _, attr := range srv.attrs {
		if attr.ID() == id {
			return attr
		}
	}
	return nil
}

// AddCommand declares a command. The handler may be nil and set later with
// OnCommand. Adding an ID twice replaces the earlier command.
func (srv *Generic) AddCommand(spec *clientsapi.CommandSpec, handler CommandHandler) {
	srv.commands = slices.DeleteFunc(srv.commands, func(c *command) bool { return c.spec.GetId() == spec.GetId() })
	srv.commands = append(srv.commands, &command{spec: spec, handler: handler})
}

// CommandSpec returns the spec of a command, or nil. Modify it (e.g. to set
// the enum options of a param) before the service is registered with a
// device.
func (srv *Generic) CommandSpec(id string) *clientsapi.CommandSpec {
	if c := srv.command(id); c != nil {
		return c.spec
	}
	return nil
}

// OnCommand sets the handler for a declared command. It panics if the command
// was not declared, since that is a programming error.
func (srv *Generic) OnCommand(id string, handler CommandHandler) {
	c := srv.command(id)
	if c == nil {
		panic(fmt.Sprintf("service %q has no command %q", srv.id, id))
	}
	c.handler = handler
}

func (srv *Generic) command(id string) *command {
	for _, c := range srv.commands {
		if c.spec.GetId() == id {
			return c
		}
	}
	return nil
}

// AddEvent declares an event. Adding an ID twice replaces the earlier event.
func (srv *Generic) AddEvent(spec *clientsapi.EventSpec) {
	srv.events = slices.DeleteFunc(srv.events, func(e *clientsapi.EventSpec) bool { return e.GetId() == spec.GetId() })
	srv.events = append(srv.events, spec)
}

// Emit sends an event seen now. The fields are validated against the event's
// spec.
func (srv *Generic) Emit(id string, fields ...*clientsapi.FieldValue) error {
	return srv.EmitAt(time.Now(), id, fields...)
}

// EmitAt sends an event the device saw at t.
func (srv *Generic) EmitAt(t time.Time, id string, fields ...*clientsapi.FieldValue) error {
	i := slices.IndexFunc(srv.events, func(e *clientsapi.EventSpec) bool { return e.GetId() == id })
	if i < 0 {
		return fmt.Errorf("%w: %q", ErrEventNotFound, id)
	}
	if err := values.ValidateFields(srv.events[i].GetFields(), fields); err != nil {
		return fmt.Errorf("event %q: %w", id, err)
	}
	if srv.emit == nil {
		return fmt.Errorf("event %q: %w", id, ErrNotRegistered)
	}
	srv.emit(&clientsapi.Event{
		ServiceId: srv.id,
		Id:        id,
		Fields:    slices.Clone(fields),
		Time:      values.TimeToPb(t),
	})
	return nil
}

// OnAction overrides the default action handling (attribute writes and
// commands), e.g. to report progress such as SENT. The handler can send
// feedback at any time and returns when finished: a nil error completes the
// action with the returned result, an error fails it with the error message.
func (srv *Generic) OnAction(handler ActionHandler) {
	srv.onAction = handler
}

// OnImage overrides the default image request handler, which serves the
// requested image attribute. See OnAction for feedback and the result.
func (srv *Generic) OnImage(handler ImageHandler) {
	srv.onImage = handler
}

func (srv *Generic) ID() string                          { return srv.id }
func (srv *Generic) Typ() clientsapi.Service_ServiceType { return srv.typ }
func (srv *Generic) Features() []string                  { return slices.Clone(srv.features) }

func (srv *Generic) HandleAction(request *clientsapi.ActionRequest, feedback func(*clientsapi.ActionResponse)) ([]*clientsapi.FieldValue, error) {
	if srv.onAction != nil {
		return srv.onAction(request, feedback)
	}
	switch action := request.GetAction().(type) {
	case *clientsapi.ActionRequest_Write:
		return nil, srv.handleWrite(action.Write.GetValues())
	case *clientsapi.ActionRequest_Command:
		return srv.handleCommand(action.Command, feedback)
	}
	return nil, invalid(ErrNoAction)
}

// handleWrite checks every value before writing any, so a bad value in a
// multi-attribute write changes nothing.
func (srv *Generic) handleWrite(writes []*clientsapi.FieldValue) error {
	for _, w := range writes {
		attr := srv.Attribute(w.GetId())
		if attr == nil {
			return invalid(fmt.Errorf("%w: %q", ErrAttributeNotFound, w.GetId()))
		}
		if attr.Perms() != clientsapi.Permissions_PERM_READWRITE {
			return attributes.ErrReadOnly
		}
		if err := values.Validate(attr.Pb().GetSpec(), w.GetValue()); err != nil {
			return invalid(fmt.Errorf("%s: %w", w.GetId(), err))
		}
	}
	for _, w := range writes {
		if err := srv.Attribute(w.GetId()).HandleWrite(w.GetValue()); err != nil {
			return err
		}
	}
	return nil
}

func (srv *Generic) handleCommand(cmd *clientsapi.CommandAction, feedback func(*clientsapi.ActionResponse)) ([]*clientsapi.FieldValue, error) {
	c := srv.command(cmd.GetId())
	if c == nil {
		return nil, invalid(fmt.Errorf("%w: %q", ErrCommandNotFound, cmd.GetId()))
	}
	if err := values.ValidateFields(c.spec.GetParams(), cmd.GetParams()); err != nil {
		return nil, invalid(fmt.Errorf("command %q: %w", cmd.GetId(), err))
	}
	if c.handler == nil {
		return nil, fmt.Errorf("command %q: %w", cmd.GetId(), ErrNoHandler)
	}
	return c.handler(cmd.GetParams(), feedback)
}

func (srv *Generic) HandleImage(request *clientsapi.ImageRequest, feedback func(*clientsapi.ImageResponse)) ([]byte, error) {
	if srv.onImage != nil {
		return srv.onImage(request, feedback)
	}
	img, ok := srv.Attribute(request.GetAttributeId()).(*attributes.Image)
	if !ok {
		if srv.Attribute(request.GetAttributeId()) == nil {
			return nil, ErrAttributeNotFound
		}
		return nil, ErrAttributeNotImage
	}
	return img.HandleImage()
}

func (srv *Generic) Push(push func(*clientsapi.Service), emit func(*clientsapi.Event)) {
	srv.push = push
	srv.emit = emit
}

// pusher forwards attribute updates once the service is registered with a
// device. Before that they are dropped: the attribute keeps its state and the
// full state sent on registration carries it.
func (srv *Generic) pusher(attr *clientsapi.Attribute) {
	if srv.push == nil {
		return
	}
	srv.push(&clientsapi.Service{
		Id:    srv.id,
		Typ:   srv.typ,
		Alias: srv.alias,
		Attrs: []*clientsapi.Attribute{attr},
	})
}

func (srv *Generic) Pb() *clientsapi.Service {
	pb := &clientsapi.Service{
		Id:       srv.id,
		Typ:      srv.typ,
		Alias:    srv.alias,
		Attrs:    []*clientsapi.Attribute{},
		Features: slices.Clone(srv.features),
		Events:   slices.Clone(srv.events),
	}
	for _, attr := range srv.attrs {
		if attr.Optional() == attributes.Required || attr.IsSet() {
			pb.Attrs = append(pb.Attrs, attr.Pb())
		}
	}
	for _, c := range srv.commands {
		pb.Commands = append(pb.Commands, c.spec)
	}
	return pb
}
