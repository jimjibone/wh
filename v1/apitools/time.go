package apitools

import (
	"time"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

func AttributeToTime(t *clientsapi.TimeAttribute) time.Time {
	return time.Unix(t.GetSeconds(), int64(t.GetNanos()))
}

func ValueToTime(t *clientsapi.TimeValue) time.Time {
	return time.Unix(t.GetSeconds(), int64(t.GetNanos()))
}

func TimeToAttribute(t time.Time, p clientsapi.Permissions) *clientsapi.TimeAttribute {
	secs, nanos := t.Unix(), int32(t.Nanosecond())
	if t.IsZero() {
		secs, nanos = 0, 0
	}
	return &clientsapi.TimeAttribute{
		Seconds: secs,
		Nanos:   nanos,
		Perms:   p,
	}
}

func TimeToValue(t time.Time) *clientsapi.TimeValue {
	secs, nanos := t.Unix(), int32(t.Nanosecond())
	if t.IsZero() {
		secs, nanos = 0, 0
	}
	return &clientsapi.TimeValue{
		Seconds: secs,
		Nanos:   nanos,
	}
}
