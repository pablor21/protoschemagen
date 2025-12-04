package main

import (
	"os"

	"github.com/pablor21/protoschemagen/cmd"
)

func main() {
	// Check if "generate" subcommand is provided
	if len(os.Args) > 1 && os.Args[1] == "generate" {
		// Remove "generate" from args so flag parsing works correctly
		os.Args = append(os.Args[:1], os.Args[2:]...)
		cmd.Generate()
	} else if len(os.Args) > 1 && os.Args[1] == "generate-stubs" {
		// Remove "generate-stubs" from args so flag parsing works correctly
		os.Args = append(os.Args[:1], os.Args[2:]...)
		cmd.GenerateStubs()
	} else {
		// Default behavior - just generate
		cmd.Generate()
	}
}
