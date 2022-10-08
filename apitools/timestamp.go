package apitools

import (
	"time"

	api "github.com/jimjibone/woodhouse-4/api/go"
)

func TimeToTimestamp(t time.Time) *api.Timestamp {
	if t.IsZero() {
		return &api.Timestamp{}
	}
	return &api.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}

func TimestampToTime(t *api.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}
	if t.Seconds == 0 && t.Nanos == 0 {
		return time.Time{}
	}
	return time.Unix(t.Seconds, int64(t.Nanos))
}
