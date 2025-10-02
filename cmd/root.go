package cmd

import (
	"os"

	"github.com/pureinsights/pdp-cli/cmd/config"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/cobra"
)

// NewRootCommand creates and sets up the root command of the Discovery CLI
func newRootCommand(d cli.Discovery) *cobra.Command {
	discovery := &cobra.Command{
		Use:   "discovery [subcommand]",
		Short: "A CLI to assist with operations on Pureinsights Discovery",
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

	return discovery
}

// Run executes the root command
func Run() (cli.ExitCode, error) {
	ios := iostreams.IOStreams{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}

	configPath, err := cli.SetDiscoveryDir()
	if err != nil {
		return cli.ErrorExitCode, cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not set up Discovery's directory in User's home directory")
	}

	viper, err := cli.InitializeConfig(ios, configPath)
	if err != nil {
		return cli.ErrorExitCode, cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not initialize configuration")
	}
	d := cli.NewDiscovery(&ios, viper, configPath)
	root := newRootCommand(d)
	err = root.Execute()
	if err != nil {
		return cli.ErrorExitCode, cli.FromError(err)
	}
	return cli.SuccessExitCode, nil
}
