package main

import (
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v1/bridges"
	"github.com/jimjibone/wh/v1/bridges/services"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

type FakeLightbulbColor struct {
	dev       *bridges.Device
	info      *services.Info
	online    *services.Online
	lightbulb *services.Lightbulb
}

func NewFakeLightbulbColor(id, name string) *FakeLightbulbColor {
	dev := &FakeLightbulbColor{
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

	// Support renaming. A real bridge would forward the new name to its
	// backend, return any error, and let the backend's state update drive
	// Name.Set; here the backend is simulated with a short delay.
	dev.info.EnableRename(func(newName string) error {
		time.Sleep(500 * time.Millisecond)
		log.Infof("lightbulb %q renamed to %q", id, newName)
		dev.info.Name.Set(newName)
		return nil
	})

	// Set up the light service.
	dev.lightbulb.On.OnAction(func(val bool) {

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
		dev.lightbulb.ColorMode.Set(services.LightbulbColorModeColorTemp)
	})
	dev.lightbulb.Color.OnAction(func(huesat *clientsapi.ColorHueSat, xy *clientsapi.ColorXY) {
		dev.lightbulb.ColorMode.Set(services.LightbulbColorModeColor)
		if huesat != nil {
			log.Infof("color set to hue %0.0f°, sat %0.0f%%", huesat.Hue, huesat.Sat)
			dev.lightbulb.Color.SetHueSat(huesat.Hue, huesat.Sat)
		} else if xy != nil {
			log.Infof("color set to x %0.3f, y %0.3f", xy.X, xy.Y)
			dev.lightbulb.Color.SetXY(xy.X, xy.Y)
		}
	})

	// Set default values.
	dev.lightbulb.On.Set(false)
	dev.lightbulb.Brightness.Set(75)
	dev.lightbulb.ColorTemp.Set(454)
	dev.lightbulb.Color.Set(32.0, 82.0, 0.0, 0.0)
	dev.lightbulb.ColorMode.SetOptions([]string{services.LightbulbColorModeColorTemp, services.LightbulbColorModeColor})
	dev.lightbulb.ColorMode.Set(services.LightbulbColorModeColorTemp)
	dev.lightbulb.Transition.Set(0)

	return dev
}
