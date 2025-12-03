package endpoints

import (
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewEndpointCommand creates the endpoint command.
func NewEndpointCommand(d cli.Discovery) *cobra.Command {
	endpoint := &cobra.Command{
		Use:   "endpoint [subcommand] [flags]",
		Short: "The command to interact with Discovery QueryFlow's endpoints.",
	}

	endpoint.AddCommand(NewGetCommand(d))

	return endpoint
}
