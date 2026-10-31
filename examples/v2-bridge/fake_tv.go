package main

import (
	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v2/bridges"
	"github.com/jimjibone/wh/v2/bridges/services"
	"github.com/jimjibone/wh/v2/types"
)

// FakeTV is a TV: power (on_off), volume and a remote. The remote's send_key
// command takes an enum whose options are this TV's keys, filled in at run
// time: the wire spec is per service, so each TV can list its own keys.
type FakeTV struct {
	dev    *bridges.Device
	power  *services.OnOff
	volume *services.Volume
	remote *services.Remote
}

var fakeTVKeys = []string{"power", "home", "back", "up", "down", "left", "right", "ok", "volume_up", "volume_down", "mute", "netflix"}

func NewFakeTV(uniqueID, name string) *FakeTV {
	f := &FakeTV{
		dev:    newDevice(uniqueID, name, "Fake TV"),
		power:  services.NewOnOff(""),
		volume: services.NewVolume(""),
		remote: services.NewRemote(""),
	}
	// Fill the key options before the service is registered.
	for _, p := range f.remote.CommandSpec("send_key").GetParams() {
		if p.GetId() == "key" {
			p.GetSpec().GetEnum().Options = fakeTVKeys
		}
	}
	f.dev.Endpoint("1").AddService(f.power, f.volume, f.remote)

	f.power.On.OnWrite(func(on bool) error { f.power.On.Set(on); return nil })
	f.power.OnToggle(func() error { f.power.On.Set(!f.power.On.Get()); return nil })
	f.volume.Level.OnWrite(func(v int64) error { f.volume.Level.Set(v); return nil })
	f.volume.Muted.OnWrite(func(v bool) error { f.volume.Muted.Set(v); return nil })
	f.volume.OnStep(func(p types.VolumeStepParams) error {
		f.volume.Level.Set(min(100, max(0, f.volume.Level.Get()+p.Delta)))
		return nil
	})
	f.remote.OnSendKey(func(p types.RemoteSendKeyParams) error {
		log.Infof("%s: key %q", f.dev.ID(), p.Key)
		switch p.Key {
		case "power":
			f.power.On.Set(!f.power.On.Get())
		case "volume_up":
			f.volume.Level.Set(min(100, f.volume.Level.Get()+1))
		case "volume_down":
			f.volume.Level.Set(max(0, f.volume.Level.Get()-1))
		case "mute":
			f.volume.Muted.Set(!f.volume.Muted.Get())
		}
		return nil
	})

	f.power.On.Set(false)
	f.volume.Level.Set(12)
	f.volume.Muted.Set(false)
	return f
}
