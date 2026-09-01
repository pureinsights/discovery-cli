package version

import (
	"fmt"
	"runtime/debug"

	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"
)

// NewVersionCommand creates the discovery version command that gets the status of every Discovery product.
func NewVersionCommand(d cli.Discovery) *cobra.Command {
	status := &cobra.Command{
		Use:   "version",
		Short: "Prints the current version of the Discovery CLI",
		Long:  "version prints the current version of the Discovery CLI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(d.IOStreams().Out, "Discovery CLI Version %s\n", Version)
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not print CLI version")
			}

			return nil
		},
		Args: cobra.NoArgs,
		Example: `	# Print the version of the Discovery CLI
	discovery version`,
	}

	return status
}

func init() {
	// An ldflags release build already supplied the version.
	if Version != "dev" && Version != "" {
		return
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
}
