package services

//go:generate go run ../../internal/whgen -kind bridge -out zz_services.go

import (
	"fmt"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/schema"
	"github.com/jimjibone/woodhouse-api/values"
)

var loaded = schema.MustLoad()

// Schema returns the service schema the typed services are built from.
func Schema() *schema.Schema { return loaded }

func schemaService(name string) *schema.Service {
	srv := loaded.Service(name)
	if srv == nil {
		panic(fmt.Sprintf("services: schema has no service %q", name))
	}
	return srv
}

func durationOf(v *clientsapi.Value) time.Duration {
	d, _ := values.AsDuration(v)
	return d
}

func timeOf(v *clientsapi.Value) time.Time {
	t, _ := values.AsTime(v)
	return t
}
