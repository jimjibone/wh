package main

import (
	"time"

	"github.com/jimjibone/wh/v2/bridges"
	"github.com/jimjibone/wh/v2/bridges/services"
	"github.com/jimjibone/wh/v2/types"
)

// FakeThermostat is a radiator thermostat with a heating schedule (a list of
// structs, replaced with the set_schedule command) and a mode service.
type FakeThermostat struct {
	dev        *bridges.Device
	thermostat *services.Thermostat
	mode       *services.Mode
	battery    *services.Battery
}

func NewFakeThermostat(uniqueID, name string, sim bool) *FakeThermostat {
	f := &FakeThermostat{
		dev:        newDevice(uniqueID, name, "Fake Radiator Valve"),
		thermostat: services.NewThermostat("", services.ThermostatFeatureSchedule),
		mode:       services.NewMode(""),
		battery:    services.NewBattery(""),
	}
	f.dev.AddService(f.battery)
	f.dev.Endpoint("1").AddService(f.thermostat, f.mode)

	f.thermostat.HeatingSetpoint.OnWrite(func(v float64) error {
		f.thermostat.HeatingSetpoint.Set(v)
		f.update()
		return nil
	})
	f.thermostat.ValvePosition.OnWrite(func(v int64) error {
		f.thermostat.ValvePosition.Set(v)
		return nil
	})
	f.thermostat.OnSetSchedule(func(p types.ThermostatSetScheduleParams) error {
		f.thermostat.SetSchedule(p.Schedule)
		return nil
	})
	f.mode.Value.SetOptions([]string{"auto", "manual", "off"})
	f.mode.Value.OnWrite(func(v string) error {
		f.mode.Value.Set(v)
		return nil
	})

	weekdays := []string{"mon", "tue", "wed", "thu", "fri"}
	f.thermostat.SetSchedule([]types.ThermostatScheduleItem{
		{Days: weekdays, Start: 6*time.Hour + 30*time.Minute, Setpoint: 20},
		{Days: weekdays, Start: 22 * time.Hour, Setpoint: 16},
		{Days: []string{"sat", "sun"}, Start: 8 * time.Hour, Setpoint: 20.5},
		{Days: []string{"sat", "sun"}, Start: 23 * time.Hour, Setpoint: 16},
	})
	f.mode.Value.Set("auto")
	f.battery.Level.Set(64)
	f.thermostat.HeatingSetpoint.Set(20)
	f.thermostat.LocalTemperature.Set(18.5)
	f.update()

	if sim {
		go f.simulate()
	}
	return f
}

// update derives the heating demand from the setpoint and the temperature.
func (f *FakeThermostat) update() {
	diff := f.thermostat.HeatingSetpoint.Get() - f.thermostat.LocalTemperature.Get()
	demand := int64(min(100, max(0, diff*40)))
	f.thermostat.PiHeatingDemand.Set(demand)
	f.thermostat.HeatingDemand.Set(demand > 0)
	f.thermostat.ValvePosition.Set(demand)
}

// simulate warms the room towards the setpoint when heating, cools it
// otherwise.
func (f *FakeThermostat) simulate() {
	for {
		time.Sleep(10 * time.Second)
		t := f.thermostat.LocalTemperature.Get()
		if f.thermostat.HeatingDemand.Get() {
			t += 0.1
		} else {
			t -= 0.05
		}
		f.thermostat.LocalTemperature.Set(float64(int(t*100)) / 100)
		f.update()
	}
}
