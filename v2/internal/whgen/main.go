// Command whgen generates wh's typed services from the woodhouse-api service
// schema (github.com/jimjibone/woodhouse-api/schema). Run it with go generate.
//
//	whgen -kind bridge -out zz_services.go
package main

import (
	"flag"
	"fmt"
	"go/format"
	"os"

	"github.com/jimjibone/woodhouse-api/schema"
)

func main() {
	kind := flag.String("kind", "", "what to generate: bridge")
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
	case "bridge":
		src, err = bridge(s)
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
