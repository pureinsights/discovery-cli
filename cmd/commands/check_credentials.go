package commands

import (
	"strings"

	"github.com/pureinsights/pdp-cli/internal/cli"
)

// CommandConfig is a struct that contains fields necessary to check the credentials.
type commandConfig struct {
	profile       string
	output        string
	url           string
	apiKey        string
	componentName string
}

// GetCommandConfig is the constructor of the commandConfig struct
func GetCommandConfig(profile, output, componentName, url string) commandConfig {
	return commandConfig{
		profile:       profile,
		output:        output,
		url:           url,
		componentName: componentName,
	}
}

// CheckCredentials verifies that both the URL and API Key are set for the given profile and component in the configuration.
// If not, it returns an error
func CheckCredentials(d cli.Discovery, profile, componentName, urlProperty, apiProperty string) error {
	missingConfig := "The Discovery %[1]s %[2]s is missing for profile \"%[3]s\".\nTo set the %[2]s for the Discovery %[1]s API, run any of the following commands:\n      discovery config  --profile \"%[3]s\"\n      discovery %[4]s config --profile \"%[3]s\""

	vpr := d.Config()
	switch {
	case !vpr.IsSet(profile + "." + urlProperty):
		return cli.NewError(cli.ErrorExitCode, missingConfig, componentName, "URL", profile, strings.ToLower(componentName))
	}
	return nil
}
