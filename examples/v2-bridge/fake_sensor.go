package main

import (
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v2/bridges"
	"github.com/jimjibone/wh/v2/bridges/services"
	"github.com/jimjibone/wh/v2/types"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
)

// FakeClimateSensor is a battery multi-sensor: temperature, humidity and
// pressure on endpoint 1 and the battery on the root endpoint. The sensor
// reports readings with its own sample time, so they are set with SetAt
// rather than stamped by the server on receipt.
type FakeClimateSensor struct {
	dev         *bridges.Device
	temperature *services.Temperature
	humidity    *services.Humidity
	pressure    *services.Pressure
	battery     *services.Battery
}

func NewFakeClimateSensor(uniqueID, name string, sim bool) *FakeClimateSensor {
	f := &FakeClimateSensor{
		dev:         newDevice(uniqueID, name, "Fake Multi-Sensor"),
		temperature: services.NewTemperature(""),
		humidity:    services.NewHumidity(""),
		pressure:    services.NewPressure(""),
		battery:     services.NewBattery(""),
	}
	f.dev.AddService(f.battery)
	f.dev.Endpoint("1").AddService(f.temperature, f.humidity, f.pressure)
	f.battery.Level.Set(92)
	f.battery.Voltage.Set(3.0)
	f.sample(time.Now())
	if sim {
		go func() {
			for {
				time.Sleep(30 * time.Second)
				// A battery sensor reports a little after it samples.
				f.sample(time.Now().Add(-2 * time.Second))
			}
		}()
	}
	return f
}

func (f *FakeClimateSensor) sample(at time.Time) {
	hour := float64(at.Hour()) + float64(at.Minute())/60
	daily := math.Sin((hour - 9) / 24 * 2 * math.Pi)
	f.temperature.Temperature.SetAt(math.Round((19+3*daily+rand.Float64()*0.2)*10)/10, at)
	f.humidity.Humidity.SetAt(math.Round(55-10*daily+rand.Float64()), at)
	f.pressure.Pressure.SetAt(math.Round(1013+rand.Float64()*4-2), at)
}

// FakePresence is an mmWave presence sensor that also senses motion: presence
// and motion on one endpoint. Motion emits a detected event each time it
// starts, besides setting the motion attribute.
type FakePresence struct {
	dev      *bridges.Device
	presence *services.Presence
	motion   *services.Motion
}

func NewFakePresence(uniqueID, name string, sim bool) *FakePresence {
	f := &FakePresence{
		dev:      newDevice(uniqueID, name, "Fake Presence Sensor"),
		presence: services.NewPresence("", services.PresenceFeatureDistance),
		motion:   services.NewMotion(""),
	}
	f.dev.Endpoint("1").AddService(f.presence, f.motion)
	f.presence.Presence.Set(false)
	f.presence.Distance.Set(0)
	f.motion.Motion.Set(false)
	if sim {
		go f.simulate()
	}
	return f
}

func (f *FakePresence) setMotion(on bool) {
	if f.motion.Motion.Set(on) && on {
		if err := f.motion.EmitDetected(); err != nil {
			log.Errorf("%s: %s", f.dev.ID(), err)
		}
	}
}

// simulate walks someone into the room, has them sit still, then leave.
func (f *FakePresence) simulate() {
	for {
		for dist := 6.0; dist >= 1.0; dist -= 0.5 {
			f.presence.Distance.Set(dist)
			f.presence.Presence.Set(true)
			f.setMotion(true)
			time.Sleep(time.Second)
		}
		f.setMotion(false)
		time.Sleep(10 * time.Second) // still, but present
		for dist := 1.0; dist <= 6.0; dist += 0.5 {
			f.presence.Distance.Set(dist)
			f.setMotion(true)
			time.Sleep(time.Second)
		}
		f.presence.Presence.Set(false)
		f.presence.Distance.Set(0)
		f.setMotion(false)
		time.Sleep(20 * time.Second)
	}
}

// FakeTankLevel is an analogue level sensor (a water tank) with the range_map
// feature: the raw 0-100% reading is mapped onto litres.
type FakeTankLevel struct {
	dev    *bridges.Device
	analog *services.AnalogInput
}

func NewFakeTankLevel(uniqueID, name string, sim bool) *FakeTankLevel {
	f := &FakeTankLevel{
		dev:    newDevice(uniqueID, name, "Fake Level Sensor"),
		analog: services.NewAnalogInput("", services.AnalogInputFeatureRangeMap),
	}
	f.dev.Endpoint("1").AddService(f.analog)
	f.analog.SetRangeMap(types.AnalogInputRangeMap{Min: 0, Max: 1000})
	f.analog.RangeMap.OnWrite(func(fields []*clientsapi.FieldValue) error {
		r := types.ParseAnalogInputRangeMap(fields)
		if r.Max <= r.Min {
			return errors.New("max must be above min")
		}
		f.analog.SetRangeMap(r)
		f.read(f.analog.Value.Get())
		return nil
	})
	f.read(72)
	if sim {
		go func() {
			for {
				time.Sleep(20 * time.Second)
				f.read(max(0, f.analog.Value.Get()-0.5))
			}
		}()
	}
	return f
}

func (f *FakeTankLevel) read(percent float64) {
	r := f.analog.RangeMapValue()
	f.analog.Value.Set(percent)
	f.analog.Mapped.Set(r.Min + (r.Max-r.Min)*percent/100)
}
