package main

import (
	"context"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v1/bridges"
	"github.com/jimjibone/wh/v1/clients"
	"github.com/jimjibone/wh/v1/reactors"
	"github.com/jimjibone/wh/v1/shared/stores"
)

func main() {
	// Connect to the core as a reactor client.
	reactor := reactors.NewReactor(
		bridges.BridgeConfig{
			ClientConfig: clients.ClientConfig{
				Store:             stores.NewFSStore("reactor.db"),
				ServerAddr:        "localhost:4000",
				ClientID:          "example-reactor",
				ClientName:        "Example Reactor",
				ClientDescription: "Turns a light on when motion is detected",
				ClientVersion:     "1.0.0",
			},
			ImagesEnabled: false,
		},
	)

	// Grab typed handles to the devices we care about.
	motion := reactor.GetMotion("fake2b")   // Hallway motion sensor
	light := reactor.GetLightbulb("fake1a") // Hallway light

	// React to motion: light follows the sensor.
	motion.OnUpdate(func(changed bool) {
		if err := light.SetOn(context.Background(), motion.Motion()); err != nil {
			log.Errorf("failed to set light: %s", err)
		}
	})

	go func() {
		// Wait until the reactor is connected and devices are known.
		<-reactor.Ready()
		log.Infof("motion-light reactor started")
	}()

	if err := reactor.Run(); err != nil {
		log.Fatalln(err)
	}
}
