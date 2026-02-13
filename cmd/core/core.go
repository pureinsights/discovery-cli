package core

import (
	"github.com/pureinsights/discovery-cli/cmd/core/backuprestore"
	"github.com/pureinsights/discovery-cli/cmd/core/config"
	"github.com/pureinsights/discovery-cli/cmd/core/credentials"
	"github.com/pureinsights/discovery-cli/cmd/core/labels"
	"github.com/pureinsights/discovery-cli/cmd/core/secrets"
	"github.com/pureinsights/discovery-cli/cmd/core/servers"
	"github.com/pureinsights/discovery-cli/cmd/core/statuscheck"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewCoreCommand creates the core command.
func NewCoreCommand(d cli.Discovery) *cobra.Command {
	core := &cobra.Command{
		Use:   "core [subcommand] [flags]",
		Short: "The main command to interact with Discovery's Core",
	}

	core.AddCommand(config.NewConfigCommand(d))
	core.AddCommand(labels.NewLabelCommand(d))
	core.AddCommand(secrets.NewSecretCommand(d))
	core.AddCommand(credentials.NewCredentialCommand(d))
	core.AddCommand(servers.NewServerCommand(d))
	core.AddCommand(backuprestore.NewExportCommand(d))
	core.AddCommand(backuprestore.NewImportCommand(d))
	core.AddCommand(statuscheck.NewStatusCommand(d))

	return core
}
