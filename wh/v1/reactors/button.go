package reactors

import (
	"time"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type ButtonService struct {
	state    string
	options  []string
	duration *time.Duration
}

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
	return changed
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
