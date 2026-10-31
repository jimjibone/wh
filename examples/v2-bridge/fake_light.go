package main

import (
	"errors"
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v2/bridges"
	"github.com/jimjibone/wh/v2/bridges/services"
	"github.com/jimjibone/wh/v2/types"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
)

// FakeLight is a light with the given features on endpoint 1, and identify and
// firmware update services on the root endpoint. It shows feature-gated
// attributes, a narrowed spec (the bulb's own colour temperature range), plain
// writes and the light.set command with a transition.
type FakeLight struct {
	dev      *bridges.Device
	light    *services.Light
	identify *services.Identify
	update   *services.Update
}

func NewFakeLight(uniqueID, name string, features ...string) *FakeLight {
	f := &FakeLight{
		dev:      newDevice(uniqueID, name, "Fake Bulb"),
		light:    services.NewLight("", features...),
		identify: services.NewIdentify(""),
		update:   services.NewUpdate(""),
	}
	f.dev.AddService(f.identify, f.update)
	f.dev.Endpoint("1").AddService(f.light)

	// Each write is handled on its own; the set command applies several at
	// once with an optional transition.
	f.light.On.OnWrite(func(on bool) error {
		f.light.On.Set(on)
		return nil
	})
	if f.light.Brightness != nil {
		f.light.Brightness.OnWrite(func(v int64) error {
			f.light.Brightness.Set(v)
			return nil
		})
	}
	if f.light.ColorTemp != nil {
		// The schema's bounds are wide; this bulb does 2200K to 6500K.
		f.light.ColorTemp.SetRange(2200, 6500)
		f.light.ColorTemp.OnWrite(func(v int64) error {
			f.light.ColorTemp.Set(v)
			f.setMode("color_temp")
			return nil
		})
	}
	if f.light.Color != nil {
		f.light.Color.OnWrite(func(hs *clientsapi.ColorHueSat, xy *clientsapi.ColorXY) error {
			f.light.Color.SetAt(hs, xy, time.Time{})
			f.setMode("color")
			return nil
		})
	}
	f.light.OnSet(func(p types.LightSetParams) error {
		if p.Transition != nil {
			log.Infof("%s: transition over %s", f.dev.ID(), *p.Transition)
		}
		if p.On != nil {
			f.light.On.Set(*p.On)
		}
		if p.Brightness != nil {
			f.light.Brightness.Set(*p.Brightness)
		}
		if p.ColorTemp != nil {
			f.light.ColorTemp.Set(*p.ColorTemp)
			f.setMode("color_temp")
		}
		if p.Color != nil {
			f.light.Color.SetAt(p.Color.GetHueSat(), p.Color.GetXy(), time.Time{})
			f.setMode("color")
		}
		return nil
	})

	f.identify.Identifying.Set(false)
	f.identify.OnIdentify(func(p types.IdentifyIdentifyParams) error {
		d := 5 * time.Second
		if p.Duration != nil {
			d = *p.Duration
		}
		f.identify.Identifying.Set(true)
		time.AfterFunc(d, func() { f.identify.Identifying.Set(false) })
		return nil
	})

	f.update.Available.Set(true)
	f.update.CurrentVersion.Set("1.0.0")
	f.update.UpdateVersion.Set("1.1.0")
	f.update.Updating.Set(false)
	f.update.OnStart(func() error {
		if f.update.Updating.Get() {
			return errors.New("already updating")
		}
		go f.runUpdate()
		return nil
	})

	f.light.On.Set(false)
	if f.light.Brightness != nil {
		f.light.Brightness.Set(100)
	}
	if f.light.ColorTemp != nil {
		f.light.ColorTemp.Set(2700)
	}
	if f.light.Color != nil {
		f.light.Color.SetHueSat(30, 80)
	}
	if f.light.ColorTemp != nil && f.light.Color != nil {
		f.setMode("color_temp")
	}
	return f
}

// setMode follows the mode the bulb applied last, when it has both.
func (f *FakeLight) setMode(mode string) {
	if f.light.ColorTemp != nil && f.light.Color != nil {
		f.light.ColorMode.Set(mode)
	}
}

func (f *FakeLight) runUpdate() {
	f.update.Updating.Set(true)
	for p := int64(0); p <= 100; p += 10 {
		f.update.Progress.Set(p)
		f.update.Remaining.Set(time.Duration(100-p) * 100 * time.Millisecond)
		time.Sleep(time.Second)
	}
	f.update.Updating.Set(false)
	f.update.Available.Set(false)
	f.update.CurrentVersion.Set("1.1.0")
	f.dev.Info.FirmwareVersion.Set("1.1.0")
}
