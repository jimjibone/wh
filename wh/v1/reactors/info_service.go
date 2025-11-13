package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type InfoService struct {
	id        string
	onUpdate  func(changed bool)
	requester requester
	wait      *Waiter
	onliner

	name string
}

// Initialises the service.
func (srv *InfoService) init(serviceID string, requester requester) {
	srv.id = serviceID
	srv.requester = requester
	srv.wait = NewWaiter()
}

// Handle the update. Returns true if the values changed.
func (srv *InfoService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "name":
			if srv.name != attr.GetText().GetValue() {
				changed = true
				srv.name = attr.GetText().GetValue()
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
func (srv *InfoService) OnUpdate(handler func(changed bool)) {
	srv.onUpdate = handler
	srv.onliner.onUpdate = handler
}

// Returns a channel which is closed when the initial state of the service is received.
func (srv *InfoService) Ready() <-chan struct{} {
	return srv.wait.Wait()
}

func (srv *InfoService) Name() string {
	if srv == nil {
		return ""
	}
	return srv.name
}
