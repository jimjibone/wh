package services

import (
	"errors"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
)

var (
	ErrAttributeNotFound = errors.New("attribute not found")
	ErrAttributeNotImage = errors.New("attribute does not support image request")
	ErrCommandNotFound   = errors.New("command not found")
	ErrEventNotFound     = errors.New("event not found")
	ErrNoHandler         = errors.New("no handler")
	ErrNoAction          = errors.New("action has neither a write nor a command")
	ErrNotRegistered     = errors.New("service is not registered with a device")
)

// InvalidError marks an action that does not match the service's specs
// (unknown attribute or command, wrong value kind, out of range). The client
// reports it as ERR_INVALID rather than ERROR.
type InvalidError struct{ Err error }

func (e *InvalidError) Error() string { return e.Err.Error() }
func (e *InvalidError) Unwrap() error { return e.Err }

func invalid(err error) error { return &InvalidError{Err: err} }

type Service interface {
	ID() string
	Typ() clientsapi.Service_ServiceType
	// Features returns the enabled feature flags.
	Features() []string
	// HandleAction applies a write or runs a command. It returns the command
	// result, if any.
	HandleAction(request *clientsapi.ActionRequest, feedback func(*clientsapi.ActionResponse)) ([]*clientsapi.FieldValue, error)
	HandleImage(request *clientsapi.ImageRequest, feedback func(*clientsapi.ImageResponse)) ([]byte, error)
	// Push registers the funcs the service calls with partial updates and
	// with events.
	Push(push func(*clientsapi.Service), emit func(*clientsapi.Event))
	// Pb returns the whole service: attributes, features, command and event
	// specs.
	Pb() *clientsapi.Service
}

// ActionHandler handles a whole action request, replacing the default
// handling of writes and commands. See Generic.OnAction.
type ActionHandler func(request *clientsapi.ActionRequest, feedback func(*clientsapi.ActionResponse)) ([]*clientsapi.FieldValue, error)

// CommandHandler runs a command with its validated params and returns its
// result fields (nil when the command declares none).
type CommandHandler func(params []*clientsapi.FieldValue, feedback func(*clientsapi.ActionResponse)) ([]*clientsapi.FieldValue, error)

type ImageHandler func(request *clientsapi.ImageRequest, feedback func(*clientsapi.ImageResponse)) ([]byte, error)
