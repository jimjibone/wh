package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v2/bridges"
	"github.com/jimjibone/wh/v2/bridges/services"
	"github.com/jimjibone/wh/v2/types"
)

// FakeSwitch2 is a two-channel relay with a wall switch input per channel:
// endpoints 1 and 2 each hold on_off, power_meter and binary_input, and the
// root endpoint holds the device's temperature and a sharing service (a
// commissioning window another controller can use to add the device). The
// endpoints group each output with its own input and meter.
type FakeSwitch2 struct {
	dev     *bridges.Device
	outs    [2]*services.OnOff
	meters  [2]*services.PowerMeter
	inputs  [2]*services.BinaryInput
	temp    *services.DeviceTemperature
	sharing *services.Sharing
}

func NewFakeSwitch2(uniqueID, name string, sim bool) *FakeSwitch2 {
	f := &FakeSwitch2{
		dev:     newDevice(uniqueID, name, "Fake 2-Channel Relay"),
		temp:    services.NewDeviceTemperature(""),
		sharing: services.NewSharing(""),
	}
	f.dev.AddService(f.temp, f.sharing)
	for i := range 2 {
		ep := f.dev.Endpoint(fmt.Sprint(i + 1))
		ep.SetLabel(fmt.Sprintf("Output %d", i+1))
		f.outs[i] = services.NewOnOff("")
		f.meters[i] = services.NewPowerMeter("")
		f.inputs[i] = services.NewBinaryInput("")
		ep.AddService(f.outs[i], f.meters[i], f.inputs[i])

		f.outs[i].On.OnWrite(func(on bool) error {
			f.setOutput(i, on)
			return nil
		})
		f.outs[i].OnToggle(func() error {
			f.setOutput(i, !f.outs[i].On.Get())
			return nil
		})
		f.inputs[i].On.Set(false)
		f.setOutput(i, false)
	}
	f.temp.Temperature.Set(38)

	f.sharing.Open.Set(false)
	f.sharing.OnOpen(func(p types.SharingOpenParams) error {
		timeout := 3 * time.Minute
		if p.Timeout != nil {
			timeout = *p.Timeout
		}
		f.sharing.Code.Set(fmt.Sprintf("%011d", rand.Int63n(1e11)))
		f.sharing.QR.Set("MT:EXAMPLE0000000000")
		f.sharing.Ends.Set(time.Now().Add(timeout).Truncate(time.Second))
		f.sharing.Open.Set(true)
		time.AfterFunc(timeout, f.closeSharing)
		return nil
	})
	f.sharing.OnClose(func() error {
		f.closeSharing()
		return nil
	})

	if sim {
		go f.simulateInputs()
	}
	return f
}

func (f *FakeSwitch2) setOutput(i int, on bool) {
	f.outs[i].On.Set(on)
	if on {
		f.meters[i].Power.Set(60)
		f.meters[i].Current.Set(0.25)
	} else {
		f.meters[i].Power.Set(0)
		f.meters[i].Current.Set(0)
	}
	f.meters[i].Voltage.Set(240)
}

func (f *FakeSwitch2) closeSharing() {
	f.sharing.Open.Set(false)
	f.sharing.Code.Set("")
	f.sharing.QR.Set("")
}

// simulateInputs flips a wall switch now and then. The switch is wired to its
// own output, so the input follows the switch and the output follows the
// input; the changed event carries each flip.
func (f *FakeSwitch2) simulateInputs() {
	for {
		time.Sleep(time.Duration(10+rand.Intn(20)) * time.Second)
		i := rand.Intn(2)
		on := !f.inputs[i].On.Get()
		f.inputs[i].On.Set(on)
		if err := f.inputs[i].EmitChanged(types.BinaryInputChangedEvent{On: on}); err != nil {
			log.Errorf("%s: %s", f.dev.ID(), err)
		}
		f.setOutput(i, on)
	}
}
