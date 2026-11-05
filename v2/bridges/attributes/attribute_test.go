package attributes

import (
	"errors"
	"testing"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
)

func capture(attr Attribute) *[]*clientsapi.Attribute {
	var got []*clientsapi.Attribute
	attr.Push(func(a *clientsapi.Attribute) { got = append(got, a) })
	return &got
}

func TestSetPushesValueOnlyOnChange(t *testing.T) {
	attr := NewInt("level", clientsapi.Permissions_PERM_READONLY, Required, nil)
	got := capture(attr)

	if !attr.Set(5) || attr.Set(5) {
		t.Fatal("Set should report a change once")
	}
	if len(*got) != 1 {
		t.Fatalf("pushes = %d, want 1", len(*got))
	}
	up := (*got)[0]
	if up.Spec != nil || up.Time != nil || up.Value.GetInt() != 5 {
		t.Fatalf("partial update should carry only the value: %v", up)
	}
}

func TestSetAtCarriesObservationTime(t *testing.T) {
	attr := NewBool("on", clientsapi.Permissions_PERM_READONLY, Required)
	got := capture(attr)
	at := time.Unix(1_800_000_000, 0)

	attr.SetAt(true, at)
	if !values.TimeFromPb((*got)[0].Time).Equal(at) || !attr.Time().Equal(at) {
		t.Fatalf("time not carried: %v", (*got)[0])
	}
	attr.Set(false)
	if (*got)[1].Time != nil || !attr.Time().IsZero() {
		t.Fatal("Set should clear the observation time so the server stamps it")
	}
}

func TestSpecChangePushesWholeAttribute(t *testing.T) {
	attr := NewEnum("mode", clientsapi.Permissions_PERM_READWRITE, Required, "a", "b")
	got := capture(attr)
	attr.Set("a")
	if attr.SetOptions([]string{"a", "b"}) {
		t.Fatal("unchanged options reported as changed")
	}
	attr.SetOptions([]string{"a", "b", "c"})
	if len(*got) != 2 {
		t.Fatalf("pushes = %d, want 2", len(*got))
	}
	full := (*got)[1]
	if full.Perms != clientsapi.Permissions_PERM_READWRITE || len(full.Spec.GetEnum().Options) != 3 || full.Value.GetEnum() != "a" {
		t.Fatalf("spec update should carry the whole attribute: %v", full)
	}
}

func TestHandleWrite(t *testing.T) {
	ro := NewBool("on", clientsapi.Permissions_PERM_READONLY, Required)
	if err := ro.HandleWrite(values.Bool(true)); !errors.Is(err, ErrReadOnly) {
		t.Fatalf("read only: %v", err)
	}

	min, max := int64(0), int64(100)
	rw := NewInt("brightness", clientsapi.Permissions_PERM_READWRITE, Optional, &clientsapi.IntSpec{Min: &min, Max: &max})
	if err := rw.HandleWrite(values.Int(5)); !errors.Is(err, ErrNoHandler) {
		t.Fatalf("no handler: %v", err)
	}
	var written int64
	rw.OnWrite(func(v int64) error { written = v; return nil })
	if err := rw.HandleWrite(values.Int(101)); err == nil {
		t.Fatal("out of range write accepted")
	}
	if err := rw.HandleWrite(values.Text("x")); err == nil {
		t.Fatal("wrong kind accepted")
	}
	if err := rw.HandleWrite(values.Int(42)); err != nil || written != 42 {
		t.Fatalf("valid write: %v, written %d", err, written)
	}
}

func TestOptionalOnlyWhenSet(t *testing.T) {
	attr := NewText("model", clientsapi.Permissions_PERM_READONLY, Optional)
	if attr.IsSet() {
		t.Fatal("new attribute should not be set")
	}
	if attr.Pb().Value != nil {
		t.Fatal("unset value should be nil")
	}
	capture(attr)
	attr.Set("x")
	if !attr.IsSet() {
		t.Fatal("should be set")
	}
}
