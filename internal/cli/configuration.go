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

func (d discovery) askUserConfig(profile, propertyName, property string) error {
	ios := d.IOStreams()
	v := d.Config()

	propertyInput, err := ios.AskUser(fmt.Sprintf("%s [%s]", propertyName, v.Get(fmt.Sprintf("%s.%s", profile, property)).(string)))
	if err != nil {
		return err
	}

	switch propertyInput {
	case "":

	case " ":
		v.Set(fmt.Sprintf("%s.%s", profile, property), "")
	default:
		v.Set(fmt.Sprintf("%s.%s", profile, property), propertyInput)
	}
	return nil
}

func (d discovery) SaveCoreConfigFromUser(profile, path string, standalone bool) error {
	ios := d.IOStreams()
	v := d.Config()

	if standalone {
		fmt.Fprintf(ios.Out, "Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile)
	}

	coreUrl, err := ios.AskUser(fmt.Sprintf("Core URL [%s]", v.Get(fmt.Sprintf("%s.core_url", profile)).(string)))
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

	coreKey, err := ios.AskUser(fmt.Sprintf("Core API Key [%s]", v.Get(fmt.Sprintf("%s.core_key", profile)).(string)))
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

func (d discovery) SaveIngestionConfigFromUser(profile, path string, standalone bool) error {
	ios := d.IOStreams()
	v := d.Config()

	if standalone {
		fmt.Fprintf(ios.Out, "Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile)
	}

	ingestionUrl, err := ios.AskUser(fmt.Sprintf("Ingestion URL [%s]", v.Get(fmt.Sprintf("%s.ingestion_url", profile)).(string)))
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

	ingestionKey, err := ios.AskUser(fmt.Sprintf("Ingestion API Key [%s]", v.Get(fmt.Sprintf("%s.ingestion_key", profile)).(string)))
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

func (d discovery) SaveQueryFlowConfigFromUser(profile, path string, standalone bool) error {
	ios := d.IOStreams()
	v := d.Config()

	if standalone {
		fmt.Fprintf(ios.Out, "Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile)
	}

	queryFlowUrl, err := ios.AskUser(fmt.Sprintf("QueryFlow URL [%s]", v.Get(fmt.Sprintf("%s.queryflow_url", profile)).(string)))
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

	queryFlowKey, err := ios.AskUser(fmt.Sprintf("QueryFlow API Key [%s]", v.Get(fmt.Sprintf("%s.queryflow_key", profile)).(string)))
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

func (d discovery) SaveStagingConfigFromUser(profile, path string, standalone bool) error {
	ios := d.IOStreams()
	v := d.Config()

	if standalone {
		fmt.Fprintf(ios.Out, "Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile)
	}

	stagingUrl, err := ios.AskUser(fmt.Sprintf("Staging URL [%s]", v.Get(fmt.Sprintf("%s.staging_url", profile)).(string)))
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

	stagingKey, err := ios.AskUser(fmt.Sprintf("Staging API Key [%s]", v.Get(fmt.Sprintf("%s.staging_key", profile)).(string)))
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
	fmt.Fprintf(d.IOStreams().Out, "Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile)

	err := d.SaveCoreConfigFromUser(profile, path, false)
	if err != nil {
		return err
	}
	err = d.SaveIngestionConfigFromUser(profile, path, false)
	if err != nil {
		return err
	}
	err = d.SaveQueryFlowConfigFromUser(profile, path, false)
	if err != nil {
		return err
	}
	return d.SaveStagingConfigFromUser(profile, path, false)
}
