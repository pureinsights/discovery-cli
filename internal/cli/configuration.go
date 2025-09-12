package cli

import (
	"fmt"
	"path/filepath"

	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/viper"
)

// ReadConfigFile is an auxiliary function that is used to read the configuration values in the file located at the given path.
// When the file could not be found, an error is logged to the error stream of the IOStreams parameter.
func readConfigFile(baseName, path string, v *viper.Viper, ios *iostreams.IOStreams) error {
	v.SetConfigName(baseName)
	v.SetConfigType("toml")
	v.AddConfigPath(path)

	if err := v.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Fprintf(ios.Err,
				"Configuration file %q not found under %q; using default values.\n",
				baseName, filepath.Clean(path),
			)
			return nil
		}
		return fmt.Errorf("could not read %q from %q: %w", baseName, filepath.Clean(path), err)
	}
	return nil
}

// InitializeConfig reads the config and credentials configuration files found in the given path and sets up the Viper instance with their values.
func InitializeConfig(ios iostreams.IOStreams, path string) (*viper.Viper, error) {
	vpr := viper.New()

	vpr.SetDefault("profile", "default")
	vpr.SetDefault("default.core_url", "http://localhost:8080")
	vpr.SetDefault("default.ingestion_url", "http://localhost:8080")
	vpr.SetDefault("default.queryflow_url", "http://localhost:8088")
	vpr.SetDefault("default.staging_url", "http://localhost:8081")
	vpr.SetDefault("default.core_key", "")
	vpr.SetDefault("default.ingestion_key", "")
	vpr.SetDefault("default.queryflow_key", "")
	vpr.SetDefault("default.staging_key", "")

	if err := readConfigFile("config", path, vpr, &ios); err != nil {
		return nil, err
	}
	if err := readConfigFile("credentials", path, vpr, &ios); err != nil {
		return nil, err
	}

	return vpr, nil
}
