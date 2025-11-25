package cmd

import (
	"os"

	"github.com/pureinsights/pdp-cli/cmd/backuprestore"
	"github.com/pureinsights/pdp-cli/cmd/config"
	"github.com/pureinsights/pdp-cli/cmd/core"
	"github.com/pureinsights/pdp-cli/cmd/ingestion"
	"github.com/pureinsights/pdp-cli/cmd/queryflow"
	"github.com/pureinsights/pdp-cli/cmd/staging"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/cobra"
)

// NewRootCommand creates and sets up the root command of the Discovery CLI
func newRootCommand(d cli.Discovery) *cobra.Command {
	discovery := &cobra.Command{
		Use:   "discovery [subcommand]",
		Short: "A CLI to assist with operations on Pureinsights Discovery",
		Long:  "discovery is the Discovery CLI's root command. This is the command used to run the CLI. It contains all of the other subcommands.",
	}

	ios := d.IOStreams()

	discovery.SetIn(ios.In)
	discovery.SetOut(ios.Out)
	discovery.SetErr(ios.Err)

	discovery.PersistentFlags().StringP(
		"profile",
		"p",
		d.Config().GetString("profile"),
		"configuration profile to use",
	)

	d.Config().BindPFlag("profile", discovery.PersistentFlags().Lookup("profile"))

	discovery.AddCommand(config.NewConfigCommand(d))
	discovery.AddCommand(backuprestore.NewExportCommand(d))
	discovery.AddCommand(backuprestore.NewImportCommand(d))
	discovery.AddCommand(core.NewCoreCommand(d))
	discovery.AddCommand(ingestion.NewIngestionCommand(d))
	discovery.AddCommand(queryflow.NewQueryFlowCommand(d))
	discovery.AddCommand(staging.NewStagingCommand(d))

	return discovery
}

// Run executes the Root command
func Run() (cli.ExitCode, error) {
	ios := iostreams.IOStreams{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}

	configPath, err := cli.SetDiscoveryDir()
	if err != nil {
		cliError := cli.FromError(err)
		return cliError.ExitCode, cliError
	}

	viper, err := cli.InitializeConfig(ios, configPath)
	if err != nil {
		cliError := cli.FromError(err)
		return cliError.ExitCode, cliError
	}
	d := cli.NewDiscovery(&ios, viper, configPath)
	root := newRootCommand(d)
	err = root.Execute()
	if err != nil {
		cliError := cli.FromError(err)
		return cliError.ExitCode, cliError
	}
	return cli.SuccessExitCode, nil
}
