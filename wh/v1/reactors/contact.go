package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type ContactService struct {
	closed bool
}

func (srv *ContactService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "closed":
			if srv.closed != attr.GetBool().GetValue() {
				changed = true
				srv.closed = attr.GetBool().GetValue()
			}
		}
	}
	return changed
}

func (srv *ContactService) Closed() bool {
	if srv == nil {
		return false
	}
	return srv.closed
}
