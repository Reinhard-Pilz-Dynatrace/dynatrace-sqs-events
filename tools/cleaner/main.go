//go:build ignore
// +build ignore

// Cross-platform cleaner: removes files if they exist (no error if missing).
// Usage: go run ./tools/cleaner file1 file2 ...
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		return
	}
	var hadErr bool
	for _, p := range os.Args[1:] {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "warn: remove %s: %v\n", p, err)
			hadErr = true
		}
	}
	if hadErr {
		os.Exit(1)
	}
}
