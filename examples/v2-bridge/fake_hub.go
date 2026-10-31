package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v2/bridges"
	"github.com/jimjibone/wh/v2/bridges/services"
	"github.com/jimjibone/wh/v2/types"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
)

// FakeHub is a bridge device that pairs new devices, like a Zigbee
// coordinator or a Matter controller. Its pairing service opens a timed window
// (start returns when it closes) and accepts setup codes. A submitted code
// walks an entry of the recent list through joining, interviewing and paired
// (or failed), emits paired or failed, and adds the new device to the bridge
// so a UI can link straight to it.
//
// Codes are 8 digits; "00000000" fails, any other adds a plug.
type FakeHub struct {
	dev     *bridges.Device
	bridge  *bridges.Bridge
	pairing *services.Pairing

	mu     sync.Mutex
	recent []types.PairingRecentItem
	timer  *time.Timer
	next   int
}

func NewFakeHub(bridge *bridges.Bridge, uniqueID, name string) *FakeHub {
	id, err := bridges.DeviceID("example", uniqueID)
	if err != nil {
		log.Fatalf("bad example device id %q: %s", uniqueID, err)
	}
	f := &FakeHub{
		dev:     bridges.NewDevice(id, clientsapi.Device_BRIDGE),
		bridge:  bridge,
		pairing: services.NewPairing("", services.PairingFeatureTimedWindow, services.PairingFeatureCodeEntry),
		next:    1,
	}
	f.dev.Info.Name.Set(name)
	f.dev.Info.Model.Set("Fake Hub")
	f.dev.AddService(f.pairing)

	f.pairing.State.Set("idle")
	f.pairing.SetRecent(nil)
	f.pairing.OnStart(f.start)
	f.pairing.OnStop(func() error { f.stop(); return nil })
	f.pairing.OnSubmitCode(func(p types.PairingSubmitCodeParams) error {
		code := strings.TrimSpace(p.Code)
		if len(code) != 8 || strings.Trim(code, "0123456789") != "" {
			return errors.New("the code must be 8 digits")
		}
		go f.pair(code)
		return nil
	})
	return f
}

func (f *FakeHub) start(p types.PairingStartParams) (types.PairingStartResult, error) {
	d := 2 * time.Minute
	if p.Duration != nil && *p.Duration > 0 {
		d = *p.Duration
	}
	ends := time.Now().Add(d).Truncate(time.Second)
	f.mu.Lock()
	if f.timer != nil {
		f.timer.Stop()
	}
	f.timer = time.AfterFunc(d, f.stop)
	f.mu.Unlock()
	f.pairing.State.Set("open")
	f.pairing.Ends.Set(ends)
	f.pairing.Remaining.Set(d)
	return types.PairingStartResult{Ends: &ends}, nil
}

func (f *FakeHub) stop() {
	f.mu.Lock()
	if f.timer != nil {
		f.timer.Stop()
		f.timer = nil
	}
	f.mu.Unlock()
	f.pairing.State.Set("idle")
	f.pairing.Remaining.Set(0)
}

// pair simulates a device joining with a setup code.
func (f *FakeHub) pair(code string) {
	f.mu.Lock()
	n := f.next
	f.next++
	f.mu.Unlock()
	uniqueID := fmt.Sprintf("paired%d", n)
	deviceID, _ := bridges.DeviceID("example", uniqueID)
	name := fmt.Sprintf("Paired Plug %d", n)

	f.pairing.State.Set("busy")
	f.record(deviceID, name, "joining", "")
	time.Sleep(2 * time.Second)
	f.record(deviceID, name, "interviewing", "")
	time.Sleep(3 * time.Second)

	if code == "00000000" {
		msg := "device did not answer the interview"
		f.record(deviceID, name, "failed", msg)
		f.pairing.LastError.Set(msg)
		f.emit(f.pairing.EmitFailed(types.PairingFailedEvent{DeviceID: &deviceID, Error: msg}))
		f.pairing.State.Set("open")
		return
	}
	plug := NewFakePlug(uniqueID, name, 15)
	if err := f.bridge.AddDevice(plug.dev); err != nil {
		f.record(deviceID, name, "failed", err.Error())
		f.emit(f.pairing.EmitFailed(types.PairingFailedEvent{DeviceID: &deviceID, Error: err.Error()}))
		return
	}
	f.record(deviceID, name, "paired", "")
	f.emit(f.pairing.EmitPaired(types.PairingPairedEvent{DeviceID: deviceID, Name: &name}))
	f.pairing.State.Set("open")
}

// record adds or updates a recent entry, keeping the newest 10.
func (f *FakeHub) record(deviceID, name, state, errMsg string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	entry := types.PairingRecentItem{DeviceID: deviceID, Name: &name, State: state, Time: time.Now().Truncate(time.Second)}
	if errMsg != "" {
		entry.Error = &errMsg
	}
	for i, r := range f.recent {
		if r.DeviceID == deviceID {
			f.recent = append(f.recent[:i], f.recent[i+1:]...)
			break
		}
	}
	f.recent = append([]types.PairingRecentItem{entry}, f.recent...)
	if len(f.recent) > 10 {
		f.recent = f.recent[:10]
	}
	f.pairing.SetRecent(f.recent)
}

func (f *FakeHub) emit(err error) {
	if err != nil {
		log.Errorf("%s: %s", f.dev.ID(), err)
	}
}
