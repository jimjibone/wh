package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type UpdateService struct {
	available      bool
	currentVersion *string
	updateVersion  *string
}

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
	return changed
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
