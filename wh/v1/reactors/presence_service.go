package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type PresenceService struct {
	id       string
	onUpdate func(changed bool)
	exists   bool
	onliner

	motion   bool
	presence bool
	distance float64
}

// Initialises the service.
func (srv *PresenceService) init(serviceID string, requester requester) {
	srv.id = serviceID
}

// Handle the update. Returns true if the values changed.
func (srv *PresenceService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "motion":
			if srv.motion != attr.GetBool().GetValue() {
				changed = true
				srv.motion = attr.GetBool().GetValue()
			}
		case "presence":
			if srv.presence != attr.GetBool().GetValue() {
				changed = true
				srv.presence = attr.GetBool().GetValue()
			}
		case "distance":
			if srv.distance != attr.GetFloat().GetValue() {
				changed = true
				srv.distance = attr.GetFloat().GetValue()
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
func (srv *PresenceService) OnUpdate(handler func(changed bool)) {
	srv.onUpdate = handler
	srv.onliner.onUpdate = handler
}

// Returns whether the service exists or not. May be false until the client
// receives the initial state from the server.
func (srv *PresenceService) Exists() bool {
	return srv.exists
}

func (srv *PresenceService) Motion() bool {
	if srv == nil {
		return false
	}
	return srv.motion
}

func (srv *PresenceService) Presence() bool {
	if srv == nil {
		return false
	}
	return srv.presence
}

func (srv *PresenceService) Distance() float64 {
	if srv == nil {
		return 0.0
	}
	return srv.distance
}
