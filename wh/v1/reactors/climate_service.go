package reactors

import (
	"context"
	"fmt"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type ClimateService struct {
	id        string
	onUpdate  func(changed bool)
	requester requester
	wait      *Waiter
	onliner

	heatingSetpoint  float64
	localTemperature float64
	piHeatingDemand  *int64
}

// Initialises the service.
func (srv *ClimateService) init(serviceID string, requester requester) {
	srv.id = serviceID
	srv.requester = requester
	srv.wait = NewWaiter()
}

// Handle the update. Returns true if the values changed.
func (srv *ClimateService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	srv.id = update.GetId()
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
	if srv.onUpdate != nil {
		srv.onUpdate(changed)
	}
	srv.wait.Done()
	return changed
}

// Sets a handler to be called when the service is updated.
func (srv *ClimateService) OnUpdate(handler func(changed bool)) {
	srv.onUpdate = handler
	srv.onliner.onUpdate = handler
}

// Returns a channel which is closed when the initial state of the service is received.
func (srv *ClimateService) Ready() <-chan struct{} {
	return srv.wait.Wait()
}

func (srv *ClimateService) HeatingSetpoint() float64 {
	if srv == nil {
		return 0
	}
	return srv.heatingSetpoint
}

func (srv *ClimateService) SetHeatingSetpoint(ctx context.Context, val float64) error {
	if srv == nil {
		return fmt.Errorf("service not initialised")
	}
	return srv.requester(
		ctx,
		&clientsapi.ActionRequest{
			ServiceId: srv.id,
			Values: []*clientsapi.Value{
				{
					Id: "heating_setpoint",
					Float: &clientsapi.FloatValue{
						Value: val,
					},
				},
			},
		},
		func(resp *clientsapi.ActionResponse) {
		},
	)
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
