package reactors

import (
	"time"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type OnlineService struct {
	online   bool
	lastSeen time.Time
}

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
	return changed
}

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
