package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type BatteryService struct {
	level   int64
	voltage *float64
}

func (srv *BatteryService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "level":
			if srv.level != attr.GetInt().GetValue() {
				changed = true
				srv.level = attr.GetInt().GetValue()
			}

		case "voltage":
			if srv.voltage == nil {
				srv.voltage = new(float64)
			}
			if *srv.voltage != attr.GetFloat().GetValue() {
				changed = true
				*srv.voltage = attr.GetFloat().GetValue()
			}
		}
	}
	return changed
}

func (srv *BatteryService) Level() int64 {
	if srv == nil {
		return 0
	}
	return srv.level
}

func (srv *BatteryService) HasVoltage() bool {
	if srv == nil || srv.voltage == nil {
		return false
	}
	return true
}

func (srv *BatteryService) Voltage() float64 {
	if srv == nil || srv.voltage == nil {
		return 0
	}
	return *srv.voltage
}
