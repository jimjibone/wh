package services

import (
	"errors"
	"fmt"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
	"github.com/jimjibone/woodhouse-4/wh/v1/devices/attributes"
)

var (
	ErrReadOnly          = errors.New("write to read only attribute")
	ErrAttributeNotFound = errors.New("attribute not found")
)

func ErrIncorrectTypeFor(attr attributes.Attribute) error {
	return fmt.Errorf("incorrect type for %q", attr.ID())
}

type Service interface {
	ID() string
	Typ() clientsapi.Service_ServiceType
	HandleAction(request *clientsapi.ActionRequest, feedback func(*clientsapi.ActionResponse)) error
	Push(push func(*clientsapi.Service))
	Pb() *clientsapi.Service
}

type ActionHandler func(request *clientsapi.ActionRequest, feedback func(*clientsapi.ActionResponse)) error
