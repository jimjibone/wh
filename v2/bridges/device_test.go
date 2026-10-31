package bridges

import (
	"errors"
	"testing"
	"time"

	"github.com/jimjibone/wh/v2/bridges/attributes"
	"github.com/jimjibone/wh/v2/bridges/services"
	"github.com/jimjibone/wh/v2/types"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
)

// sink collects what a device sends.
func sink(dev *Device) <-chan *clientsapi.Device {
	ch := make(chan *clientsapi.Device, 16)
	dev.Init(func(state *clientsapi.Device) { ch <- state })
	return ch
}

func next(t *testing.T, ch <-chan *clientsapi.Device) *clientsapi.Device {
	t.Helper()
	select {
	case d := <-ch:
		return d
	case <-time.After(time.Second):
		t.Fatal("nothing sent")
		return nil
	}
}

// twoChannelRelay builds the shape of a two-output relay with a switch input
// per output: endpoints 1 and 2, each on_off + power_meter + binary_input.
func twoChannelRelay() (*Device, [2]*services.OnOff, [2]*services.BinaryInput) {
	dev := NewDevice("example2pm001", clientsapi.Device_DEVICE)
	var outs [2]*services.OnOff
	var ins [2]*services.BinaryInput
	for i, id := range []string{"1", "2"} {
		ep := dev.Endpoint(id)
		ep.SetLabel("Output " + id)
		outs[i] = services.NewOnOff("")
		ins[i] = services.NewBinaryInput("")
		ep.AddService(outs[i], services.NewPowerMeter(""), ins[i])
	}
	return dev, outs, ins
}

func TestFullStateHasEndpoints(t *testing.T) {
	dev, _, _ := twoChannelRelay()
	defer dev.Close()
	ch := sink(dev)
	dev.SendFullState()

	full := next(t, ch)
	if !full.FullState || len(full.Endpoints) != 3 {
		t.Fatalf("full state: %v", full)
	}
	root, one := full.Endpoints[0], full.Endpoints[1]
	if root.Id != RootEndpointID || len(root.Services) != 2 || root.Services[0].Typ != clientsapi.Service_INFO {
		t.Fatalf("root endpoint: %v", root)
	}
	if one.Id != "1" || one.Label != "Output 1" || len(one.Services) != 3 || one.Services[0].Id != "on_off" {
		t.Fatalf("endpoint 1: %v", one)
	}
}

func TestPartialUpdatesAndEventsCarryEndpoint(t *testing.T) {
	dev, outs, ins := twoChannelRelay()
	defer dev.Close()
	ch := sink(dev)
	dev.SendFullState() // as Bridge.AddDevice does; replaces construction-time updates
	next(t, ch)

	outs[1].On.Set(true)
	ins[1].On.Set(true)
	if err := ins[1].EmitChanged(types.BinaryInputChangedEvent{On: true}); err != nil {
		t.Fatal(err)
	}

	up := next(t, ch)
	if up.FullState || len(up.Endpoints) != 1 || up.Endpoints[0].Id != "2" || len(up.Endpoints[0].Services) != 2 {
		t.Fatalf("batched update: %v", up)
	}
	if len(up.Events) != 1 || up.Events[0].EndpointId != "2" || up.Events[0].ServiceId != "binary_input" || up.Events[0].Id != "changed" {
		t.Fatalf("event: %v", up.Events)
	}
}

func TestPendingEventsSentBeforeFullState(t *testing.T) {
	dev := NewDevice("examplebutton01", clientsapi.Device_DEVICE)
	defer dev.Close()
	button := services.NewButton("")
	dev.Endpoint("1").AddService(button)
	ch := sink(dev)

	if err := button.EmitPress(); err != nil {
		t.Fatal(err)
	}
	dev.SendFullState()

	first, second := next(t, ch), next(t, ch)
	if first.FullState || len(first.Events) != 1 {
		t.Fatalf("events should go first: %v", first)
	}
	if !second.FullState || len(second.Events) != 0 {
		t.Fatalf("then the full state, without events: %v", second)
	}
}

func TestHandleActionRoutesByEndpoint(t *testing.T) {
	dev, outs, _ := twoChannelRelay()
	defer dev.Close()
	var got [2]bool
	outs[0].On.OnWrite(func(v bool) error { got[0] = v; return nil })
	outs[1].On.OnWrite(func(v bool) error { got[1] = v; return nil })

	req := func(ep string) *clientsapi.ActionRequest {
		return &clientsapi.ActionRequest{
			EndpointId: ep, ServiceId: "on_off",
			Action: &clientsapi.ActionRequest_Write{Write: &clientsapi.WriteAction{Values: []*clientsapi.FieldValue{values.Field("on", values.Bool(true))}}},
		}
	}
	if _, err := dev.HandleAction(req("2"), nil); err != nil {
		t.Fatal(err)
	}
	if got != [2]bool{false, true} {
		t.Fatalf("writes went to %v", got)
	}

	_, err := dev.HandleAction(req("9"), nil)
	if actionErrorStatus(err) != clientsapi.ActionResponse_ERR_INVALID {
		t.Fatalf("unknown endpoint: %v", err)
	}
	_, err = dev.HandleAction(req(""), nil)
	if actionErrorStatus(err) != clientsapi.ActionResponse_ERR_INVALID {
		t.Fatalf("missing endpoint: %v", err)
	}
}

func TestActionErrorStatus(t *testing.T) {
	tests := []struct {
		err  error
		want clientsapi.ActionResponse_ActionStatus
	}{
		{attributes.ErrReadOnly, clientsapi.ActionResponse_ERR_READONLY},
		{&services.InvalidError{Err: errors.New("x")}, clientsapi.ActionResponse_ERR_INVALID},
		{errors.New("device offline"), clientsapi.ActionResponse_ERROR},
	}
	for _, tt := range tests {
		if got := actionErrorStatus(tt.err); got != tt.want {
			t.Errorf("%v: %s, want %s", tt.err, got, tt.want)
		}
	}
}

func TestRootServicesReplaceInfoAndOnline(t *testing.T) {
	dev := NewDevice("examplea", clientsapi.Device_DEVICE)
	defer dev.Close()
	info := services.NewInfo("")
	dev.AddService(info)
	if dev.Info != info {
		t.Fatal("Info should be replaced")
	}
	other := services.NewInfo("")
	dev.Endpoint("1").AddService(other)
	if dev.Info != info {
		t.Fatal("an Info on another endpoint must not replace the device's")
	}
}

func TestDeviceID(t *testing.T) {
	tests := []struct{ protocol, raw, want string }{
		{"zigbee", "0x00158D00045C089F", "zigbee00158d00045c089f"},
		{"tasmota", "AA:BB:CC:DD:EE:FF", "tasmotaaabbccddeeff"},
		{"frigate", "Front Door", "frigatefrontdoor"},
		{"matter", "vendor-1234-product-5", "mattervendor1234product5"},
	}
	for _, tt := range tests {
		got, err := DeviceID(tt.protocol, tt.raw)
		if err != nil || got != tt.want {
			t.Errorf("DeviceID(%q, %q) = %q, %v; want %q", tt.protocol, tt.raw, got, err, tt.want)
		}
	}
	for _, bad := range [][2]string{{"", "abc"}, {"Zigbee", "abc"}, {"zig-bee", "abc"}, {"zigbee", "0x"}, {"zigbee", "--"}} {
		if _, err := DeviceID(bad[0], bad[1]); err == nil {
			t.Errorf("DeviceID(%q, %q) should fail", bad[0], bad[1])
		}
	}
}
