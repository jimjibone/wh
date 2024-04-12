package services

import (
	"fmt"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
	"github.com/jimjibone/woodhouse-4/log"
	"github.com/jimjibone/woodhouse-4/wh/v1/devices/attributes"
)

type Generic struct {
	id       string
	typ      clientsapi.Service_ServiceType
	alias    string
	attrs    map[string]attributes.Attribute
	push     func(*clientsapi.Service)
	onAction ActionHandler
}

var _ Service = (*Generic)(nil)

func NewGeneric() *Generic {
	return newGeneric("generic", clientsapi.Service_GENERIC)
}

func newGeneric(id string, typ clientsapi.Service_ServiceType) *Generic {
	srv := &Generic{
		id:    id,
		typ:   typ,
		attrs: make(map[string]attributes.Attribute),
	}
	srv.onAction = func(request *clientsapi.ActionRequest, feedback func(*clientsapi.ActionResponse)) error {
		for _, req := range request.Values {
			attr, found := srv.attrs[req.GetId()]
			if !found {
				return ErrAttributeNotFound
			}

			switch attr.Perms() {
			case clientsapi.Permissions_PERM_WRITEONLY, clientsapi.Permissions_PERM_READWRITE:
				switch attr := attr.(type) {
				case *attributes.Bool:
					if req.GetBool() == nil {
						return ErrIncorrectTypeFor(attr)
					}
					attr.HandleAction(req.GetBool())

				case *attributes.Duration:
					if req.GetDuration() == nil {
						return ErrIncorrectTypeFor(attr)
					}
					attr.HandleAction(req.GetDuration())

				case *attributes.Float:
					if req.GetFloat() == nil {
						return ErrIncorrectTypeFor(attr)
					}
					attr.HandleAction(req.GetFloat())

				case *attributes.Int:
					if req.GetInt() == nil {
						return ErrIncorrectTypeFor(attr)
					}
					attr.HandleAction(req.GetInt())

				case *attributes.Text:
					if req.GetText() == nil {
						return ErrIncorrectTypeFor(attr)
					}
					attr.HandleAction(req.GetText())

				case *attributes.Time:
					if req.GetTime() == nil {
						return ErrIncorrectTypeFor(attr)
					}
					attr.HandleAction(req.GetTime())

				default:
					log.Fatalf("unknown attribute type: %T", attr)
				}

			case clientsapi.Permissions_PERM_UNDEFINED, clientsapi.Permissions_PERM_READONLY:
				return ErrReadOnly

			default:
				log.Fatalf("unknown perms value: %v", attr.Perms())
			}
		}
		return nil
	}
	return srv
}

// SetAlias sets the alias for this service. Sometimes used to differentiate
// between multiple services of the same type on a single device.
func (srv *Generic) SetAlias(alias string) {
	srv.alias = alias
}

// AddAttribute adds the attributes to the service.
func (srv *Generic) AddAttribute(attrs ...attributes.Attribute) {
	for _, attr := range attrs {
		attr.Push(srv.pusher)
		srv.attrs[attr.ID()] = attr
	}
}

// OnAction overrides the default action request handler. This is useful when
// the implementer wants to provide additional progress feedback for requests,
// such as SENT, TIMEOUT, etc.
// The handler must parse the request and pass the action values to the service
// attributes. The handler can send feedback at any time using the feedback func
// and should eventually return when finished. When the handler returns a final
// ActionResponse will be sent back. If the returned error is nil the response
// status will be COMPLETE, otherwise ERR with the details field containing the
// error message.
func (srv *Generic) OnAction(handler ActionHandler) {
	srv.onAction = handler
}

// Static assert that Generic implements the Service interface.
var _ Service = (*Generic)(nil)

func (srv *Generic) ID() string {
	return srv.id
}

func (srv *Generic) Action(request *clientsapi.ActionRequest, feedback func(*clientsapi.ActionResponse)) error {
	if srv.onAction != nil {
		return srv.onAction(request, feedback)
	}
	return fmt.Errorf("no action handler")
}

func (srv *Generic) Push(push func(*clientsapi.Service)) {
	srv.push = push
}

func (srv *Generic) pusher(attr *clientsapi.Attribute) {
	if srv.push != nil {
		srv.push(&clientsapi.Service{
			Id:    srv.id,
			Typ:   clientsapi.Service_INFO,
			Alias: srv.alias,
			Attrs: []*clientsapi.Attribute{attr},
		})
	} else {
		panic(fmt.Sprintf("service %q is not registered with a device", srv.id))
	}
}

func (srv *Generic) Pb() *clientsapi.Service {
	pb := &clientsapi.Service{
		Id:    srv.id,
		Typ:   srv.typ,
		Alias: srv.alias,
		Attrs: []*clientsapi.Attribute{},
	}
	for _, attr := range srv.attrs {
		if attr.Optional() == attributes.Required || attr.IsSet() {
			pb.Attrs = append(pb.Attrs, attr.Pb())
		}
	}
	return pb
}
