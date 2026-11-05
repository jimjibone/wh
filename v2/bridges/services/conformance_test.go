package services

import (
	"fmt"
	"testing"

	"github.com/jimjibone/wh/v2/bridges/attributes"
	"github.com/jimjibone/wh/v2/types"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/schema"
	"github.com/jimjibone/woodhouse-api/schema/schemaspec"
	"google.golang.org/protobuf/proto"
)

// narrowedPerms lists schema rw attributes that a generated service exposes
// read only until enabled (see Info.EnableRename).
var narrowedPerms = map[string]bool{"info.name": true}

// The generated services are compiled from the schema; this checks them
// against the schema of the woodhouse-api version in go.mod, for no feature,
// each single feature and all features. It fails when the api is upgraded
// without running "task generate".
func TestGeneratedServicesMatchSchema(t *testing.T) {
	s := schema.MustLoad()
	if types.SchemaVersion != s.Version {
		t.Fatalf("generated from schema version %d, go.mod's woodhouse-api has %d: run task generate", types.SchemaVersion, s.Version)
	}

	for _, srv := range s.Services {
		if srv.Open {
			continue
		}
		ctor, ok := constructors[srv.Name]
		if !ok {
			t.Errorf("%s: no generated service", srv.Name)
			continue
		}
		sets := [][]string{nil}
		var all []string
		for _, f := range srv.Features {
			sets = append(sets, []string{f.Name})
			all = append(all, f.Name)
		}
		if len(all) > 1 {
			sets = append(sets, all)
		}
		for _, features := range sets {
			t.Run(fmt.Sprintf("%s%v", srv.Name, features), func(t *testing.T) {
				checkService(t, srv, ctor("", features...), features)
			})
		}
	}
	if len(constructors) != len(s.Services)-1 { // less the open generic service
		t.Errorf("%d generated services, schema has %d", len(constructors), len(s.Services)-1)
	}
}

func checkService(t *testing.T, srv *schema.Service, got Service, features []string) {
	pb := got.Pb()
	if pb.Id != srv.Name || pb.Typ.String() != srv.ProtoName() {
		t.Errorf("id %q typ %s", pb.Id, pb.Typ)
	}
	if fmt.Sprint(pb.Features) != fmt.Sprint(features) && !(len(pb.Features) == 0 && len(features) == 0) {
		t.Errorf("features %v, want %v", pb.Features, features)
	}

	// Attributes: the enabled ones, in schema order, with the schema's spec,
	// perms and presence.
	attrs := got.(interface{ Attributes() []attributes.Attribute }).Attributes()
	var want []*schema.Field
	for _, a := range srv.Attributes {
		if schemaspec.Enabled(a.Features, features) {
			want = append(want, a)
		}
	}
	if len(attrs) != len(want) {
		t.Fatalf("%d attributes, want %d", len(attrs), len(want))
	}
	for i, a := range want {
		gotAttr := attrs[i]
		if gotAttr.ID() != a.Name {
			t.Errorf("attribute %d is %q, want %q", i, gotAttr.ID(), a.Name)
			continue
		}
		if wantSpec := schemaspec.SpecForFeatures(a, features); !proto.Equal(gotAttr.Pb().GetSpec(), wantSpec) {
			t.Errorf("%s spec %v, want %v", a.Name, gotAttr.Pb().GetSpec(), wantSpec)
		}
		wantPerms := schemaspec.Perms(a.Perms)
		if narrowedPerms[srv.Name+"."+a.Name] {
			wantPerms = clientsapi.Permissions_PERM_READONLY
		}
		if gotAttr.Perms() != wantPerms {
			t.Errorf("%s perms %s, want %s", a.Name, gotAttr.Perms(), wantPerms)
		}
		if (gotAttr.Optional() == attributes.Required) != a.Required {
			t.Errorf("%s required %t, want %t", a.Name, gotAttr.Optional() == attributes.Required, a.Required)
		}
	}

	// Commands and events: the enabled ones with the schema's specs.
	var wantCmds []*clientsapi.CommandSpec
	for _, c := range srv.Commands {
		if schemaspec.Enabled(c.Features, features) {
			wantCmds = append(wantCmds, schemaspec.CommandSpecFromSchema(c, features))
		}
	}
	if !specsEqual(pb.Commands, wantCmds) {
		t.Errorf("commands %v, want %v", pb.Commands, wantCmds)
	}
	var wantEvents []*clientsapi.EventSpec
	for _, e := range srv.Events {
		if schemaspec.Enabled(e.Features, features) {
			wantEvents = append(wantEvents, schemaspec.EventSpecFromSchema(e, features))
		}
	}
	if !specsEqual(pb.Events, wantEvents) {
		t.Errorf("events %v, want %v", pb.Events, wantEvents)
	}
}

func specsEqual[T proto.Message](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !proto.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}
