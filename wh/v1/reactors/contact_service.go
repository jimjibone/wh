package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type ContactService struct {
	id       string
	onUpdate func(changed bool)
	wait     *Waiter
	onliner

	closed bool
}

// Initialises the service.
func (srv *ContactService) init(serviceID string, requester requester) {
	srv.id = serviceID
	srv.wait = NewWaiter()
}

// Handle the update. Returns true if the values changed.
func (srv *ContactService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "closed":
			if srv.closed != attr.GetBool().GetValue() {
				changed = true
				srv.closed = attr.GetBool().GetValue()
			}
		}
	}
	if srv.onUpdate != nil {
		srv.onUpdate(changed)
	}
	srv.wait.Done()
	return changed
}

// Sets a handler to be called when the service is updated.
func (srv *ContactService) OnUpdate(handler func(changed bool)) {
	srv.onUpdate = handler
	srv.onliner.onUpdate = handler
}

// Returns a channel which is closed when the initial state of the service is received.
func (srv *ContactService) Ready() <-chan struct{} {
	return srv.wait.Wait()
}

func (srv *ContactService) Closed() bool {
	if srv == nil {
		return false
	}
	return srv.closed
}
