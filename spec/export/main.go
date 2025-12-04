package main

import (
	"encoding/json"
	"os"

	"github.com/pablor21/protoschemagen/spec"
)

// This utility generates the Protobuf format generator specs JSON file
func main() {
	fname := "../../specs.json"
	// open the file and write the specs
	b, err := json.MarshalIndent(spec.Specs, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(fname, b, 0644); err != nil {
		panic(err)
	}
}
