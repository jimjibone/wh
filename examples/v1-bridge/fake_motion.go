package main

import (
	"time"

	"github.com/jimjibone/wh/v1/bridges"
	"github.com/jimjibone/wh/v1/bridges/services"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

type FakeMotion struct {
	dev    *bridges.Device
	motion *services.Motion
}

func NewFakeMotion(id, name string, sim bool) *FakeMotion {
	dev := &FakeMotion{
		dev:    bridges.NewDevice(id, clientsapi.Device_DEVICE),
		motion: services.NewMotion(""),
	}
	dev.dev.AddService(dev.motion)

	// Set up the info service.
	dev.dev.Info.Name.Set(name)
	dev.dev.Info.Model.Set("Fake Presence Thing")
	dev.dev.Info.Manufacturer.Set("Fake Things Inc")
	dev.dev.Online.Online.Set(true)
	dev.dev.Online.LastSeen.Set(time.Now())

	// Set default values.
	dev.motion.Motion.Set(false)

	// Forever simulate presence changes in the background.
	if sim {
		go func() {
			for {
				// Simulate person entering the room.
				dev.motion.Motion.Set(true)

				// Person stays still for a few seconds, motion goes false.
				time.Sleep(5 * time.Second)
				dev.motion.Motion.Set(false)

				// Person moves around in the room.
				time.Sleep(4 * time.Second)
				dev.motion.Motion.Set(true)

				// Person stays still again.
				time.Sleep(5 * time.Second)
				dev.motion.Motion.Set(false)

				// Wait before simulating the next entry.
				time.Sleep(5 * time.Second)
			}
		}()
	}

	return dev
}
