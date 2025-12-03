package cli

import (
	"github.com/google/uuid"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
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
	GetEntity(client Getter, id uuid.UUID, printer Printer) error
	GetEntities(client Getter, printer Printer) error
	searchEntity(client Searcher, id string) (gjson.Result, error)
	SearchEntity(client Searcher, id string, printer Printer) error
	SearchEntities(client Searcher, filter gjson.Result, printer Printer) error
	UpsertEntities(client Creator, configurations gjson.Result, abortOnError bool, printer Printer) error
	DeleteEntity(client Deleter, id uuid.UUID, printer Printer) error
	SearchDeleteEntity(client SearchDeleter, name string, printer Printer) error
	StartSeed(client IngestionSeedController, name string, scanType discoveryPackage.ScanType, properties gjson.Result, printer Printer) error
	HaltSeed(client IngestionSeedController, name string, printer Printer) error
	HaltSeedExecution(client IngestionSeedExecutionController, execution uuid.UUID, printer Printer) error
	ExportEntitiesFromClient(client BackupRestore, path string, printer Printer) error
	ExportEntitiesFromClients(clients []BackupRestoreClientEntry, path string, printer Printer) error
	ImportEntitiesToClient(client BackupRestore, path string, onConflict discoveryPackage.OnConflict, printer Printer) error
	ImportEntitiesToClients(clients []BackupRestoreClientEntry, path string, onConflict discoveryPackage.OnConflict, printer Printer) error
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

// ConfigPath returns the path that contains Discovery's configuration.
func (d discovery) ConfigPath() string {
	return d.configPath
}
