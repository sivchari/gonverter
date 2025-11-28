// Package main provides the gonverter CLI command.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sivchari/gonverter/internal/gonverter"
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gonverter <package>")
		os.Exit(1)
	}

	if err := gonverter.Run(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
