package services

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jimjibone/wh/v2/bridges/attributes"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
)

type recorder struct {
	updates []*clientsapi.Service
	events  []*clientsapi.Event
}

func record(srv Service) *recorder {
	r := &recorder{}
	srv.Push(
		func(s *clientsapi.Service) { r.updates = append(r.updates, s) },
		func(e *clientsapi.Event) { r.events = append(r.events, e) },
	)
	return r
}

func write(srv Service, fields ...*clientsapi.FieldValue) error {
	_, err := srv.HandleAction(&clientsapi.ActionRequest{
		Action: &clientsapi.ActionRequest_Write{Write: &clientsapi.WriteAction{Values: fields}},
	}, nil)
	return err
}

func invoke(srv Service, id string, params ...*clientsapi.FieldValue) ([]*clientsapi.FieldValue, error) {
	return srv.HandleAction(&clientsapi.ActionRequest{
		Action: &clientsapi.ActionRequest_Command{Command: &clientsapi.CommandAction{Id: id, Params: params}},
	}, nil)
}

// Every generated service has the schema's service type, sends its required
// attributes and declares the schema's commands and events.
func TestGeneratedServicesMatchSchema(t *testing.T) {
	built := map[string]Service{
		"info": NewInfo(""), "online": NewOnline(""), "on_off": NewOnOff(""),
		"light": NewLight("", LightFeatureBrightness, LightFeatureColorTemp, LightFeatureHS, LightFeatureXY),
		"thermostat": NewThermostat("", ThermostatFeatureSchedule), "button": NewButton(""),
		"pairing": NewPairing("", PairingFeatureTimedWindow, PairingFeatureCodeEntry),
	}
	for name, srv := range built {
		s := Schema().Service(name)
		pb := srv.Pb()
		if pb.Typ.String() != s.ProtoName() || pb.Id != name {
			t.Errorf("%s: typ %s id %q", name, pb.Typ, pb.Id)
		}
		for _, a := range s.Attributes {
			if !a.Required {
				continue
			}
			found := false
			for _, got := range pb.Attrs {
				found = found || got.Id == a.Name
			}
			if !found {
				t.Errorf("%s: required attribute %q not sent", name, a.Name)
			}
		}
		if len(pb.Commands) != len(s.Commands) || len(pb.Events) != len(s.Events) {
			t.Errorf("%s: %d commands, %d events; schema has %d, %d", name, len(pb.Commands), len(pb.Events), len(s.Commands), len(s.Events))
		}
	}
}

func TestFeaturesGateAttributesAndCommands(t *testing.T) {
	plain := NewLight("")
	if plain.Brightness != nil || plain.Color != nil {
		t.Fatal("feature attributes should be nil without their features")
	}
	cover := NewCover("")
	if cover.CommandSpec("go_to") != nil {
		t.Fatal("go_to needs feature position")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("OnGoTo without feature position should panic")
		}
	}()
	cover.OnGoTo(func(CoverGoToParams) error { return nil })
}

func TestUnknownFeaturePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("unknown feature should panic")
		}
	}()
	NewLight("", "dimmable")
}

func TestTypedCommand(t *testing.T) {
	light := NewLight("", LightFeatureBrightness)
	record(light)
	var got LightSetParams
	light.OnSet(func(p LightSetParams) error { got = p; return nil })

	_, err := invoke(light, "set", values.Field("brightness", values.Int(40)), values.Field("transition", values.Duration(3*time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	if got.Brightness == nil || *got.Brightness != 40 || got.Transition == nil || *got.Transition != 3*time.Second || got.On != nil {
		t.Fatalf("params not parsed: %+v", got)
	}

	_, err = invoke(light, "set", values.Field("brightness", values.Int(400)))
	var invalid *InvalidError
	if !errors.As(err, &invalid) {
		t.Fatalf("out of range brightness should be invalid: %v", err)
	}
	if _, err = invoke(light, "set", values.Field("color_temp", values.Int(3000))); !errors.As(err, &invalid) {
		t.Fatalf("color_temp without its feature should be invalid: %v", err)
	}
	if _, err = invoke(light, "nope"); !errors.Is(err, ErrCommandNotFound) {
		t.Fatalf("unknown command: %v", err)
	}
}

func TestCommandResult(t *testing.T) {
	pairing := NewPairing("", PairingFeatureTimedWindow)
	record(pairing)
	ends := time.Unix(1_800_000_000, 0)
	pairing.OnStart(func(p PairingStartParams) (PairingStartResult, error) {
		if p.Duration == nil || *p.Duration != time.Minute {
			t.Errorf("duration param: %+v", p)
		}
		return PairingStartResult{Ends: &ends}, nil
	})
	result, err := invoke(pairing, "start", values.Field("duration", values.Duration(time.Minute)))
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := values.AsTime(values.Fields(result)["ends"]); !got.Equal(ends) {
		t.Fatalf("result: %v", result)
	}
}

func TestStructuredAttributeAndCommand(t *testing.T) {
	th := NewThermostat("", ThermostatFeatureSchedule)
	r := record(th)
	sched := []ThermostatScheduleItem{
		{Days: []string{"mon", "tue"}, Start: 7 * time.Hour, Setpoint: 20.5},
		{Days: []string{"sat"}, Start: 9 * time.Hour, Setpoint: 19},
	}
	if !th.SetSchedule(sched) || th.SetSchedule(sched) {
		t.Fatal("SetSchedule should report one change")
	}
	got := th.ScheduleItems()
	if len(got) != 2 || got[0].Days[1] != "tue" || got[1].Start != 9*time.Hour {
		t.Fatalf("round trip: %+v", got)
	}
	if len(r.updates) != 1 {
		t.Fatalf("updates = %d", len(r.updates))
	}

	var received []ThermostatScheduleItem
	th.OnSetSchedule(func(p ThermostatSetScheduleParams) error { received = p.Schedule; return nil })
	if _, err := invoke(th, "set_schedule", values.Field("schedule", values.List(sched[0].value()))); err != nil {
		t.Fatal(err)
	}
	if len(received) != 1 || received[0].Setpoint != 20.5 {
		t.Fatalf("command params: %+v", received)
	}
}

func TestEvents(t *testing.T) {
	button := NewButton("")
	r := record(button)
	if err := button.EmitPress(); err != nil {
		t.Fatal(err)
	}
	if err := button.EmitMultiPress(ButtonMultiPressEvent{Count: 2}); err != nil {
		t.Fatal(err)
	}
	if err := button.EmitMultiPress(ButtonMultiPressEvent{Count: 0}); err == nil {
		t.Fatal("count below min should be rejected")
	}
	if len(r.events) != 2 || r.events[0].Id != "press" || r.events[1].Fields[0].Value.GetInt() != 2 || r.events[1].Time == nil {
		t.Fatalf("events: %v", r.events)
	}
	if err := button.Emit("nope"); !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("unknown event: %v", err)
	}
}

func TestWrites(t *testing.T) {
	sw := NewOnOff("")
	record(sw)
	var on bool
	sw.On.OnWrite(func(v bool) error { on = v; return nil })
	if err := write(sw, values.Field("on", values.Bool(true))); err != nil || !on {
		t.Fatalf("write: %v", err)
	}
	var invalid *InvalidError
	if err := write(sw, values.Field("nope", values.Bool(true))); !errors.As(err, &invalid) {
		t.Fatalf("unknown attribute: %v", err)
	}
	if _, err := sw.HandleAction(&clientsapi.ActionRequest{}, nil); !errors.As(err, &invalid) {
		t.Fatalf("empty action: %v", err)
	}
}

// A multi-attribute write changes nothing if any value is bad.
func TestWriteIsAllOrNothing(t *testing.T) {
	light := NewLight("", LightFeatureBrightness)
	record(light)
	var writes []string
	light.On.OnWrite(func(bool) error { writes = append(writes, "on"); return nil })
	light.Brightness.OnWrite(func(int64) error { writes = append(writes, "brightness"); return nil })
	err := write(light, values.Field("on", values.Bool(true)), values.Field("brightness", values.Int(500)))
	if err == nil || len(writes) != 0 {
		t.Fatalf("err %v, writes %v", err, writes)
	}
}

func TestInfoRename(t *testing.T) {
	info := NewInfo("")
	record(info)
	if info.Name.Perms() != clientsapi.Permissions_PERM_READONLY {
		t.Fatal("name should be read only by default")
	}
	if err := write(info, values.Field("name", values.Text("x"))); !errors.Is(err, attributes.ErrReadOnly) {
		t.Fatalf("write without rename: %v", err)
	}

	var got string
	info.EnableRename(func(n string) error { got = n; return nil })
	if info.Pb().Attrs[0].Perms != clientsapi.Permissions_PERM_READWRITE {
		t.Fatal("name should be writable after EnableRename")
	}
	if err := write(info, values.Field("name", values.Text("  New Name  "))); err != nil || got != "New Name" {
		t.Fatalf("rename: %v %q", err, got)
	}
	if err := write(info, values.Field("name", values.Text("   "))); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("empty name: %v", err)
	}
	if err := write(info, values.Field("model", values.Text("x"))); !errors.Is(err, attributes.ErrReadOnly) {
		t.Fatalf("other attributes stay read only: %v", err)
	}
}
