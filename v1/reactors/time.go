package reactors

import (
	"time"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

func timeFromPb(t *clientsapi.TimeAttribute) time.Time {
	if t == nil {
		return time.Time{}
	}
	if t.Seconds == 0 && t.Nanos == 0 {
		return time.Time{}
	}
	return time.Unix(t.Seconds, int64(t.Nanos))
}

func timeToPb(t time.Time, p clientsapi.Permissions) *clientsapi.TimeValue {
	secs, nanos := t.Unix(), int32(t.Nanosecond())
	if t.IsZero() {
		secs, nanos = 0, 0
	}
	return &clientsapi.TimeValue{
		Seconds: secs,
		Nanos:   nanos,
	}
}
