package services

//go:generate go run ../../internal/whgen -kind bridge -out zz_services.go

import (
	"fmt"

	"github.com/jimjibone/woodhouse-api/schema"
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
