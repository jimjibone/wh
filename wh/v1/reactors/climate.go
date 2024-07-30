package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type ClimateService struct {
	heatingSetpoint  float64
	localTemperature float64
	piHeatingDemand  *int64
}

func (srv *ClimateService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "heating_setpoint":
			if srv.heatingSetpoint != attr.GetFloat().GetValue() {
				changed = true
				srv.heatingSetpoint = attr.GetFloat().GetValue()
			}

		case "local_temperature":
			if srv.localTemperature != attr.GetFloat().GetValue() {
				changed = true
				srv.localTemperature = attr.GetFloat().GetValue()
			}

		case "pi_heating_demand":
			if srv.piHeatingDemand == nil {
				srv.piHeatingDemand = new(int64)
			}
			if *srv.piHeatingDemand != attr.GetInt().GetValue() {
				changed = true
				*srv.piHeatingDemand = attr.GetInt().GetValue()
			}
		}
	}
	return changed
}

func (srv *ClimateService) HeatingSetpoint() float64 {
	if srv == nil {
		return 0
	}
	return srv.heatingSetpoint
}

func (srv *ClimateService) LocalTemperature() float64 {
	if srv == nil {
		return 0
	}
	return srv.localTemperature
}

func (srv *ClimateService) HasPiHeatingDemand() bool {
	if srv == nil || srv.piHeatingDemand == nil {
		return false
	}
	return true
}

func (srv *ClimateService) PiHeatingDemand() int64 {
	if srv == nil || srv.piHeatingDemand == nil {
		return 0
	}
	return *srv.piHeatingDemand
}
