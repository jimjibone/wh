package reactors

import (
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type BatteryService struct {
	id       string
	onUpdate func(changed bool)
	wait     *Waiter
	onliner

	level   int64
	voltage *float64
}

// Initialises the service.
func (srv *BatteryService) init(serviceID string, requester requester) {
	srv.id = serviceID
	srv.wait = NewWaiter()
}

// Handle the update. Returns true if the values changed.
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
	if srv.onUpdate != nil {
		srv.onUpdate(changed)
	}
	srv.wait.Done()
	return changed
}

// Sets a handler to be called when the service is updated.
func (srv *BatteryService) OnUpdate(handler func(changed bool)) {
	srv.onUpdate = handler
	srv.onliner.onUpdate = handler
}

// Returns a channel which is closed when the initial state of the service is received.
func (srv *BatteryService) Ready() <-chan struct{} {
	return srv.wait.Wait()
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
