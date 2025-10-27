package cli

import (
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
	SaveCoreConfigFromUser(profile string) error
	SaveIngestionConfigFromUser(profile string) error
	SaveQueryFlowConfigFromUser(profile string) error
	SaveStagingConfigFromUser(profile string) error
	PrintConfigToUser(profile string, sensitive bool) error
	PrintCoreConfigToUser(profile string, sensitive bool) error
	PrintIngestionConfigToUser(profile string, sensitive bool) error
	PrintQueryFlowConfigToUser(profile string, sensitive bool) error
	PrintStagingConfigToUser(profile string, sensitive bool) error
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

// ConfigPath returns the address that contains Discovery's configuration.
func (d discovery) ConfigPath() string {
	return d.configPath
}
