package reactors

import (
	"context"
	"fmt"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type RelayService struct {
	id        string
	onUpdate  func(changed bool)
	requester requester
	exists    bool
	onliner

	on          bool
	voltage     *float64
	current     *float64
	power       *float64
	temperature *float64
}

// Initialises the service.
func (srv *RelayService) init(serviceID string, requester requester) {
	srv.id = serviceID
	srv.requester = requester
}

// Handle the update. Returns true if the values changed.
func (srv *RelayService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	srv.id = update.GetId()
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "on":
			if srv.on != attr.GetBool().GetValue() {
				changed = true
				srv.on = attr.GetBool().GetValue()
			}

		case "voltage":
			if srv.voltage == nil {
				srv.voltage = new(float64)
			}
			if *srv.voltage != attr.GetFloat().GetValue() {
				changed = true
				*srv.voltage = attr.GetFloat().GetValue()
			}

		case "current":
			if srv.current == nil {
				srv.current = new(float64)
			}
			if *srv.current != attr.GetFloat().GetValue() {
				changed = true
				*srv.current = attr.GetFloat().GetValue()
			}

		case "power":
			if srv.power == nil {
				srv.power = new(float64)
			}
			if *srv.power != attr.GetFloat().GetValue() {
				changed = true
				*srv.power = attr.GetFloat().GetValue()
			}

		case "temperature":
			if srv.temperature == nil {
				srv.temperature = new(float64)
			}
			if *srv.temperature != attr.GetFloat().GetValue() {
				changed = true
				*srv.temperature = attr.GetFloat().GetValue()
			}
		}
	}
	if srv.onUpdate != nil {
		srv.onUpdate(changed)
	}
	srv.exists = true
	return changed
}

// Sets a handler to be called when the service is updated.
func (srv *RelayService) OnUpdate(handler func(changed bool)) {
	srv.onUpdate = handler
	srv.onliner.onUpdate = handler
}

// Returns whether the service exists or not. May be false until the client
// receives the initial state from the server.
func (srv *RelayService) Exists() bool {
	return srv.exists
}

func (srv *RelayService) On() bool {
	if srv == nil {
		return false
	}
	return srv.on
}

func (srv *RelayService) SetOn(ctx context.Context, on bool) error {
	if srv == nil {
		return fmt.Errorf("service not initialised")
	}
	return srv.requester(
		ctx,
		&clientsapi.ActionRequest{
			ServiceId: srv.id,
			Values: []*clientsapi.Value{
				{
					Id: "on",
					Bool: &clientsapi.BoolValue{
						Value: on,
					},
				},
			},
		},
		func(resp *clientsapi.ActionResponse) {
		},
	)
}

func (srv *RelayService) HasVoltage() bool {
	if srv == nil || srv.voltage == nil {
		return false
	}
	return true
}

func (srv *RelayService) Voltage() float64 {
	if srv == nil || srv.voltage == nil {
		return 0.0
	}
	return *srv.voltage
}

func (srv *RelayService) HasCurrent() bool {
	if srv == nil || srv.current == nil {
		return false
	}
	return true
}

func (srv *RelayService) Current() float64 {
	if srv == nil || srv.current == nil {
		return 0.0
	}
	return *srv.current
}

func (srv *RelayService) HasPower() bool {
	if srv == nil || srv.power == nil {
		return false
	}
	return true
}

func (srv *RelayService) Power() float64 {
	if srv == nil || srv.power == nil {
		return 0.0
	}
	return *srv.power
}

func (srv *RelayService) HasTemperature() bool {
	if srv == nil || srv.temperature == nil {
		return false
	}
	return true
}

func (srv *RelayService) Temperature() float64 {
	if srv == nil || srv.temperature == nil {
		return 0.0
	}
	return *srv.temperature
}
