package main

import (
	"fmt"
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v2/bridges"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
)

// newDevice creates a device with the example ID convention
// ("example<unique-id>") and fills in its info service. Renaming is supported:
// a real bridge would forward the name to its device and let the device's
// report drive Name.Set; here the device is simulated with a short delay.
// Renaming to "fail" shows an error reaching the user.
func newDevice(uniqueID, name, model string) *bridges.Device {
	id, err := bridges.DeviceID("example", uniqueID)
	if err != nil {
		log.Fatalf("bad example device id %q: %s", uniqueID, err)
	}
	dev := bridges.NewDevice(id, clientsapi.Device_DEVICE)
	dev.Info.Name.Set(name)
	dev.Info.Model.Set(model)
	dev.Info.Manufacturer.Set("Fake Things Inc")
	dev.Info.SerialNumber.Set(uniqueID)
	dev.Info.FirmwareVersion.Set("1.0.0")
	dev.Info.EnableRename(func(newName string) error {
		time.Sleep(500 * time.Millisecond)
		if newName == "fail" {
			return fmt.Errorf("device rejected the name %q", newName)
		}
		log.Infof("%s renamed to %q", id, newName)
		dev.Info.Name.Set(newName)
		return nil
	})
	return dev
}

// ptr returns a pointer to v, for optional fields of generated types.
func ptr[T any](v T) *T { return &v }
