package reactors

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/jimjibone/wh/v2/bridges"
	"github.com/jimjibone/wh/v2/clients"
	"github.com/jimjibone/wh/v2/shared/stores"
	"github.com/jimjibone/wh/v2/types"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
	"google.golang.org/grpc"
)

const testDevice = "examplelamp01"

func newTestReactor() *Reactor {
	return NewReactor(bridges.BridgeConfig{ClientConfig: clients.ClientConfig{
		Store:      stores.NewMemStore(),
		ServerAddr: "127.0.0.1:1",
		ClientID:   "reactor-1",
	}})
}

func attr(id string, v *clientsapi.Value, spec *clientsapi.Spec) *clientsapi.Attribute {
	return &clientsapi.Attribute{Id: id, Value: v, Spec: spec, Perms: clientsapi.Permissions_PERM_READWRITE}
}

var (
	boolSpec = &clientsapi.Spec{Kind: &clientsapi.Spec_Bool{Bool: &clientsapi.BoolSpec{}}}
	intSpec  = &clientsapi.Spec{Kind: &clientsapi.Spec_Int{Int: &clientsapi.IntSpec{}}}
)

// fullState is a lamp with a dimmable light on endpoint 1 and a button on
// endpoint 2.
func fullState() *clientsapi.DeviceStreamResponse {
	return &clientsapi.DeviceStreamResponse{Device: &clientsapi.Device{
		Id: testDevice, FullState: true, Typ: clientsapi.Device_DEVICE,
		Endpoints: []*clientsapi.Endpoint{
			{Id: "0", Services: []*clientsapi.Service{
				{Id: "info", Typ: clientsapi.Service_INFO, Attrs: []*clientsapi.Attribute{attr("name", values.Text("Desk lamp"), nil)}},
				{Id: "online", Typ: clientsapi.Service_ONLINE, Attrs: []*clientsapi.Attribute{attr("online", values.Bool(true), nil)}},
			}},
			{Id: "1", Label: "Lamp", Services: []*clientsapi.Service{
				{Id: "light", Typ: clientsapi.Service_LIGHT, Features: []string{"brightness"}, Attrs: []*clientsapi.Attribute{
					attr("on", values.Bool(false), boolSpec),
					attr("brightness", values.Int(30), intSpec),
				}},
			}},
			{Id: "2", Services: []*clientsapi.Service{
				{Id: "button", Typ: clientsapi.Service_BUTTON, Attrs: []*clientsapi.Attribute{attr("last_event", values.Enum("press"), nil)}},
			}},
		},
	}}
}

func partial(endpoint, service string, attrs ...*clientsapi.Attribute) *clientsapi.DeviceStreamResponse {
	return &clientsapi.DeviceStreamResponse{Device: &clientsapi.Device{
		Id:        testDevice,
		Endpoints: []*clientsapi.Endpoint{{Id: endpoint, Services: []*clientsapi.Service{{Id: service, Attrs: attrs}}}},
	}}
}

func eventUpdate(e *clientsapi.Event) *clientsapi.DeviceStreamResponse {
	return &clientsapi.DeviceStreamResponse{Device: &clientsapi.Device{Id: testDevice, Events: []*clientsapi.Event{e}}}
}

func TestViewsMirrorState(t *testing.T) {
	rc := newTestReactor()
	light := rc.Light(testDevice, "1")
	var changes []bool
	light.OnUpdate(func(changed bool) { changes = append(changes, changed) })

	if light.Exists() || light.On() || light.Brightness() != 0 {
		t.Fatal("view should read as zero before the device is known")
	}

	rc.handle(fullState())
	if !light.Exists() || light.On() || light.Brightness() != 30 || !light.HasFeature("brightness") {
		t.Fatalf("after full state: exists %t on %t brightness %d", light.Exists(), light.On(), light.Brightness())
	}
	if light.DeviceName() != "Desk lamp" || !light.Online() {
		t.Fatal("device info and online not read from the root endpoint")
	}

	// A value-only partial update keeps the known spec.
	rc.handle(partial("1", "light", &clientsapi.Attribute{Id: "brightness", Value: values.Int(80)}))
	if light.Brightness() != 80 || light.Attribute("brightness").GetSpec() == nil {
		t.Fatalf("partial update: %v", light.Attribute("brightness"))
	}
	// The same value again is an update without a change.
	rc.handle(partial("1", "light", &clientsapi.Attribute{Id: "brightness", Value: values.Int(80)}))
	if len(changes) != 3 || !changes[0] || !changes[1] || changes[2] {
		t.Fatalf("changes = %v, want [true true false]", changes)
	}

	// A view on the wrong endpoint or of the wrong type sees nothing.
	if rc.Light(testDevice, "2").Exists() || rc.OnOff(testDevice, "1", "light").Exists() {
		t.Fatal("views must match endpoint and type")
	}
	if !rc.Generic(testDevice, "1", "light").Exists() {
		t.Fatal("a generic view takes any type")
	}

	dev := rc.Device(testDevice)
	if eps := dev.Endpoints(); len(eps) != 3 || eps[1].Label != "Lamp" || eps[1].Services[0].Typ != clientsapi.Service_LIGHT {
		t.Fatalf("endpoints: %+v", eps)
	}
}

func TestDeviceRemoved(t *testing.T) {
	rc := newTestReactor()
	light := rc.Light(testDevice, "1")
	rc.handle(fullState())
	var removed bool
	light.OnUpdate(func(changed bool) { removed = changed && !light.Exists() })
	rc.handle(&clientsapi.DeviceStreamResponse{DeviceRemoved: testDevice})
	if !removed || rc.Device(testDevice).Exists() {
		t.Fatal("removal should notify views and clear the device")
	}
}

func TestEventsAndReplayPolicy(t *testing.T) {
	rc := newTestReactor()
	now := time.Unix(1_800_000_000, 0)
	rc.now = func() time.Time { return now }
	button := rc.Button(testDevice, "2")

	var presses []EventInfo
	var multi []int64
	var history []EventInfo
	button.OnPress(func(info EventInfo) { presses = append(presses, info) })
	button.OnMultiPress(func(e types.ButtonMultiPressEvent, info EventInfo) { multi = append(multi, e.Count) })
	button.OnPress(func(info EventInfo) { history = append(history, info) }, MaxAge(time.Hour))

	ev := func(id string, seq uint64, age time.Duration, replayed bool, fields ...*clientsapi.FieldValue) *clientsapi.Event {
		return &clientsapi.Event{
			EndpointId: "2", ServiceId: "button", Id: id, Fields: fields,
			Time: values.TimeToPb(now.Add(-age)), Epoch: "e1", Seq: seq, Replayed: replayed,
		}
	}

	rc.handle(eventUpdate(ev("press", 1, 0, false)))
	rc.handle(eventUpdate(ev("multi_press", 2, 0, false, values.Field("count", values.Int(2)))))
	rc.handle(eventUpdate(ev("press", 3, 10*time.Second, true)))  // recent replay: delivered
	rc.handle(eventUpdate(ev("press", 4, 10*time.Minute, true)))  // stale replay: only to the history handler
	rc.handle(eventUpdate(ev("press", 5, 10*time.Minute, false))) // old but live: delivered

	if len(presses) != 3 || presses[1].Seq != 3 || !presses[1].Replayed || presses[2].Seq != 5 {
		t.Fatalf("presses: %+v", presses)
	}
	if len(history) != 4 {
		t.Fatalf("history handler got %d presses, want 4", len(history))
	}
	if len(multi) != 1 || multi[0] != 2 {
		t.Fatalf("multi presses: %v", multi)
	}
	if rc.cursor.GetEpoch() != "e1" || rc.cursor.GetSeq() != 5 {
		t.Fatalf("cursor: %v", rc.cursor)
	}

	// Events for another endpoint do not reach the view.
	other := ev("press", 6, 0, false)
	other.EndpointId = "1"
	rc.handle(eventUpdate(other))
	if len(presses) != 3 {
		t.Fatal("event for another endpoint delivered")
	}
}

func TestEventsLost(t *testing.T) {
	rc := newTestReactor()
	lost := false
	rc.OnEventsLost(func() { lost = true })
	rc.handle(&clientsapi.DeviceStreamResponse{EventsLost: true})
	if !lost {
		t.Fatal("events lost not reported")
	}
}

// fakeActions answers SendAction with canned responses and records requests.
type fakeActions struct {
	clientsapi.ClientServiceClient
	responses []*clientsapi.ActionResponse
	got       *clientsapi.ActionRequest
}

func (f *fakeActions) SendAction(ctx context.Context, in *clientsapi.ActionRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[clientsapi.ActionResponse], error) {
	f.got = in
	return &fakeStream{responses: f.responses}, nil
}

type fakeStream struct {
	grpc.ClientStream
	responses []*clientsapi.ActionResponse
}

func (s *fakeStream) Recv() (*clientsapi.ActionResponse, error) {
	if len(s.responses) == 0 {
		return nil, io.EOF
	}
	r := s.responses[0]
	s.responses = s.responses[1:]
	return r, nil
}

func TestCommandsAndWrites(t *testing.T) {
	rc := newTestReactor()
	ends := time.Unix(1_800_000_600, 0)
	fake := &fakeActions{responses: []*clientsapi.ActionResponse{
		{Status: clientsapi.ActionResponse_QUEUED},
		{Status: clientsapi.ActionResponse_COMPLETE, Result: []*clientsapi.FieldValue{values.Field("ends", values.Time(ends))}},
	}}
	rc.service = fake

	pairing := rc.Pairing("examplehub01", "0")
	d := 2 * time.Minute
	res, err := pairing.Start(context.Background(), types.PairingStartParams{Duration: &d})
	if err != nil {
		t.Fatal(err)
	}
	if res.Ends == nil || !res.Ends.Equal(ends) {
		t.Fatalf("result: %+v", res)
	}
	cmd := fake.got.GetCommand()
	if fake.got.GetDeviceId() != "examplehub01" || fake.got.GetEndpointId() != "0" || fake.got.GetServiceId() != "pairing" || cmd.GetId() != "start" {
		t.Fatalf("request: %v", fake.got)
	}
	if got, _ := values.AsDuration(values.Fields(cmd.GetParams())["duration"]); got != d {
		t.Fatalf("duration param: %v", cmd.GetParams())
	}

	fake.responses = []*clientsapi.ActionResponse{{Status: clientsapi.ActionResponse_ERR_READONLY, Details: "nope"}}
	err = rc.Light(testDevice, "1").SetBrightness(context.Background(), 50)
	var actionErr *ActionError
	if !errors.As(err, &actionErr) || actionErr.Status != clientsapi.ActionResponse_ERR_READONLY {
		t.Fatalf("write error: %v", err)
	}
	if w := fake.got.GetWrite().GetValues(); len(w) != 1 || w[0].GetId() != "brightness" || w[0].GetValue().GetInt() != 50 {
		t.Fatalf("write request: %v", fake.got)
	}

	fake.responses = []*clientsapi.ActionResponse{{Status: clientsapi.ActionResponse_SENT}}
	if err := rc.OnOff(testDevice, "1").Toggle(context.Background()); !errors.Is(err, ErrNoFinalStatus) {
		t.Fatalf("no final status: %v", err)
	}
}
