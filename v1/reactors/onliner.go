package reactors

import (
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

// onliner is a simple embeddable struct which can handle online service updates for use in other services.
// Note that the handleOnline is very similar to OnlineService's handleUpdate function.
type onliner struct {
	name     string
	online   bool
	lastSeen time.Time
	onUpdate func(changed bool)
}

// Returns the name of the device for this service.
func (srv *onliner) DeviceName() string {
	return srv.name
}

// Returns true if the device for this service is online.
func (srv *onliner) Online() bool {
	return srv.online
}

// Returns the last seen time of the device for this service.
func (srv *onliner) LastSeen() time.Time {
	return srv.lastSeen
}

// Handle the info update.
func (srv *onliner) handleInfo(update *clientsapi.Service) bool {
	changed := false
	if update.Typ == clientsapi.Service_INFO {
		for _, attr := range update.Attrs {
			switch attr.GetId() {
			case "name":
				if srv.name != attr.GetText().GetValue() {
					// changed = true
					srv.name = attr.GetText().GetValue()
				}
			}
		}
		if srv.onUpdate != nil {
			srv.onUpdate(changed)
		}
	}
	return changed
}

// Handle the online update.
func (srv *onliner) handleOnline(update *clientsapi.Service) bool {
	changed := false
	if update.Typ == clientsapi.Service_ONLINE {
		for _, attr := range update.Attrs {
			switch attr.GetId() {
			case "online":
				if srv.online != attr.GetBool().GetValue() {
					// changed = true
					srv.online = attr.GetBool().GetValue()
				}

			case "last_seen":
				t := timeFromPb(attr.GetTime())
				if !srv.lastSeen.Equal(t) {
					//changed = true
					srv.lastSeen = t
				}
			}
		}
		if srv.onUpdate != nil {
			srv.onUpdate(changed)
		}
	}
	return changed
}
