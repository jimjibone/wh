package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

type UpdateService struct {
	id        string
	onUpdate  func(changed bool)
	requester requester
	exists    bool
	onliner

	available      bool
	currentVersion *string
	updateVersion  *string
}

// Initialises the service.
func (srv *UpdateService) init(serviceID string, requester requester) {
	srv.id = serviceID
	srv.requester = requester
}

// Handle the update. Returns true if the values changed.
func (srv *UpdateService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "available":
			if srv.available != attr.GetBool().GetValue() {
				changed = true
				srv.available = attr.GetBool().GetValue()
			}

		case "current_version":
			if srv.currentVersion == nil {
				srv.currentVersion = new(string)
			}
			if *srv.currentVersion != attr.GetText().GetValue() {
				changed = true
				*srv.currentVersion = attr.GetText().GetValue()
			}

		case "update_version":
			if srv.updateVersion == nil {
				srv.updateVersion = new(string)
			}
			if *srv.updateVersion != attr.GetText().GetValue() {
				changed = true
				*srv.updateVersion = attr.GetText().GetValue()
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
func (srv *UpdateService) OnUpdate(handler func(changed bool)) {
	srv.onUpdate = handler
	srv.onliner.onUpdate = handler
}

// Returns whether the service exists or not. May be false until the client
// receives the initial state from the server.
func (srv *UpdateService) Exists() bool {
	return srv.exists
}

func (srv *UpdateService) Available() bool {
	if srv == nil {
		return false
	}
	return srv.available
}

func (srv *UpdateService) HasCurrentVersion() bool {
	if srv == nil || srv.currentVersion == nil {
		return false
	}
	return true
}

func (srv *UpdateService) CurrentVersion() string {
	if srv == nil || srv.currentVersion == nil {
		return ""
	}
	return *srv.currentVersion
}

func (srv *UpdateService) HasUpdateVersion() bool {
	if srv == nil || srv.updateVersion == nil {
		return false
	}
	return true
}

func (srv *UpdateService) UpdateVersion() string {
	if srv == nil || srv.updateVersion == nil {
		return ""
	}
	return *srv.updateVersion
}
