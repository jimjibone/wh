package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type InputService struct {
	on bool
}

func (srv *InputService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "on":
			if srv.on != attr.GetBool().GetValue() {
				changed = true
				srv.on = attr.GetBool().GetValue()
			}
		}
	}
	return changed
}

func (srv *InputService) On() bool {
	if srv == nil {
		return false
	}
	return srv.on
}
