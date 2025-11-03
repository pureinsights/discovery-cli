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
func GetCommandConfig(profile, output, componentName, url, apiKey string) commandConfig {
	return commandConfig{
		profile:       profile,
		output:        output,
		url:           url,
		apiKey:        apiKey,
		componentName: componentName,
	}
}

// CheckCredentials verifies that both the URL and API Key are set for the given profile and component in the configuration.
// If not, it returns an error
func CheckCredentials(d cli.Discovery, profile, componentName, urlProperty, apiProperty string) error {
	missingConfig := "The Discovery %s %s is missing for profile %q.\nTo set the %s for the Discovery %s API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery %s config --profile {profile}"

	vpr := d.Config()
	switch {
	case !vpr.IsSet(profile + "." + urlProperty):
		return cli.NewError(cli.ErrorExitCode, missingConfig, componentName, "URL", profile, "URL", componentName, strings.ToLower(componentName))
	case !vpr.IsSet(profile + "." + apiProperty):
		return cli.NewError(cli.ErrorExitCode, missingConfig, componentName, "API key", profile, "API key", componentName, strings.ToLower(componentName))
	}
	return nil
}
