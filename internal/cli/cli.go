package cli

import (
	"strings"

	"github.com/google/uuid"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/viper"
)

// Discovery is the interface for the struct that represents Discovery.
// This interface allows mocking Discovery in tests.
type Discovery interface {
	IOStreams() *iostreams.IOStreams
	Config() *viper.Viper
	ConfigPath() string
	SaveConfigFromUser(profile string) error
	SaveCoreConfigFromUser(profile string, standalone bool) error
	SaveIngestionConfigFromUser(profile string, standalone bool) error
	SaveQueryFlowConfigFromUser(profile string, standalone bool) error
	SaveStagingConfigFromUser(profile string, standalone bool) error
	PrintConfigToUser(profile string, sensitive bool) error
	PrintCoreConfigToUser(profile string, sensitive, standalone bool) error
	PrintIngestionConfigToUser(profile string, sensitive, standalone bool) error
	PrintQueryFlowConfigToUser(profile string, sensitive, standalone bool) error
	PrintStagingConfigToUser(profile string, sensitive, standalone bool) error
	GetEntity(client getter, id uuid.UUID, printer Printer) error
	GetEntities(client getter, printer Printer) error
}

// Discovery is the struct that has the implementation of Discovery's CLI.
type discovery struct {
	config     *viper.Viper
	configPath string
	iostreams  *iostreams.IOStreams
}

// IOStreams is a getter method to obtain the CLI's IO streams.
func (d discovery) IOStreams() *iostreams.IOStreams {
	return d.iostreams
}

// Config is a getter method to get Discovery's Viper configuration.
func (d discovery) Config() *viper.Viper {
	return d.config
}

// NewDiscovery is a constructor of the discovery struct.
func NewDiscovery(io *iostreams.IOStreams, config *viper.Viper, configPath string) discovery {
	return discovery{
		config:     config,
		iostreams:  io,
		configPath: configPath,
	}
}

func (d discovery) ConfigPath() string {
	return d.configPath
}

func checkCredentials(d Discovery, profile, componentName, urlProperty, apiProperty string) error {
	missingConfig := "The Discovery %s %s is missing for profile %q.\nTo set the %s for the Discovery %s API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery %s config --profile {profile}"

	vpr := d.Config()
	switch {
	case !vpr.IsSet(profile + "." + urlProperty):
		return NewError(ErrorExitCode, missingConfig, componentName, "URL", profile, "URL", componentName, strings.ToLower(componentName))
	case !vpr.IsSet(profile + "." + apiProperty):
		return NewError(ErrorExitCode, missingConfig, componentName, "API key", profile, "API key", componentName, strings.ToLower(componentName))
	}
	return nil
}
