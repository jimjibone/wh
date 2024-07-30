package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type EnvironmentService struct {
	temperature *float64
	humidity    *float64
	pressure    *float64
}

func (srv *EnvironmentService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "temperature":
			if srv.temperature == nil {
				srv.temperature = new(float64)
			}
			if *srv.temperature != attr.GetFloat().GetValue() {
				changed = true
				*srv.temperature = attr.GetFloat().GetValue()
			}

		case "humidity":
			if srv.humidity == nil {
				srv.humidity = new(float64)
			}
			if *srv.humidity != attr.GetFloat().GetValue() {
				changed = true
				*srv.humidity = attr.GetFloat().GetValue()
			}

		case "pressure":
			if srv.pressure == nil {
				srv.pressure = new(float64)
			}
			if *srv.pressure != attr.GetFloat().GetValue() {
				changed = true
				*srv.pressure = attr.GetFloat().GetValue()
			}
		}
	}
	return changed
}

func (srv *EnvironmentService) HasTemperature() bool {
	if srv == nil || srv.temperature == nil {
		return false
	}
	return true
}

func (srv *EnvironmentService) Temperature() float64 {
	if srv == nil || srv.temperature == nil {
		return 0
	}
	return *srv.temperature
}

func (srv *EnvironmentService) HasHumidity() bool {
	if srv == nil || srv.humidity == nil {
		return false
	}
	return true
}

func (srv *EnvironmentService) Humidity() float64 {
	if srv == nil || srv.humidity == nil {
		return 0
	}
	return *srv.humidity
}

func (srv *EnvironmentService) HasPressure() bool {
	if srv == nil || srv.pressure == nil {
		return false
	}
	return true
}

func (srv *EnvironmentService) Pressure() float64 {
	if srv == nil || srv.pressure == nil {
		return 0
	}
	return *srv.pressure
}
