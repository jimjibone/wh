package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type EnumService struct {
	value   string
	options []string
}

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
	return changed
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
