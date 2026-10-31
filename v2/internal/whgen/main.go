// Command whgen generates wh's typed services from the woodhouse-api service
// schema (github.com/jimjibone/woodhouse-api/schema). Run it with go generate.
//
//	whgen -kind types   -out zz_types.go     (package types: params, results, events, structs)
//	whgen -kind bridge  -out zz_services.go  (package services: typed bridge services)
//	whgen -kind reactor -out zz_views.go     (package reactors: typed reactor views)
package main

import (
	"flag"
	"fmt"
	"go/format"
	"os"

	"github.com/jimjibone/woodhouse-api/schema"
)

func main() {
	kind := flag.String("kind", "", "what to generate: types, bridge or reactor")
	out := flag.String("out", "", "output file")
	flag.Parse()
	if *out == "" {
		fail("-out is required")
	}

	s, err := schema.Load()
	if err != nil {
		fail("load schema: %v", err)
	}

	var src []byte
	switch *kind {
	case "types":
		src, err = typesFile(s)
	case "bridge":
		src, err = bridge(s)
	case "reactor":
		src, err = reactor(s)
	default:
		fail("unknown -kind %q", *kind)
	}
	if err != nil {
		fail("generate: %v", err)
	}

	formatted, err := format.Source(src)
	if err != nil {
		// Write the unformatted source so the error's line numbers can be found.
		_ = os.WriteFile(*out, src, 0o644)
		fail("format %s: %v", *out, err)
	}
	if err := os.WriteFile(*out, formatted, 0o644); err != nil {
		fail("write: %v", err)
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "whgen: "+format+"\n", args...)
	os.Exit(1)
}
