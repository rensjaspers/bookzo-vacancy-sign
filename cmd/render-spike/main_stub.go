//go:build !sdl

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "build with -tags sdl to run the native render spike")
	os.Exit(1)
}
