package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type MotionService struct {
	id       string
	onUpdate func(changed bool)
	exists   bool
	onliner

	motion bool
}

// Initialises the service.
func (srv *MotionService) init(serviceID string, requester requester) {
	srv.id = serviceID
}

// Handle the update. Returns true if the values changed.
func (srv *MotionService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "motion":
			if srv.motion != attr.GetBool().GetValue() {
				changed = true
				srv.motion = attr.GetBool().GetValue()
			}
		}
	}
	if srv.onUpdate != nil {
		srv.onUpdate(changed)
	}
	srv.exists = true
	return changed
}

// Sets a handler to be called when the service is updated.
func (srv *MotionService) OnUpdate(handler func(changed bool)) {
	srv.onUpdate = handler
	srv.onliner.onUpdate = handler
}

// Returns whether the service exists or not. May be false until the client
// receives the initial state from the server.
func (srv *MotionService) Exists() bool {
	return srv.exists
}

func (srv *MotionService) Motion() bool {
	if srv == nil {
		return false
	}
	return srv.motion
}
