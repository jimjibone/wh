package main

import (
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v2/bridges"
)

// SimPairing repeatedly adds and removes a fake plug, showing how a bridge
// adds and removes devices at runtime. RemoveDevice removes the plug from
// Woodhouse even though it is online, as when a device is unpaired from its
// bridge. Blocks forever, so run it in a goroutine.
func SimPairing(bridge *bridges.Bridge, uniqueID, name string, interval time.Duration) {
	for {
		plug := NewFakePlug(uniqueID, name, 5)
		if err := bridge.AddDevice(plug.dev); err != nil {
			log.Errorf("failed to add device %q: %s", plug.dev.ID(), err)
			return
		}
		log.Infof("paired device %q", plug.dev.ID())
		time.Sleep(interval)
		// RemoveDevice closes the device, so a new one is made for the next
		// pairing.
		if err := bridge.RemoveDevice(plug.dev.ID()); err != nil {
			log.Errorf("failed to remove device %q: %s", plug.dev.ID(), err)
			return
		}
		log.Infof("unpaired device %q", plug.dev.ID())
		time.Sleep(interval)
	}
}
