package credentials

import (
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewCredentialCommand creates the credential command.
func NewCredentialCommand(d cli.Discovery) *cobra.Command {
	credential := &cobra.Command{
		Use:   "credential [subcommand] [flags]",
		Short: "The command to interact with Discovery Core's credentials.",
	}

	credential.AddCommand(NewGetCommand(d))
	credential.AddCommand(NewDeleteCommand(d))

	return credential
}
