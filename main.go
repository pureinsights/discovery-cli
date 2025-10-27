package main

import (
	"os"

	"github.com/pureinsights/pdp-cli/cmd"
)

func main() {
	exitCode, _ := cmd.Run()
	os.Exit(int(exitCode))
}
