package reactors

import (
	"context"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

type Service interface {
	// Initialises the service.
	init(serviceID string, requester requester)

	// Handle the update. Returns true if the values changed.
	handleUpdate(update *clientsapi.Service) bool

	// Handle the info update.
	handleInfo(update *clientsapi.Service) bool

	// Handle the online update.
	handleOnline(update *clientsapi.Service) bool

	// Sets a handler to be called when the service is updated.
	OnUpdate(handler func(changed bool))

	// Returns whether the service exists or not. May be false until the client
	// receives the initial state from the server.
	Exists() bool

	// Returns the name of the device for this service.
	DeviceName() string

	// Returns true if the device for this service is online.
	Online() bool

	// Returns the last seen time of the device for this service.
	LastSeen() time.Time
}

type requester func(ctx context.Context, req *clientsapi.ActionRequest, handler func(resp *clientsapi.ActionResponse)) error
