package reactors

import (
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
)

// DefaultMaxEventAge is how old a replayed event may be and still reach a
// handler. Events missed during a reconnect are replayed by the server; acting
// on a stale button press is usually worse than missing it.
const DefaultMaxEventAge = 30 * time.Second

// EventInfo describes one occurrence of an event.
type EventInfo struct {
	// Time is when the bridge saw the event.
	Time time.Time
	// Replayed is true for an event sent from the server's buffer after a
	// reconnect rather than live.
	Replayed bool
	// Epoch and Seq are the server's position for the event.
	Epoch string
	Seq   uint64
}

// Age returns how long ago the event happened.
func (info EventInfo) Age() time.Duration { return time.Since(info.Time) }

func eventInfo(e *clientsapi.Event) EventInfo {
	return EventInfo{
		Time:     values.TimeFromPb(e.GetTime()),
		Replayed: e.GetReplayed(),
		Epoch:    e.GetEpoch(),
		Seq:      e.GetSeq(),
	}
}

// EventOption changes how events reach a handler.
type EventOption func(*eventHandler)

// MaxAge sets how old a replayed event may be and still reach the handler.
// Use a long age for events whose history matters (a doorbell, a pairing
// result); 0 drops every replayed event. Live events are always delivered.
func MaxAge(age time.Duration) EventOption {
	return func(h *eventHandler) { h.maxAge = age }
}

type eventHandler struct {
	id     string // event ID, "" for every event
	fn     func(*clientsapi.Event, EventInfo)
	maxAge time.Duration
}

// wants reports whether the handler takes the event.
func (h *eventHandler) wants(e *clientsapi.Event, info EventInfo, now time.Time) bool {
	if h.id != "" && h.id != e.GetId() {
		return false
	}
	return !info.Replayed || now.Sub(info.Time) <= h.maxAge
}
