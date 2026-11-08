// Package attributes holds the typed attributes a bridge service exposes. Each
// attribute has a spec (type and constraints), a current value and the time
// the value was observed. Setting a value pushes a partial update to the
// service; changing the spec pushes the whole attribute.
//
// Attributes are safe for concurrent use. Updates are pushed while the
// attribute's lock is held, so they reach the service in the order they were
// made. Write handlers run without the lock and may set the attribute.
package attributes

import (
	"errors"
	"fmt"
	"sync"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
	"google.golang.org/protobuf/proto"
)

type OptionalType int

const (
	Optional OptionalType = iota
	Required
)

var (
	ErrReadOnly  = errors.New("write to read only attribute")
	ErrNoHandler = errors.New("attribute has no write handler")
)

type Attribute interface {
	ID() string
	Perms() clientsapi.Permissions
	Optional() OptionalType
	// IsSet reports whether a value has been set. Optional attributes are only
	// sent once set.
	IsSet() bool
	// Push registers the func the attribute calls with every update.
	Push(push func(*clientsapi.Attribute))
	// Pb returns the whole attribute: perms, spec, value and time.
	Pb() *clientsapi.Attribute
	// HandleWrite validates a written value against the spec and passes it to
	// the write handler. It fails for read only attributes.
	HandleWrite(v *clientsapi.Value) error
}

// base implements Attribute on raw API values. The typed attributes wrap it.
// Stored values and specs are never modified in place, only replaced, so they
// may be handed out after the lock is released.
type base struct {
	id string // immutable

	mu       sync.Mutex
	perms    clientsapi.Permissions
	optional OptionalType
	spec     *clientsapi.Spec
	value    *clientsapi.Value
	at       time.Time // observation time, zero means the server stamps it
	push     func(*clientsapi.Attribute)
	onWrite  func(*clientsapi.Value) error
}

func newBase(id string, perms clientsapi.Permissions, optional OptionalType, spec *clientsapi.Spec) base {
	if perms == clientsapi.Permissions_PERM_UNDEFINED {
		perms = clientsapi.Permissions_PERM_READONLY
	}
	return base{id: id, perms: perms, optional: optional, spec: spec}
}

func (b *base) ID() string { return b.id }

func (b *base) Perms() clientsapi.Permissions {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.perms
}

func (b *base) Optional() OptionalType {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.optional
}

func (b *base) IsSet() bool { return values.IsSet(b.load()) }

func (b *base) Push(push func(*clientsapi.Attribute)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.push = push
}

// SetOptional changes whether the attribute is sent before it has a value.
func (b *base) SetOptional(optional OptionalType) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.optional = optional
}

// SetPerms changes the permissions, e.g. to make a schema rw attribute read
// only when the device cannot be written. Changes are pushed.
func (b *base) SetPerms(perms clientsapi.Permissions) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.perms == perms {
		return
	}
	b.perms = perms
	if b.push != nil {
		b.push(b.pb())
	}
}

// Time returns when the current value was observed, or the zero time if the
// server stamps it.
func (b *base) Time() time.Time {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.at
}

// Spec returns the attribute's spec. Do not modify it; use the typed setters.
func (b *base) Spec() *clientsapi.Spec { return b.loadSpec() }

func (b *base) Pb() *clientsapi.Attribute {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.pb()
}

// pb must be called with the lock held.
func (b *base) pb() *clientsapi.Attribute {
	pb := &clientsapi.Attribute{
		Id:    b.id,
		Perms: b.perms,
		Spec:  b.spec,
		Value: b.value,
	}
	if !b.at.IsZero() {
		pb.Time = values.TimeToPb(b.at)
	}
	return pb
}

func (b *base) HandleWrite(v *clientsapi.Value) error {
	b.mu.Lock()
	perms, spec, onWrite := b.perms, b.spec, b.onWrite
	b.mu.Unlock()

	if perms != clientsapi.Permissions_PERM_READWRITE {
		return ErrReadOnly
	}
	if err := values.Validate(spec, v); err != nil {
		return fmt.Errorf("%s: %w", b.id, err)
	}
	if onWrite == nil {
		return ErrNoHandler
	}
	return onWrite(v)
}

// load returns the current value. Do not modify it.
func (b *base) load() *clientsapi.Value {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.value
}

// loadSpec returns the current spec. Do not modify it.
func (b *base) loadSpec() *clientsapi.Spec {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.spec
}

func (b *base) setOnWrite(onWrite func(*clientsapi.Value) error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onWrite = onWrite
}

// set stores a value observed at t (zero: the server stamps it) and pushes a
// value-only update when it changed. Reports whether it changed.
func (b *base) set(v *clientsapi.Value, t time.Time) bool {
	return b.setUnless(nil, v, t)
}

// setUnless is set, but also skips the update when same reports the current
// value (only called once set) as the same as v. The check and the update are
// atomic.
func (b *base) setUnless(same func(old *clientsapi.Value) bool, v *clientsapi.Value, t time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if proto.Equal(b.value, v) || (same != nil && values.IsSet(b.value) && same(b.value)) {
		return false
	}
	b.value = v
	b.at = t
	update := &clientsapi.Attribute{Id: b.id, Value: v}
	if !t.IsZero() {
		update.Time = values.TimeToPb(t)
	}
	b.send(update)
	return true
}

// setSpec replaces the spec and pushes the whole attribute when it changed.
func (b *base) setSpec(spec *clientsapi.Spec) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if proto.Equal(b.spec, spec) {
		return false
	}
	b.spec = spec
	b.send(b.pb())
	return true
}

// send must be called with the lock held.
func (b *base) send(update *clientsapi.Attribute) {
	if b.push == nil {
		panic(fmt.Sprintf("attribute %q is not registered with a service", b.id))
	}
	b.push(update)
}
