// Package types holds the Go types generated from the woodhouse-api service
// schema: the structs of list and struct attributes, command params and
// results, and event fields, with their conversions to and from API values.
// The bridge services and the reactor views share them.
package types

//go:generate go run ../internal/whgen -kind types -out zz_types.go

import (
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
)

// DurationOf reads a duration value; other kinds read as 0.
func DurationOf(v *clientsapi.Value) time.Duration {
	d, _ := values.AsDuration(v)
	return d
}

// TimeOf reads a time value; other kinds read as the zero time.
func TimeOf(v *clientsapi.Value) time.Time {
	t, _ := values.AsTime(v)
	return t
}
