package attributes

import (
	"fmt"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type Image struct {
	id             string
	push           func(*clientsapi.Attribute)
	onImageRequest func() ([]byte, error)
}

func NewImage(id string) *Image {
	return &Image{
		id: id,
	}
}

// HandleImage calls the attribute's OnImageRequest handler if set.
func (attr *Image) HandleImage() ([]byte, error) {
	if attr.onImageRequest != nil {
		return attr.onImageRequest()
	}
	return nil, fmt.Errorf("no handler for image request")
}

// OnImageRequest sets the handler which is called when an action is received for this
// attribute. This allows the implementer to forward the request to the end
// device.
func (attr *Image) OnImageRequest(handler func() ([]byte, error)) {
	attr.onImageRequest = handler
}

// Static assert that Image implements the Attribute interface.
var _ Attribute = (*Image)(nil)

func (attr *Image) ID() string                            { return attr.id }
func (attr *Image) Perms() clientsapi.Permissions         { return clientsapi.Permissions_PERM_WRITEONLY }
func (attr *Image) Optional() OptionalType                { return Required }
func (attr *Image) IsSet() bool                           { return true }
func (attr *Image) Push(push func(*clientsapi.Attribute)) { attr.push = push }
func (attr *Image) Pb() *clientsapi.Attribute {
	return &clientsapi.Attribute{
		Id:    attr.id,
		Image: &clientsapi.ImageAttribute{},
	}
}
