package reactors

import (
	"fmt"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type ButtonService struct {
	id       string
	onUpdate func(changed bool)
	exists   bool
	onliner

	state    string
	options  []string
	duration *time.Duration
}

// Initialises the service.
func (srv *ButtonService) init(serviceID string, requester requester) {
	srv.id = serviceID
}

// Handle the update. Returns true if the values changed.
func (srv *ButtonService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "state":
			if srv.state != attr.GetEnum().GetValue() {
				changed = true
				srv.state = attr.GetEnum().GetValue()
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

		case "duration":
			if srv.duration == nil {
				srv.duration = new(time.Duration)
			}
			val := time.Duration(attr.GetDuration().GetValue()) * time.Millisecond
			if *srv.duration != val {
				changed = true
				*srv.duration = val
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
func (srv *ButtonService) OnUpdate(handler func(changed bool)) {
	srv.onUpdate = handler
	srv.onliner.onUpdate = handler
}

// Returns whether the service exists or not. May be false until the client
// receives the initial state from the server.
func (srv *ButtonService) Exists() bool {
	return srv.exists
}

func (srv *ButtonService) State() string {
	if srv == nil {
		return ""
	}
	return srv.state
}

func (srv *ButtonService) Options() []string {
	if srv == nil {
		return nil
	}
	return srv.options
}

func (srv *ButtonService) HasDuration() bool {
	if srv == nil || srv.duration == nil {
		return false
	}
	return true
}

func (srv *ButtonService) Duration() time.Duration {
	if srv == nil || srv.duration == nil {
		return 0
	}
	return *srv.duration
}

func (srv *ButtonService) String() string {
	return fmt.Sprintf("{state:%q, duration:%s}", srv.state, srv.duration)
}

func (srv *ButtonService) StringLong() string {
	return fmt.Sprintf("{state:%q, duration:%s, options:%q}", srv.state, srv.duration, srv.options)
}
