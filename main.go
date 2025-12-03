package main

import (
	"os"

	"github.com/pureinsights/discovery-cli/cmd"
)

func main() {
	exitCode, _ := cmd.Run()
	os.Exit(int(exitCode))
}
