package main

import (
	"fmt"
	"os"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v1"
	"github.com/jimjibone/wh/v1/shared/stores"
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
				Usage: "simulate sensors",
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
			client := wh.NewClient(
				store,
				args.String("addr"),
				wh.WithClientID(args.String("id")),
				wh.WithClientInfo("Test Client", "Client for testing Woodhouse functionality", "0.1.0"),
				wh.WithImages(),
			)

			fake1 := NewFakeLightbulbColor("fake1", "Living Room Light")
			if err := client.AddDevice(fake1.dev); err != nil {
				log.Fatalf("failed to add device: %s", err)
			}

			fake1a := NewFakeLightbulbColorTemp("fake1a", "Hallway Light")
			if err := client.AddDevice(fake1a.dev); err != nil {
				log.Fatalf("failed to add device: %s", err)
			}

			fake1b := NewFakeLightbulbColorTemp("fake1b", "Kitchen Light")
			if err := client.AddDevice(fake1b.dev); err != nil {
				log.Fatalf("failed to add device: %s", err)
			}

			fake1c := NewFakeLightbulbColor("fake1c", "Bedroom Light")
			if err := client.AddDevice(fake1c.dev); err != nil {
				log.Fatalf("failed to add device: %s", err)
			}

			fake1d := NewFakeLightbulbColorTemp("fake1d", "Bedroom Lamp")
			if err := client.AddDevice(fake1d.dev); err != nil {
				log.Fatalf("failed to add device: %s", err)
			}

			fake1e := NewFakeLightbulbColorTemp("fake1e", "Hallway Lamp")
			if err := client.AddDevice(fake1e.dev); err != nil {
				log.Fatalf("failed to add device: %s", err)
			}

			fake2 := NewFakeRelay("fake2", "Boiler")
			if err := client.AddDevice(fake2.dev); err != nil {
				log.Fatalf("failed to add device: %s", err)
			}

			fake2a := NewFakeRelay("fake2a", "Washing Machine")
			if err := client.AddDevice(fake2a.dev); err != nil {
				log.Fatalf("failed to add device: %s", err)
			}

			fake3 := NewFakePresence("fake3", "Kitchen Presence", args.Bool("sim-sensors"))
			if err := client.AddDevice(fake3.dev); err != nil {
				log.Fatalf("failed to add device: %s", err)
			}

			fake4 := NewFakeCamera("fake4", "Front Door Camera")
			if err := client.AddDevice(fake4.dev); err != nil {
				log.Fatalf("failed to add device: %s", err)
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
