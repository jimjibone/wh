package reactors

import (
	"github.com/jimjibone/log"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

type CameraService struct {
	id       string
	onUpdate func(changed bool)
	exists   bool
	onliner
}

// Initialises the service.
func (srv *CameraService) init(serviceID string, requester requester) {
	srv.id = serviceID
}

// Handle the update. Returns true if the values changed.
func (srv *CameraService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	log.Infof("CameraService.handleUpdate: %s", update)
	// for _, attr := range update.Attrs {
	// 	switch attr.GetId() {
	// 	case "heating_setpoint":
	// 		if srv.heatingSetpoint != attr.GetFloat().GetValue() {
	// 			changed = true
	// 			srv.heatingSetpoint = attr.GetFloat().GetValue()
	// 		}

	// 	case "local_temperature":
	// 		if srv.localTemperature != attr.GetFloat().GetValue() {
	// 			changed = true
	// 			srv.localTemperature = attr.GetFloat().GetValue()
	// 		}

	// 	case "pi_heating_demand":
	// 		if srv.piHeatingDemand == nil {
	// 			srv.piHeatingDemand = new(int64)
	// 		}
	// 		if *srv.piHeatingDemand != attr.GetInt().GetValue() {
	// 			changed = true
	// 			*srv.piHeatingDemand = attr.GetInt().GetValue()
	// 		}
	// 	}
	// }
	if srv.onUpdate != nil {
		srv.onUpdate(changed)
	}
	srv.exists = true
	return changed
}

// Sets a handler to be called when the service is updated.
func (srv *CameraService) OnUpdate(handler func(changed bool)) {
	srv.onUpdate = handler
	srv.onliner.onUpdate = handler
}

// Returns whether the service exists or not. May be false until the client
// receives the initial state from the server.
func (srv *CameraService) Exists() bool {
	return srv.exists
}
