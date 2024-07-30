package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
	"github.com/jimjibone/woodhouse-4/log"
)

type CameraService struct {
}

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
	return changed
}
