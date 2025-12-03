package main

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/pureinsights/discovery-cli/cmd"
	"github.com/pureinsights/discovery-cli/internal/cli"
)

func main() {
	exitCode, err := cmd.Run()
	if err != nil {
		var cliError *cli.Error
		if errors.As(err, &cliError) && slices.Contains([]string{"Could not access the user's Home directory", "Could not create the /.discovery directory", "Could not read the configuration file", "Could not read the credentials file"}, cliError.Message) {
			fmt.Fprintf(os.Stderr, "An error occurred: %s", err)
		}
	}
	os.Exit(int(exitCode))
}
