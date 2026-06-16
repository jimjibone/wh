package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

type EnumService struct {
	id        string
	onUpdate  func(changed bool)
	requester requester
	exists    bool
	onliner

	value   string
	options []string
}

// Initialises the service.
func (srv *EnumService) init(serviceID string, requester requester) {
	srv.id = serviceID
	srv.requester = requester
}

// Handle the update. Returns true if the values changed.
func (srv *EnumService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "value":
			if srv.value != attr.GetEnum().GetValue() {
				changed = true
				srv.value = attr.GetEnum().GetValue()
			}
			if len(srv.options) != len(attr.GetEnum().GetOptions()) {
				changed = true
			} else {
				for i := range srv.options {
					if srv.options[i] != attr.GetEnum().GetOptions()[i] {
						changed = true
						break
					}
				}
			}
			srv.options = attr.GetEnum().GetOptions()
		}
	}
	if srv.onUpdate != nil {
		srv.onUpdate(changed)
	}
	srv.exists = true
	return changed
}

// Sets a handler to be called when the service is updated.
func (srv *EnumService) OnUpdate(handler func(changed bool)) {
	srv.onUpdate = handler
	srv.onliner.onUpdate = handler
}

// Returns whether the service exists or not. May be false until the client
// receives the initial state from the server.
func (srv *EnumService) Exists() bool {
	return srv.exists
}

func (srv *EnumService) Value() string {
	if srv == nil {
		return ""
	}
	return srv.value
}

func (srv *EnumService) Options() []string {
	if srv == nil {
		return nil
	}
	return srv.options
}
