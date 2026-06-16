package main

import (
	"time"

	"github.com/jimjibone/wh/v1/devices"
	"github.com/jimjibone/wh/v1/devices/services"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

type FakePresence struct {
	dev      *devices.Device
	info     *services.Info
	online   *services.Online
	presence *services.Presence
}

func NewFakePresence(id, name string, sim bool) *FakePresence {
	dev := &FakePresence{
		dev:      devices.NewDevice(id, clientsapi.Device_DEVICE),
		info:     services.NewInfo(),
		online:   services.NewOnline(),
		presence: services.NewPresence(""),
	}
	dev.dev.AddService(dev.info, dev.online, dev.presence)

	// Set up the info service.
	dev.info.Name.Set(name)
	dev.info.Model.Set("Fake Presence Thing")
	dev.info.Manufacturer.Set("Fake Things Inc")
	dev.online.Online.Set(true)
	dev.online.LastSeen.Set(time.Now())

	// Set default values.
	dev.presence.Motion.Set(false)
	dev.presence.Presence.Set(false)
	dev.presence.Distance.Set(0.0)

	// Forever simulate presence changes in the background.
	if sim {
		go func() {
			for {
				// Simulate person entering the room and moving closer.
				// Walk from 7.5m (outside range) to 1.0m.
				for dist := 7.5; dist >= 1.0; dist -= 0.5 {
					if dist <= 6.0 {
						dev.presence.Distance.Set(dist)
						dev.presence.Presence.Set(true)
						dev.presence.Motion.Set(true)
					} else {
						dev.presence.Presence.Set(false)
						dev.presence.Motion.Set(false)
					}
					time.Sleep(time.Second)
				}

				// Person stays still for a few seconds, motion goes false.
				time.Sleep(1 * time.Second)
				dev.presence.Motion.Set(false)
				time.Sleep(4 * time.Second)

				// Person moves around in the room.
				for i := range 5 {
					dev.presence.Motion.Set(true)
					dev.presence.Distance.Set(1.0 + float64(i)*0.8)
					dev.presence.Presence.Set(true)
					time.Sleep(time.Second)
				}

				// Person stays still again briefly.
				time.Sleep(1 * time.Second)
				dev.presence.Motion.Set(false)
				time.Sleep(3 * time.Second)

				// Person walks away and exits the room.
				for dist := 3.0; dist <= 7.5; dist += 0.5 {
					if dist > 6.0 {
						if dist > 7.0 {
							dev.presence.Distance.Set(0.0)
						}
						dev.presence.Presence.Set(false)
						dev.presence.Motion.Set(false)
					} else {
						dev.presence.Distance.Set(dist)
						dev.presence.Presence.Set(true)
						dev.presence.Motion.Set(true)
					}
					time.Sleep(time.Second)
				}

				// Person has left, no motion, no presence.
				dev.presence.Motion.Set(false)
				dev.presence.Presence.Set(false)

				// Wait before simulating the next entry.
				time.Sleep(5 * time.Second)
			}
		}()
	}

	return dev
}
