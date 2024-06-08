package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type InfoService struct {
	name string
}

func (srv *InfoService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "name":
			if srv.name != attr.GetText().GetValue() {
				changed = true
				srv.name = attr.GetText().GetValue()
			}
		}
	}
	return changed
}

func (srv *InfoService) Name() string {
	if srv == nil {
		return ""
	}
	return srv.name
}
