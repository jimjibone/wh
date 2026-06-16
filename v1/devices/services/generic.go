package services

import (
	"fmt"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v1/devices/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

// Static assert that Generic implements the Service interface.
var _ Service = (*Generic)(nil)

type Generic struct {
	id       string
	typ      clientsapi.Service_ServiceType
	alias    string
	attrs    map[string]attributes.Attribute
	push     func(*clientsapi.Service)
	onAction ActionHandler
	onImage  ImageHandler
}

func init() {
	registerDefaultServiceID(clientsapi.Service_GENERIC, "generic")
}

// New Generic service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "generic").
func NewGeneric(id string) *Generic {
	return newGeneric(id, clientsapi.Service_GENERIC)
}

func newGeneric(id string, typ clientsapi.Service_ServiceType) *Generic {
	if id == "" {
		id = DefaultServiceID(typ)
	}
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

				case *attributes.Image:
					if req.GetImage() == nil {
						return ErrIncorrectTypeFor(attr)
					}
					// attr.HandleAction(req.GetImage())
					return fmt.Errorf("actions not supported for image")

				case *attributes.Color:
					if req.GetColor() == nil {
						return ErrIncorrectTypeFor(attr)
					}
					attr.HandleAction(req.GetColor())

				case *attributes.Duration:
					if req.GetDuration() == nil {
						return ErrIncorrectTypeFor(attr)
					}
					attr.HandleAction(req.GetDuration())

				case *attributes.Enum:
					if req.GetEnum() == nil {
						return ErrIncorrectTypeFor(attr)
					}
					attr.HandleAction(req.GetEnum())

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
	srv.onImage = func(req *clientsapi.ImageRequest, feedback func(*clientsapi.ImageResponse)) ([]byte, error) {
		attr, found := srv.attrs[req.GetAttributeId()]
		if !found {
			return nil, ErrAttributeNotFound
		}

		switch attr := attr.(type) {
		case *attributes.Image:
			data, err := attr.HandleImage()
			if err != nil {
				return nil, err
			}
			return data, nil

		default:
			return nil, ErrAttributeNotImage
		}
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

// OnImage overrides the default image request handler. This is useful when
// the implementer wants to provide additional progress feedback for requests,
// such as SENT, TIMEOUT, etc.
// The handler must parse the request and pass the request values to the service
// attributes. The handler can send feedback at any time using the feedback func
// and should eventually return when finished. When the handler returns a final
// ImageResponse will be sent back. If the returned error is nil the response
// status will be COMPLETE and the data field will be set with the returned byte
// slice, otherwise ERR with the details field containing the error message.
func (srv *Generic) OnImage(handler ImageHandler) {
	srv.onImage = handler
}

// Static assert that Generic implements the Service interface.
var _ Service = (*Generic)(nil)

func (srv *Generic) ID() string {
	return srv.id
}

func (srv *Generic) Typ() clientsapi.Service_ServiceType {
	return srv.typ
}

func (srv *Generic) HandleAction(request *clientsapi.ActionRequest, feedback func(*clientsapi.ActionResponse)) error {
	if srv.onAction != nil {
		return srv.onAction(request, feedback)
	}
	return fmt.Errorf("no action handler")
}

func (srv *Generic) HandleImage(request *clientsapi.ImageRequest, feedback func(*clientsapi.ImageResponse)) ([]byte, error) {
	if srv.onImage != nil {
		return srv.onImage(request, feedback)
	}
	return nil, fmt.Errorf("no image handler")
}

func (srv *Generic) Push(push func(*clientsapi.Service)) {
	srv.push = push
}

func (srv *Generic) pusher(attr *clientsapi.Attribute) {
	if srv.push != nil {
		srv.push(&clientsapi.Service{
			Id:    srv.id,
			Typ:   srv.typ,
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
