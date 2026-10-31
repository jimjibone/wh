package main

import (
	"time"

	"github.com/jimjibone/wh/v2/bridges"
	"github.com/jimjibone/wh/v2/bridges/services"
)

// FakePlug is a smart plug: on_off and power_meter on endpoint 1, and the
// plug's own temperature on the root endpoint. Switching is slow, as with a
// device on a radio network, so it reports SENT progress through OnAction.
type FakePlug struct {
	dev   *bridges.Device
	out   *services.OnOff
	meter *services.PowerMeter
	temp  *services.DeviceTemperature
}

func NewFakePlug(uniqueID, name string, watts float64) *FakePlug {
	f := &FakePlug{
		dev:   newDevice(uniqueID, name, "Fake Plug"),
		out:   services.NewOnOff(""),
		meter: services.NewPowerMeter(""),
		temp:  services.NewDeviceTemperature(""),
	}
	f.dev.AddService(f.temp)
	f.dev.Endpoint("1").AddService(f.out, f.meter)

	f.out.On.OnWrite(func(on bool) error {
		time.Sleep(500 * time.Millisecond)
		f.set(on, watts)
		return nil
	})
	f.out.OnToggle(func() error {
		f.set(!f.out.On.Get(), watts)
		return nil
	})

	f.meter.Energy.Set(12.5)
	f.set(false, watts)
	return f
}

func (f *FakePlug) set(on bool, watts float64) {
	f.out.On.Set(on)
	if on {
		f.meter.Voltage.Set(240)
		f.meter.Power.Set(watts)
		f.meter.Current.Set(watts / 240)
		f.temp.Temperature.Set(45)
	} else {
		f.meter.Voltage.Set(239)
		f.meter.Power.Set(0)
		f.meter.Current.Set(0)
		f.temp.Temperature.Set(40)
	}
}
