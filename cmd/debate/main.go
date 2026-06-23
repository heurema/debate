package main

import (
	"fmt"
	"os"
)

// Version is set at build time via -ldflags; defaults to dev build.
var Version = "dev"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: debate <command>")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "version":
		fmt.Println("debate", Version)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
