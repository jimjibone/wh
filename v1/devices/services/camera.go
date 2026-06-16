package services

import (
	"github.com/jimjibone/wh/v1/devices/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

// Static assert that Camera implements the Service interface.
var _ Service = (*Camera)(nil)

type Camera struct {
	*Generic
	Image *attributes.Image
}

func init() {
	registerDefaultServiceID(clientsapi.Service_CAMERA, "camera")
}

// New Camera service. The service ID must be unique within the device and is
// normally the service name in lowercase (e.g. "camera").
func NewCamera(id string) *Camera {
	if id == "" {
		id = DefaultServiceID(clientsapi.Service_CAMERA)
	}
	srv := &Camera{
		Generic: newGeneric(id, clientsapi.Service_CAMERA),
		Image:   attributes.NewImage("image"),
	}
	srv.AddAttribute(
		srv.Image,
	)
	return srv
}
