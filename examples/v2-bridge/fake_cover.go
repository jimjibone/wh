package main

import (
	"sync"
	"time"

	"github.com/jimjibone/wh/v2/bridges"
	"github.com/jimjibone/wh/v2/bridges/services"
	"github.com/jimjibone/wh/v2/types"
)

// FakeCover is a motorised blind with position and tilt. It moves 10% a
// second, reporting opening or closing while it moves.
type FakeCover struct {
	dev   *bridges.Device
	cover *services.Cover

	mu     sync.Mutex
	target int64
	moving bool
}

func NewFakeCover(uniqueID, name string) *FakeCover {
	f := &FakeCover{
		dev:   newDevice(uniqueID, name, "Fake Blind"),
		cover: services.NewCover("", services.CoverFeaturePosition, services.CoverFeatureTilt),
	}
	f.dev.Endpoint("1").AddService(f.cover)

	f.cover.OnOpen(func() error { f.moveTo(100); return nil })
	f.cover.OnClose(func() error { f.moveTo(0); return nil })
	f.cover.OnStop(func() error { f.moveTo(f.cover.Position.Get()); return nil })
	f.cover.OnGoTo(func(p types.CoverGoToParams) error { f.moveTo(p.Position); return nil })
	f.cover.OnTiltTo(func(p types.CoverTiltToParams) error { f.cover.Tilt.Set(p.Tilt); return nil })

	f.cover.Position.Set(0)
	f.cover.Tilt.Set(50)
	f.cover.State.Set("closed")
	return f
}

func (f *FakeCover) moveTo(target int64) {
	f.mu.Lock()
	f.target = target
	start := !f.moving
	f.moving = true
	f.mu.Unlock()
	if start {
		go f.run()
	}
}

func (f *FakeCover) run() {
	for {
		f.mu.Lock()
		pos, target := f.cover.Position.Get(), f.target
		if pos == target {
			f.moving = false
			f.mu.Unlock()
			switch target {
			case 0:
				f.cover.State.Set("closed")
			case 100:
				f.cover.State.Set("open")
			default:
				f.cover.State.Set("stopped")
			}
			return
		}
		f.mu.Unlock()
		if target > pos {
			f.cover.State.Set("opening")
			pos = min(target, pos+10)
		} else {
			f.cover.State.Set("closing")
			pos = max(target, pos-10)
		}
		f.cover.Position.Set(pos)
		time.Sleep(time.Second)
	}
}
