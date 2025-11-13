package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type InputService struct {
	id        string
	onUpdate  func(changed bool)
	requester requester
	wait      *Waiter
	onliner

	on bool
}

// Initialises the service.
func (srv *InputService) init(serviceID string, requester requester) {
	srv.id = serviceID
	srv.requester = requester
	srv.wait = NewWaiter()
}

// Handle the update. Returns true if the values changed.
func (srv *InputService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "on":
			if srv.on != attr.GetBool().GetValue() {
				changed = true
				srv.on = attr.GetBool().GetValue()
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
func (srv *InputService) OnUpdate(handler func(changed bool)) {
	srv.onUpdate = handler
	srv.onliner.onUpdate = handler
}

// Returns a channel which is closed when the initial state of the service is received.
func (srv *InputService) Ready() <-chan struct{} {
	return srv.wait.Wait()
}

func (srv *InputService) On() bool {
	if srv == nil {
		return false
	}
	return srv.on
}
