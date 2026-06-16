package main

import (
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v1/devices"
	"github.com/jimjibone/wh/v1/devices/services"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

type FakeRelay struct {
	dev    *devices.Device
	info   *services.Info
	online *services.Online
	relay  *services.Relay
}

func NewFakeRelay(id, name string) *FakeRelay {
	dev := &FakeRelay{
		dev:    devices.NewDevice(id, clientsapi.Device_DEVICE),
		info:   services.NewInfo(),
		online: services.NewOnline(),
		relay:  services.NewRelay("relay"),
	}
	dev.dev.AddService(dev.info, dev.online, dev.relay)

	// Set up the info service.
	dev.info.Name.Set(name)
	dev.info.Model.Set("Fake Relay Thing")
	dev.info.Manufacturer.Set("Fake Things Inc")
	dev.online.Online.Set(true)
	dev.online.LastSeen.Set(time.Now())

	// Set up the light service.
	dev.relay.OnAction(dev.handleAction)
	dev.relay.On.OnAction(func(val bool) {
		time.Sleep(500 * time.Millisecond)
		log.Infof("on set to %t", val)
		dev.relay.On.Set(val)
	})

	// Set default values.
	dev.relay.On.Set(false)
	dev.relay.Voltage.Set(239.0)
	dev.relay.Current.Set(0.0)
	dev.relay.Power.Set(0.0)
	dev.relay.Temperature.Set(40.0)

	return dev
}

func (dev *FakeRelay) handleAction(request *clientsapi.ActionRequest, feedback func(*clientsapi.ActionResponse)) error {
	feedback(&clientsapi.ActionResponse{
		ActionId: request.ActionId,
		Status:   clientsapi.ActionResponse_SENT,
	})

	for _, req := range request.Values {
		switch req.Id {
		case dev.relay.On.ID():
			if req.GetBool() == nil {
				return services.ErrIncorrectTypeFor(dev.relay.On)
			}
			dev.relay.On.HandleAction(req.GetBool())
		}
	}

	if dev.relay.On.Get() {
		dev.relay.Voltage.Set(240.0)
		dev.relay.Current.Set(0.02)
		dev.relay.Power.Set(5.0)
		dev.relay.Temperature.Set(45.0)
	} else {
		dev.relay.Voltage.Set(239.0)
		dev.relay.Current.Set(0.0)
		dev.relay.Power.Set(0.0)
		dev.relay.Temperature.Set(40.0)
	}

	return nil
}
