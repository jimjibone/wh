package main

import (
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v1/bridges"
)

// SimPairing repeatedly pairs and unpairs a fake relay, showing how a bridge
// adds and removes devices at runtime. RemoveDevice removes the relay from
// Woodhouse even though it is online, as when a device is unpaired from its
// bridge. Blocks forever, so run it in a goroutine.
func SimPairing(bridge *bridges.Bridge, id, name string, interval time.Duration) {
	for {
		relay := NewFakeRelay(id, name)
		if err := bridge.AddDevice(relay.dev); err != nil {
			log.Errorf("failed to add device %q: %s", id, err)
			return
		}
		log.Infof("paired device %q", id)
		time.Sleep(interval)

		// RemoveDevice closes the device, so a new one is made for the next
		// pairing.
		if err := bridge.RemoveDevice(id); err != nil {
			log.Errorf("failed to remove device %q: %s", id, err)
			return
		}
		log.Infof("unpaired device %q", id)
		time.Sleep(interval)
	}
}
