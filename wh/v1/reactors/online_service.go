package reactors

import (
	"time"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type OnlineService struct {
	id        string
	onUpdate  func(changed bool)
	requester requester
	wait      *Waiter
	onliner

	online   bool
	lastSeen time.Time
}

// Initialises the service.
func (srv *OnlineService) init(serviceID string, requester requester) {
	srv.id = serviceID
	srv.requester = requester
	srv.wait = NewWaiter()
}

// Handle the update. Returns true if the values changed.
func (srv *OnlineService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "online":
			if srv.online != attr.GetBool().GetValue() {
				changed = true
				srv.online = attr.GetBool().GetValue()
			}

		case "last_seen":
			t := timeFromPb(attr.GetTime())
			if !srv.lastSeen.Equal(t) {
				changed = true
				srv.lastSeen = t
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
func (srv *OnlineService) OnUpdate(handler func(changed bool)) {
	srv.onUpdate = handler
	srv.onliner.onUpdate = handler
}

// Returns a channel which is closed when the initial state of the service is received.
func (srv *OnlineService) Ready() <-chan struct{} {
	return srv.wait.Wait()
}

// Returns true if the device for this service is online.
func (srv *OnlineService) Online() bool {
	if srv == nil {
		return false
	}
	return srv.online
}

func (srv *OnlineService) LastSeen() time.Time {
	if srv == nil {
		return time.Time{}
	}
	return srv.lastSeen
}
