package cli

import (
	"fmt"
	"math"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/viper"
)

const (
	DefaultCoreURL      string = "http://localhost:12010"
	DefaultStagingURL   string = "http://localhost:12020"
	DefaultIngestionURL string = "http://localhost:12030"
	DefaultQueryFlowURL string = "http://localhost:12040"
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

	defaultProfile := "default"
	vpr.SetDefault("profile", defaultProfile)

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

// Obfuscate modifies a string so that at least 60% of its characters are replaced by '*' characters.
func obfuscate(s string) string {
	if s == "" {
		return ""
	}

	r := []rune(s)
	n := len(r)

	maskCount := int(math.Ceil(0.6 * float64(n)))

	for i := 0; i < maskCount; i++ {
		r[i] = '*'
	}

	return string(r)
}

// AskUserConfig is an auxiliary function asks the user for the value they want to assign to a configuration property in the given profile.
// If the user inputs an empty string, the value is not changed.
// If the user inputs a space, the value is set to an empty string.
// If the user inputs a new value, the property is modified.
func (d discovery) askUserConfig(profile, propertyName, property string, sensitive bool) error {
	ios := d.IOStreams()
	v := d.Config()

	var value string
	if !(sensitive) {
		value = v.GetString(fmt.Sprintf("%s.%s", profile, property))
	} else {
		value = obfuscate(v.GetString(fmt.Sprintf("%s.%s", profile, property)))
	}

	propertyInput, err := ios.AskUser(fmt.Sprintf("%s [%s]: ", propertyName, value))
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

// SaveConfig separates de API Keys from Discovery's Viper configuration and writes the config and credentials into their own files.
func (d discovery) saveConfig() error {
	v := d.Config()
	apiKeys := []string{"core_key", "ingestion_key", "queryflow_key", "staging_key"}

	config := viper.New()
	credentials := viper.New()

	for _, setting := range v.AllKeys() {
		if setting != "profile" {
			parts := strings.Split(setting, ".")
			if slices.Contains(apiKeys, parts[len(parts)-1]) {
				credentials.Set(setting, v.Get(setting))
			} else {
				config.Set(setting, v.Get(setting))
			}
		} else {
			config.Set("profile", v.Get("profile"))
		}
	}

	err := config.WriteConfigAs(filepath.Join(d.ConfigPath(), "config.toml"))
	if err != nil {
		return err
	}

	return credentials.WriteConfigAs(filepath.Join(d.ConfigPath(), "credentials.toml"))
}

// SaveCoreConfigFromUser asks the user for the values it wants to set for Discovery Core's configuration properties for the given profile.
// It then writes the new configuration to the Discovery struct's Config Path.
// The standalone parameter is used to display the instructions in case this function is used by itself and not by the SaveConfigFromUser() function.
func (d discovery) SaveCoreConfigFromUser(profile string, standalone bool) error {
	ios := d.IOStreams()

	if standalone {
		fmt.Fprintf(ios.Out, "Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile)
	}

	urlErr := d.askUserConfig(profile, "Core URL", "core_url", false)
	if urlErr != nil {
		return urlErr
	}

	keyErr := d.askUserConfig(profile, "Core API Key", "core_key", true)
	if keyErr != nil {
		return keyErr
	}

	saveErr := d.saveConfig()
	if saveErr != nil {
		return saveErr
	}

	if standalone {
		fmt.Fprint(ios.Out, "Core configuration saved successfully")
	}

	return nil
}

// SaveIngestionConfigFromUser asks the user for the values it wants to set for Discovery Ingestion's configuration properties for the given profile.
// It then writes the new configuration to the Discovery struct's Config Path.
// The standalone parameter is used to display the instructions in case this function is used by itself and not by the SaveConfigFromUser() function.
func (d discovery) SaveIngestionConfigFromUser(profile string, standalone bool) error {
	ios := d.IOStreams()

	if standalone {
		fmt.Fprintf(ios.Out, "Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile)
	}

	urlErr := d.askUserConfig(profile, "Ingestion URL", "ingestion_url", false)
	if urlErr != nil {
		return urlErr
	}

	keyErr := d.askUserConfig(profile, "Ingestion API Key", "ingestion_key", true)
	if keyErr != nil {
		return keyErr
	}

	saveErr := d.saveConfig()
	if saveErr != nil {
		return saveErr
	}

	if standalone {
		fmt.Fprint(ios.Out, "Ingestion configuration saved successfully")
	}

	return nil
}

// SaveQueryFlowConfigFromUser asks the user for the values it wants to set for Discovery QueryFlow's configuration properties for the given profile.
// It then writes the new configuration to the Discovery struct's Config Path.
// The standalone parameter is used to display the instructions in case this function is used by itself and not by the SaveConfigFromUser() function.
func (d discovery) SaveQueryFlowConfigFromUser(profile string, standalone bool) error {
	ios := d.IOStreams()

	if standalone {
		fmt.Fprintf(ios.Out, "Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile)
	}

	urlErr := d.askUserConfig(profile, "QueryFlow URL", "queryflow_url", false)
	if urlErr != nil {
		return urlErr
	}

	keyErr := d.askUserConfig(profile, "QueryFlow API Key", "queryflow_key", true)
	if keyErr != nil {
		return keyErr
	}

	saveErr := d.saveConfig()
	if saveErr != nil {
		return saveErr
	}

	if standalone {
		fmt.Fprint(ios.Out, "QueryFlow configuration saved successfully")
	}

	return nil
}

// SaveStagingConfigFromUser asks the user for the values it wants to set for Discovery Staging's configuration properties for the given profile.
// It then writes the new configuration to the Discovery struct's Config Path.
// The standalone parameter is used to display the instructions in case this function is used by itself and not by the SaveConfigFromUser() function.
func (d discovery) SaveStagingConfigFromUser(profile string, standalone bool) error {
	ios := d.IOStreams()

	if standalone {
		fmt.Fprintf(ios.Out, "Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile)
	}

	urlErr := d.askUserConfig(profile, "Staging URL", "staging_url", false)
	if urlErr != nil {
		return urlErr
	}

	keyErr := d.askUserConfig(profile, "Staging API Key", "staging_key", true)
	if keyErr != nil {
		return keyErr
	}

	saveErr := d.saveConfig()
	if saveErr != nil {
		return saveErr
	}

	if standalone {
		fmt.Fprint(ios.Out, "Staging configuration saved successfully")
	}

	return nil
}

// SaveConfigFromUser asks the user for the URLs and API Keys of the Discovery's components to save them in a profile.
// It then writes the current configuration into the given file.
func (d discovery) SaveConfigFromUser(profile string) error {
	fmt.Fprintf(d.IOStreams().Out, "Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile)

	err := d.SaveCoreConfigFromUser(profile, false)
	if err != nil {
		return err
	}
	err = d.SaveIngestionConfigFromUser(profile, false)
	if err != nil {
		return err
	}
	err = d.SaveQueryFlowConfigFromUser(profile, false)
	if err != nil {
		return err
	}
	err = d.SaveStagingConfigFromUser(profile, false)
	if err != nil {
		return err
	}

	fmt.Fprint(d.IOStreams().Out, "Configuration saved successfully")

	return nil
}
