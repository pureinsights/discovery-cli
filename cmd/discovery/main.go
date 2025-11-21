package main

import (
	"fmt"
	"os"

	"github.com/pureinsights/pdp-cli/cmd"
)

func main() {
	exitCode, err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred: %s", err)
	}
	os.Exit(int(exitCode))
}
