package cli

import (
	"fmt"
	"path/filepath"

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

func (d discovery) SaveCoreConfigFromUser(profile, path string) error {
	ios := d.IOStreams()
	v := d.Config()

	coreUrl, err := ios.AskUser("Core URL", v.Get(fmt.Sprintf("%s.core_url", profile)).(string))
	if err != nil {
		return err
	}

	switch coreUrl {
	case "":

	case " ":
		v.Set(fmt.Sprintf("%s.core_url", profile), "")
	default:
		v.Set(fmt.Sprintf("%s.core_url", profile), coreUrl)
	}

	coreKey, err := ios.AskUser("Core API Key", v.Get(fmt.Sprintf("%s.core_key", profile)).(string))
	if err != nil {
		return err
	}

	switch coreKey {
	case "":

	case " ":
		v.Set(fmt.Sprintf("%s.core_url", profile), "")
	default:
		v.Set(fmt.Sprintf("%s.core_url", profile), coreKey)
	}

	err = v.WriteConfigAs(path)
	return err
}

func (d discovery) SaveIngestionConfigFromUser(profile, path string) error {
	ios := d.IOStreams()
	v := d.Config()

	ingestionUrl, err := ios.AskUser("Ingestion URL", v.Get(fmt.Sprintf("%s.ingestion_url", profile)).(string))
	if err != nil {
		return err
	}

	switch ingestionUrl {
	case "":

	case " ":
		v.Set(fmt.Sprintf("%s.ingestion_url", profile), "")
	default:
		v.Set(fmt.Sprintf("%s.ingestion_url", profile), ingestionUrl)
	}

	ingestionKey, err := ios.AskUser("Ingestion API Key", v.Get(fmt.Sprintf("%s.ingestion_key", profile)).(string))
	if err != nil {
		return err
	}

	switch ingestionKey {
	case "":

	case " ":
		v.Set(fmt.Sprintf("%s.ingestion_url", profile), "")
	default:
		v.Set(fmt.Sprintf("%s.ingestion_url", profile), ingestionKey)
	}

	err = v.WriteConfigAs(path)
	return err
}

func (d discovery) SaveQueryFlowConfigFromUser(profile, path string) error {
	ios := d.IOStreams()
	v := d.Config()

	queryFlowUrl, err := ios.AskUser("QueryFlow URL", v.Get(fmt.Sprintf("%s.queryflow_url", profile)).(string))
	if err != nil {
		return err
	}

	switch queryFlowUrl {
	case "":

	case " ":
		v.Set(fmt.Sprintf("%s.queryflow_url", profile), "")
	default:
		v.Set(fmt.Sprintf("%s.queryflow_url", profile), queryFlowUrl)
	}

	queryFlowKey, err := ios.AskUser("QueryFlow API Key", v.Get(fmt.Sprintf("%s.queryflow_key", profile)).(string))
	if err != nil {
		return err
	}

	switch queryFlowKey {
	case "":

	case " ":
		v.Set(fmt.Sprintf("%s.queryflow_url", profile), "")
	default:
		v.Set(fmt.Sprintf("%s.queryflow_url", profile), queryFlowKey)
	}

	err = v.WriteConfigAs(path)
	return err
}

func (d discovery) SaveStagingConfigFromUser(profile, path string) error {
	ios := d.IOStreams()
	v := d.Config()

	stagingUrl, err := ios.AskUser("Staging URL", v.Get(fmt.Sprintf("%s.staging_url", profile)).(string))
	if err != nil {
		return err
	}

	switch stagingUrl {
	case "":

	case " ":
		v.Set(fmt.Sprintf("%s.staging_url", profile), "")
	default:
		v.Set(fmt.Sprintf("%s.staging_url", profile), stagingUrl)
	}

	stagingKey, err := ios.AskUser("Staging API Key", v.Get(fmt.Sprintf("%s.staging_key", profile)).(string))
	if err != nil {
		return err
	}

	switch stagingKey {
	case "":

	case " ":
		v.Set(fmt.Sprintf("%s.staging_url", profile), "")
	default:
		v.Set(fmt.Sprintf("%s.staging_url", profile), stagingKey)
	}

	err = v.WriteConfigAs(path)
	return err
}

// SaveConfigFromUser asks the user for the URLs and API Keys of the Discovery's components to save them in a profile.
// It then writes the current configuration into the given file.
func (d discovery) SaveConfigFromUser(profile string, path string) error {
	err := d.SaveCoreConfigFromUser(profile, path)
	if err != nil {
		return err
	}
	err = d.SaveIngestionConfigFromUser(profile, path)
	if err != nil {
		return err
	}
	err = d.SaveQueryFlowConfigFromUser(profile, path)
	if err != nil {
		return err
	}
	return d.SaveStagingConfigFromUser(profile, path)
}
