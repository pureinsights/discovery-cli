package cli

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/viper"
)

const (
	DefaultCoreURL      string = "http://localhost:8080"
	DefaultIngestionURL string = "http://localhost:8080"
	DefaultQueryFlowURL string = "http://localhost:8088"
	DefaultStagingURL   string = "http://localhost:8081"
)

// ReadConfigFile is an auxiliary function that is used to read the configuration values in the file located at the given path.
// When the file could not be found, an error is logged to the error stream of the IOStreams parameter.
func readConfigFile(baseName, path string, v *viper.Viper, ios *iostreams.IOStreams) (bool, error) {
	v.SetConfigName(baseName)
	v.SetConfigType("toml")
	v.AddConfigPath(path)

	if err := v.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Fprintf(ios.Err,
				"Configuration file %q not found under %q; using default values.\n",
				baseName, filepath.Clean(path),
			)
			return false, nil
		}
		return true, fmt.Errorf("could not read %q from %q: %w", baseName, filepath.Clean(path), err)
	}
	return true, nil
}

// InitializeConfig reads the config and credentials configuration files found in the given path and sets up the Viper instance with their values.
func InitializeConfig(ios iostreams.IOStreams, path string) (*viper.Viper, error) {
	vpr := viper.New()

	vpr.SetDefault("profile", "default")
	defaultProfile := "default"

	if exists, err := readConfigFile("config", path, vpr, &ios); err != nil {
		return nil, err
	} else {
		if !exists {
			vpr.SetDefault(fmt.Sprintf("%s.core_url", defaultProfile), DefaultCoreURL)
			vpr.SetDefault(fmt.Sprintf("%s.ingestion_url", defaultProfile), DefaultIngestionURL)
			vpr.SetDefault(fmt.Sprintf("%s.queryflow_url", defaultProfile), DefaultQueryFlowURL)
			vpr.SetDefault(fmt.Sprintf("%s.staging_url", defaultProfile), DefaultStagingURL)
		}
	}
	if exists, err := readConfigFile("credentials", path, vpr, &ios); err != nil {
		return nil, err
	} else {
		if !exists {
			vpr.SetDefault(fmt.Sprintf("%s.core_key", defaultProfile), "")
			vpr.SetDefault(fmt.Sprintf("%s.ingestion_key", defaultProfile), "")
			vpr.SetDefault(fmt.Sprintf("%s.queryflow_key", defaultProfile), "")
			vpr.SetDefault(fmt.Sprintf("%s.staging_key", defaultProfile), "")
		}
	}

	return vpr, nil
}

// SaveConfigFromUser asks the user for the URLs and API Keys of the Discovery's components to save them in a profile.
// It then writes the current configuration into the given file.
func (d discovery) SaveConfigFromUser(ios iostreams.IOStreams, profile string, path string) error {
	reader := bufio.NewReader(ios.In)

	v := d.Config()

	fmt.Fprintf(ios.Out, "Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile)
	configKeys := []string{"core_url", "core_key", "ingestion_url", "ingestion_key", "queryflow_url", "queryflow_key", "staging_url", "staging_key"}
	for _, k := range configKeys {
		curr := v.Get(fmt.Sprintf("%s.%s", profile, k))
		if curr == nil {
			curr = "There is no value set for this property"
		}
		prompt := fmt.Sprintf("%s.%s [%q]: ", profile, k, curr)
		if _, err := fmt.Fprint(ios.Out, prompt); err != nil {
			return err
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		input := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")

		switch input {
		case "":
			continue
		case " ":
			v.Set(fmt.Sprintf("%s.%s", profile, k), "")
		default:
			v.Set(fmt.Sprintf("%s.%s", profile, k), input)
		}
	}

	err := v.WriteConfigAs(path)
	return err
}
