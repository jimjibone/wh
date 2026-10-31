package main

import (
	"math/rand"
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v2/bridges"
	"github.com/jimjibone/wh/v2/bridges/services"
	"github.com/jimjibone/wh/v2/types"
)

// FakeRemote is a battery remote with two buttons, one per endpoint ("1" On,
// "2" Off). Every press is an event; last_event holds the latest one, and its
// options list the events this remote emits.
type FakeRemote struct {
	dev     *bridges.Device
	battery *services.Battery
	buttons [2]*services.Button
}

func NewFakeRemote(uniqueID, name string, sim bool) *FakeRemote {
	f := &FakeRemote{
		dev:     newDevice(uniqueID, name, "Fake Remote"),
		battery: services.NewBattery(""),
	}
	f.dev.AddService(f.battery)
	f.battery.Level.Set(87)
	for i, label := range []string{"On", "Off"} {
		ep := f.dev.Endpoint([]string{"1", "2"}[i])
		ep.SetLabel(label)
		f.buttons[i] = services.NewButton("")
		ep.AddService(f.buttons[i])
		f.buttons[i].LastEvent.SetOptions([]string{"press", "release", "hold", "long_release", "multi_press"})
		f.buttons[i].LastEvent.Set("release")
	}
	if sim {
		go f.simulate()
	}
	return f
}

// simulate presses buttons: single presses, a double press and a long press.
func (f *FakeRemote) simulate() {
	for {
		time.Sleep(time.Duration(15+rand.Intn(15)) * time.Second)
		b := f.buttons[rand.Intn(2)]
		var err error
		switch rand.Intn(3) {
		case 0: // single press
			err = f.emit(b, "press", func() error { return b.EmitPress() })
			if err == nil {
				err = f.emit(b, "release", func() error { return b.EmitRelease() })
			}
			if err == nil {
				err = f.emit(b, "multi_press", func() error { return b.EmitMultiPress(types.ButtonMultiPressEvent{Count: 1}) })
			}
		case 1: // double press
			err = f.emit(b, "multi_press", func() error { return b.EmitMultiPress(types.ButtonMultiPressEvent{Count: 2}) })
		case 2: // long press
			err = f.emit(b, "hold", func() error { return b.EmitHold() })
			time.Sleep(2 * time.Second)
			if err == nil {
				b.LastDuration.Set(2 * time.Second)
				err = f.emit(b, "long_release", func() error { return b.EmitLongRelease(types.ButtonLongReleaseEvent{Duration: 2 * time.Second}) })
			}
		}
		if err != nil {
			log.Errorf("%s: %s", f.dev.ID(), err)
		}
	}
}

// emit sends an event and records it as the last event.
func (f *FakeRemote) emit(b *services.Button, name string, send func() error) error {
	b.LastEvent.Set(name)
	return send()
}
