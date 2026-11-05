package services

//go:generate go run ../../internal/whgen -kind bridge -out zz_services.go

import (
	"fmt"
	"slices"
)

// Helpers for the generated constructors in zz_services.go.

// checkFeatures panics on a feature the service type does not have, since
// that is a programming error.
func checkFeatures(service string, features []string, known ...string) {
	for _, f := range features {
		if !slices.Contains(known, f) {
			panic(fmt.Sprintf("services: %s has no feature %q", service, f))
		}
	}
}

func ptr[T any](v T) *T { return &v }
