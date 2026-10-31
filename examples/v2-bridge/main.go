package main

import (
	"fmt"
	"os"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v2/bridges"
	"github.com/jimjibone/wh/v2/bridges/services"
	"github.com/jimjibone/wh/v2/clients"
	"github.com/jimjibone/wh/v2/shared/stores"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "woodhouse-bridge",
		Usage:                "Implements a fake bridge for testing",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:     "store",
				Usage:    "path to config storage location",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "addr",
				Usage:    "woodhouse server address (disables automatic discovery)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "id",
				Usage: "ID used by this bridge",
				Value: "example-bridge",
			},
			&cli.BoolFlag{
				Name:  "sim-sensors",
				Usage: "simulate sensors, inputs, buttons and other changes",
			},
			&cli.DurationFlag{
				Name:  "sim-pairing",
				Usage: "pair and unpair a fake plug at this interval (0 disables)",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"v"},
				Usage:   "enable debug logging",
			},
		},
		Before: func(args *cli.Context) error {
			// Setup logging.
			if args.Bool("debug") {
				log.SetOptions(log.WithMinLevel(log.DebugLevel))
			} else {
				log.SetOptions(log.WithMinLevel(log.InfoLevel))
			}
			return nil
		},
		Action: func(args *cli.Context) error {
			// Create the store.
			store := stores.NewFSStore(args.String("store"))

			// Create the client.
			client := bridges.NewBridge(bridges.BridgeConfig{
				ClientConfig: clients.ClientConfig{
					Store:             store,
					ServerAddr:        args.String("addr"),
					ClientID:          args.String("id"),
					ClientName:        "Example Bridge",
					ClientDescription: "Client for testing Woodhouse functionality",
					ClientVersion:     "0.1.0",
				},
				ImagesEnabled: true,
			})

			sim := args.Bool("sim-sensors")
			devices := []*bridges.Device{
				// Lights: features gate attributes and the set command's params.
				NewFakeLight("lamp1", "Living Room Light", services.LightFeatureBrightness, services.LightFeatureColorTemp, services.LightFeatureHS).dev,
				NewFakeLight("lamp2", "Hallway Light", services.LightFeatureBrightness, services.LightFeatureColorTemp).dev,
				NewFakeLight("lamp3", "Porch Light").dev,
				// Outputs and inputs: endpoints group each output with its meter and input.
				NewFakePlug("plug1", "Boiler", 3).dev,
				NewFakePlug("plug2", "Washing Machine", 2000).dev,
				NewFakeSwitch2("switch1", "Landing Switch", sim).dev,
				NewFakeRemote("remote1", "Bedside Remote", sim).dev,
				// Sensors, with device sample times and events.
				NewFakeClimateSensor("sensor1", "Bedroom Sensor", sim).dev,
				NewFakePresence("presence1", "Kitchen Presence", sim).dev,
				NewFakeTankLevel("tank1", "Rain Tank", sim).dev,
				NewFakeCamera("camera1", "Front Door Camera", sim).dev,
				// Commands with params, lists of structs and run-time enum options.
				NewFakeThermostat("trv1", "Bedroom Radiator", sim).dev,
				NewFakeCover("blind1", "Study Blind").dev,
				NewFakeSpeaker("speaker1", "Kitchen Speaker").dev,
				NewFakeTV("tv1", "Lounge TV").dev,
				// Pairing: start returns a result; submit_code adds a device.
				NewFakeHub(client, "hub1", "Example Hub").dev,
			}
			for _, dev := range devices {
				if err := client.AddDevice(dev); err != nil {
					log.Fatalf("failed to add device %q: %s", dev.ID(), err)
				}
			}

			if interval := args.Duration("sim-pairing"); interval > 0 {
				go SimPairing(client, "plug9", "Spare Plug", interval)
			}

			err := client.Run()
			if err != nil {
				return fmt.Errorf("failed to run client: %s", err)
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}
