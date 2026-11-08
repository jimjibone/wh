package attributes

import (
	"errors"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
)

var ErrNoImageHandler = errors.New("no handler for image request")

// Image is a read only attribute whose data is fetched on request (see
// ImageRequest) rather than sent with updates.
type Image struct {
	base
	onImageRequest func() ([]byte, error)
}

// Static assert that Image implements the Attribute interface.
var _ Attribute = (*Image)(nil)

func NewImage(id string, optional OptionalType) *Image {
	attr := &Image{base: newBase(id, clientsapi.Permissions_PERM_READONLY, optional, &clientsapi.Spec{Kind: &clientsapi.Spec_Image{Image: &clientsapi.ImageSpec{}}})}
	attr.value = values.Image()
	return attr
}

// HandleImage calls the attribute's OnImageRequest handler.
func (attr *Image) HandleImage() ([]byte, error) {
	attr.mu.Lock()
	handler := attr.onImageRequest
	attr.mu.Unlock()
	if handler == nil {
		return nil, ErrNoImageHandler
	}
	return handler()
}

// OnImageRequest sets the handler that fetches the image from the device.
func (attr *Image) OnImageRequest(handler func() ([]byte, error)) {
	attr.mu.Lock()
	defer attr.mu.Unlock()
	attr.onImageRequest = handler
}
