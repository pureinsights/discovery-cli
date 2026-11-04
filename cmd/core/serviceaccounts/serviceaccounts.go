package serviceaccounts

import (
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/spf13/cobra"
)

// NewServiceAccountCommand creates the service account command.
func NewServiceAccountCommand(d cli.Discovery) *cobra.Command {
	serviceAccount := &cobra.Command{
		Use:   "service-account [subcommand] [flags]",
		Short: "The command to interact with Discovery Core's service accounts.",
	}

	serviceAccount.AddCommand(NewGetCommand(d))
	serviceAccount.AddCommand(NewCreateCommand(d))
	serviceAccount.AddCommand(NewUpdateCommand(d))
	serviceAccount.AddCommand(NewEnableCommand(d))
	serviceAccount.AddCommand(NewDisableCommand(d))
	serviceAccount.AddCommand(NewDeleteCommand(d))

	return serviceAccount
}
