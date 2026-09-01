package version

import (
	"fmt"
	"runtime/debug"

	"github.com/pureinsights/discovery-cli/v2/internal/cli"
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

// computeVersion updates the Version based on the LD Flags or debug.BuildInfo
func computeVersion(currentVersion string, buildInfo *debug.BuildInfo, ok bool) string {
	if currentVersion != "dev" && currentVersion != "" {
		return currentVersion
	}
	if !ok || buildInfo.Main.Version == "" || buildInfo.Main.Version == "(devel)" {
		return currentVersion
	}
	return buildInfo.Main.Version
}

// SetVersion sets the correct Version
func SetVersion() {
	info, ok := debug.ReadBuildInfo()
	Version = computeVersion(Version, info, ok)
}
