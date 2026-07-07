package main

import (
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v1/bridges"
	"github.com/jimjibone/wh/v1/bridges/services"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

type FakeLightbulbColorTemp struct {
	dev       *bridges.Device
	info      *services.Info
	online    *services.Online
	lightbulb *services.Lightbulb
}

func NewFakeLightbulbColorTemp(id, name string) *FakeLightbulbColorTemp {
	dev := &FakeLightbulbColorTemp{
		dev:       bridges.NewDevice(id, clientsapi.Device_DEVICE),
		info:      services.NewInfo(),
		online:    services.NewOnline(),
		lightbulb: services.NewLightbulb("lightbulb"),
	}
	dev.dev.AddService(dev.info, dev.online, dev.lightbulb)

	// Set up the info service.
	dev.info.Name.Set(name)
	dev.info.Model.Set("Fake Lightbulb Thing")
	dev.info.Manufacturer.Set("Fake Things Inc")
	dev.online.Online.Set(true)
	dev.online.LastSeen.Set(time.Now())

	// Set up the light service.
	dev.lightbulb.On.OnAction(func(val bool) {
		time.Sleep(2 * time.Second)
		log.Infof("on set to %t", val)
		dev.lightbulb.On.Set(val)
	})
	dev.lightbulb.Brightness.OnAction(func(val int64) {
		log.Infof("brightness set to %d%%", val)
		dev.lightbulb.Brightness.Set(val)
	})
	dev.lightbulb.ColorTemp.OnAction(func(val int64) {
		log.Infof("color temperature set to %d", val)
		dev.lightbulb.ColorTemp.Set(val)
	})

	// Set default values.
	dev.lightbulb.On.Set(false)
	dev.lightbulb.Brightness.Set(75)
	dev.lightbulb.ColorTemp.Set(454)
	dev.lightbulb.ColorTemp.SetLimits(153, 555, 1)
	dev.lightbulb.Transition.Set(0)

	return dev
}
