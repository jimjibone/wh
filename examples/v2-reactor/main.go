// Command v2-reactor automates the devices of the v2-bridge example. It shows
// typed views addressed by device and endpoint, attribute updates, events
// (live and replayed), commands with params and results, writes, feature
// checks, a generic view, a device view, and a device the reactor provides
// itself.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v2/bridges"
	"github.com/jimjibone/wh/v2/bridges/services"
	"github.com/jimjibone/wh/v2/clients"
	"github.com/jimjibone/wh/v2/reactors"
	"github.com/jimjibone/wh/v2/shared/stores"
	"github.com/jimjibone/wh/v2/types"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/urfave/cli/v2"
)

// Devices of the v2-bridge example.
const (
	livingRoomLight = "examplelamp1"
	hallwayLight    = "examplelamp2"
	landingSwitch   = "exampleswitch1"
	bedsideRemote   = "exampleremote1"
	kitchenPresence = "examplepresence1"
	rainTank        = "exampletank1"
	frontCamera     = "examplecamera1"
	bedroomTRV      = "exampletrv1"
	loungeTV        = "exampletv1"
	exampleHub      = "examplehub1"
)

func main() {
	app := &cli.App{
		Name:  "woodhouse-reactor",
		Usage: "Automates the devices of the example bridge",
		Flags: []cli.Flag{
			&cli.PathFlag{Name: "store", Usage: "path to config storage location", Required: true},
			&cli.StringFlag{Name: "addr", Usage: "woodhouse server address", Required: true},
			&cli.StringFlag{Name: "id", Usage: "ID used by this reactor", Value: "example-reactor"},
			&cli.BoolFlag{Name: "debug", Aliases: []string{"v"}, Usage: "enable debug logging"},
		},
		Before: func(args *cli.Context) error {
			if args.Bool("debug") {
				log.SetOptions(log.WithMinLevel(log.DebugLevel))
			} else {
				log.SetOptions(log.WithMinLevel(log.InfoLevel))
			}
			return nil
		},
		Action: func(args *cli.Context) error {
			rc := reactors.NewReactor(bridges.BridgeConfig{
				ClientConfig: clients.ClientConfig{
					Store:             stores.NewFSStore(args.String("store")),
					ServerAddr:        args.String("addr"),
					ClientID:          args.String("id"),
					ClientName:        "Example Reactor",
					ClientDescription: "Automates the example bridge's devices",
					ClientVersion:     "0.1.0",
				},
			})
			away := awayMode(rc)
			lights(rc, away)
			remote(rc)
			inputs(rc)
			security(rc)
			pairing(rc)
			go startup(rc)
			return rc.Run()
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}

// awayMode is a device the reactor provides itself: an on_off service other
// automations read, and which users can switch.
func awayMode(rc *reactors.Reactor) *services.OnOff {
	dev := bridges.NewDevice("examplereactoraway", clientsapi.Device_REACTOR)
	dev.Info.Name.Set("Away Mode")
	sw := services.NewOnOff("")
	dev.Endpoint("1").AddService(sw)
	sw.On.Set(false)
	sw.On.OnWrite(func(on bool) error {
		log.Infof("away mode %t", on)
		sw.On.Set(on)
		return nil
	})
	sw.OnToggle(func() error { sw.On.Set(!sw.On.Get()); return nil })
	if err := rc.Bridge().AddDevice(dev); err != nil {
		log.Fatalf("add away mode device: %s", err)
	}
	return sw
}

// lights turns the hallway light on with motion and off when nobody is there.
// Motion arrives as an event, presence as an attribute change.
func lights(rc *reactors.Reactor, away *services.OnOff) {
	presence := rc.Presence(kitchenPresence, "1")
	motion := rc.Motion(kitchenPresence, "1")
	hall := rc.Light(hallwayLight, "1")

	motion.OnDetected(func(info reactors.EventInfo) {
		if away.On.Get() || hall.On() {
			return
		}
		log.Infof("motion in the kitchen at %s: hallway light on", info.Time.Format(time.TimeOnly))
		p := types.LightSetParams{On: ptr(true), Transition: ptr(2 * time.Second)}
		if hall.HasFeature(services.LightFeatureBrightness) {
			p.Brightness = ptr(int64(60))
		}
		if err := hall.Set(context.Background(), p); err != nil {
			log.Errorf("hallway light: %s", err)
		}
	})
	presence.OnUpdate(func(changed bool) {
		if !changed || presence.Presence() || !hall.On() {
			return
		}
		log.Infof("kitchen empty: hallway light off over 10s")
		if err := hall.Set(context.Background(), types.LightSetParams{On: ptr(false), Transition: ptr(10 * time.Second)}); err != nil {
			log.Errorf("hallway light: %s", err)
		}
	})
}

// remote drives the living room light from the two-button remote: one view
// per endpoint (button). A single press toggles, a double press goes warm, a
// hold dims; the Off button turns everything off.
func remote(rc *reactors.Reactor) {
	onButton := rc.Button(bedsideRemote, "1")
	offButton := rc.Button(bedsideRemote, "2")
	living := rc.Light(livingRoomLight, "1")
	hall := rc.Light(hallwayLight, "1")
	ctx := context.Background()

	onButton.OnMultiPress(func(e types.ButtonMultiPressEvent, info reactors.EventInfo) {
		var err error
		switch e.Count {
		case 1:
			err = living.SetOn(ctx, !living.On())
		case 2:
			if living.HasColorTemp() {
				err = living.Set(ctx, types.LightSetParams{On: ptr(true), ColorTemp: ptr(int64(2700))})
			}
		}
		if err != nil {
			log.Errorf("living room light: %s", err)
		}
	})
	onButton.OnHold(func(info reactors.EventInfo) {
		if err := living.SetBrightness(ctx, max(1, living.Brightness()-30)); err != nil {
			log.Errorf("living room light: %s", err)
		}
	})
	offButton.OnPress(func(info reactors.EventInfo) {
		for _, l := range []*reactors.LightView{living, hall} {
			if err := l.SetOn(ctx, false); err != nil {
				log.Errorf("%s: %s", l, err)
			}
		}
	})
}

// inputs logs the landing switch's wall inputs, one view per endpoint.
func inputs(rc *reactors.Reactor) {
	for _, ep := range []string{"1", "2"} {
		in := rc.BinaryInput(landingSwitch, ep)
		in.OnChanged(func(e types.BinaryInputChangedEvent, info reactors.EventInfo) {
			log.Infof("landing switch input %s changed to %t", ep, e.On)
		})
	}
}

// security logs camera detections, keeping replayed ones up to 10 minutes
// old: a detection matters even if the reactor was briefly offline.
func security(rc *reactors.Reactor) {
	cam := rc.Camera(frontCamera, "1")
	cam.OnDetection(func(e types.CameraDetectionEvent, info reactors.EventInfo) {
		score := 0.0
		if e.Score != nil {
			score = *e.Score
		}
		replayed := ""
		if info.Replayed {
			replayed = fmt.Sprintf(" (replayed, %s ago)", info.Age().Truncate(time.Second))
		}
		log.Infof("front door: %s %.0f%%%s", e.Label, score*100, replayed)
	}, reactors.MaxAge(10*time.Minute))
	rc.OnEventsLost(func() { log.Warnf("some events were lost while disconnected") })
}

// pairing reports devices paired through the hub.
func pairing(rc *reactors.Reactor) {
	hub := rc.Pairing(exampleHub, "0")
	hub.OnPaired(func(e types.PairingPairedEvent, info reactors.EventInfo) {
		name := e.DeviceID
		if e.Name != nil {
			name = *e.Name
		}
		log.Infof("paired %s (%s)", name, e.DeviceID)
	}, reactors.MaxAge(time.Hour))
	hub.OnFailed(func(e types.PairingFailedEvent, info reactors.EventInfo) {
		log.Warnf("pairing failed: %s", e.Error)
	}, reactors.MaxAge(time.Hour))
}

// startup runs once the initial device states are in: reads that need state,
// and commands with results.
func startup(rc *reactors.Reactor) {
	<-rc.Ready()
	ctx := context.Background()
	log.Infof("reactor ready")

	// A device view lists endpoints and their services.
	for _, ep := range rc.Device(landingSwitch).Endpoints() {
		log.Infof("landing switch endpoint %q (%s): %d services", ep.ID, ep.Label, len(ep.Services))
	}

	// A struct list attribute reads as typed items.
	trv := rc.Thermostat(bedroomTRV, "1")
	for _, e := range trv.Schedule() {
		log.Infof("bedroom schedule: %v from %s at %.1fC", e.Days, e.Start, e.Setpoint)
	}

	// A generic view reads any attribute, here the tank's mapped level.
	tank := rc.Generic(rainTank, "1", "analog_input")
	log.Infof("rain tank: %.0f litres", tank.Attribute("mapped").GetValue().GetFloat())

	// The TV lists its keys as the options of the send_key param.
	tv := rc.Remote(loungeTV, "1")
	if spec := tv.CommandSpec("send_key"); spec != nil {
		log.Infof("tv keys: %v", spec.GetParams()[0].GetSpec().GetEnum().GetOptions())
	}
	if err := tv.SendKey(ctx, types.RemoteSendKeyParams{Key: "home"}); err != nil {
		log.Errorf("tv: %s", err)
	}

	// A command with a result: the hub says when its pairing window closes.
	hub := rc.Pairing(exampleHub, "0")
	res, err := hub.Start(ctx, types.PairingStartParams{Duration: ptr(2 * time.Minute)})
	if err != nil {
		log.Errorf("hub: %s", err)
	} else if res.Ends != nil {
		log.Infof("hub pairing open until %s", res.Ends.Format(time.TimeOnly))
	}

	// A failing write shows the error from the bridge.
	if err := rc.Light("examplenosuchlamp", "1").SetOn(ctx, true); err != nil {
		log.Infof("write to an unknown device fails as expected: %s", err)
	}
}

func ptr[T any](v T) *T { return &v }
