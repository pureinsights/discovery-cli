package cli

import (
	"fmt"
	"math"
	"os"
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
	SaveHeader          string = "Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n"
	PrintHeader         string = "Showing the configuration of profile %q:\n\n"
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
		return true, err
	}
	return true, nil
}

// InitializeConfig reads the config and credentials configuration files found in the given path and sets up the Viper instance with their values.
func InitializeConfig(ios iostreams.IOStreams, path string) (*viper.Viper, error) {
	vpr := viper.New()

	defaultProfile := "default"
	vpr.SetDefault("profile", defaultProfile)

	if exists, err := readConfigFile("config", path, vpr, &ios); err != nil {
		return nil, NewErrorWithCause(ErrorExitCode, err, "Could not read the configuration file")
	} else {
		if !exists {
			vpr.SetDefault(fmt.Sprintf("%s.core_url", defaultProfile), DefaultCoreURL)
			vpr.SetDefault(fmt.Sprintf("%s.ingestion_url", defaultProfile), DefaultIngestionURL)
			vpr.SetDefault(fmt.Sprintf("%s.queryflow_url", defaultProfile), DefaultQueryFlowURL)
			vpr.SetDefault(fmt.Sprintf("%s.staging_url", defaultProfile), DefaultStagingURL)
		}
	}
	if exists, err := readConfigFile("credentials", path, vpr, &ios); err != nil {
		return nil, NewErrorWithCause(ErrorExitCode, err, "Could not read the credentials file")
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
	temporaryProperties := []string{"profile"}

	config := viper.New()
	credentials := viper.New()

	for _, setting := range v.AllKeys() {
		switch {
		case slices.Contains(temporaryProperties, setting):

		case slices.Contains(apiKeys, strings.Split(setting, ".")[len(strings.Split(setting, "."))-1]):
			credentials.Set(setting, v.Get(setting))
		default:
			config.Set(setting, v.Get(setting))
		}

	}

	err := config.WriteConfigAs(filepath.Join(d.ConfigPath(), "config.toml"))
	if err != nil {
		return err
	}

	return credentials.WriteConfigAs(filepath.Join(d.ConfigPath(), "credentials.toml"))
}

// SetDiscoveryDir creates the Discovery directory if it does not exist and returns its path if an error did not occur.
func SetDiscoveryDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(home, ".discovery")

	if err := os.MkdirAll(configPath, 0o700); err != nil {
		return "", err
	}

	return configPath, nil
}

// SaveUrlAndAPIKey asks the user for the URL and API key of a Discovery component and saves them.
func (d discovery) saveUrlAndAPIKey(profile, component, componentName string, standalone bool) error {
	ios := d.IOStreams()

	if standalone {
		fmt.Fprintf(ios.Out, SaveHeader, profile)
	}

	err := d.askUserConfig(profile, fmt.Sprintf("%s URL", componentName), fmt.Sprintf("%s_url", component), false)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Failed to get %s's URL", componentName)
	}

	err = d.askUserConfig(profile, fmt.Sprintf("%s API key", componentName), fmt.Sprintf("%s_key", component), false)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Failed to get %s's API key", componentName)
	}

	err = d.saveConfig()
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Failed to save %s's configuration", componentName)
	}

	return nil
}

// SaveCoreConfigFromUser asks the user for the values it wants to set for Discovery Core's configuration properties for the given profile.
// It then writes the new configuration to the Discovery struct's Config Path.
// The standalone parameter is used to display the instructions in case this function is used by itself and not by the SaveConfigFromUser() function.
func (d discovery) SaveCoreConfigFromUser(profile string, standalone bool) error {
	return d.saveUrlAndAPIKey(profile, "core", "Core", standalone)
}

// SaveIngestionConfigFromUser asks the user for the values it wants to set for Discovery Ingestion's configuration properties for the given profile.
// It then writes the new configuration to the Discovery struct's Config Path.
// The standalone parameter is used to display the instructions in case this function is used by itself and not by the SaveConfigFromUser() function.
func (d discovery) SaveIngestionConfigFromUser(profile string, standalone bool) error {
	return d.saveUrlAndAPIKey(profile, "ingestion", "Ingestion", standalone)
}

// SaveQueryFlowConfigFromUser asks the user for the values it wants to set for Discovery QueryFlow's configuration properties for the given profile.
// It then writes the new configuration to the Discovery struct's Config Path.
// The standalone parameter is used to display the instructions in case this function is used by itself and not by the SaveConfigFromUser() function.
func (d discovery) SaveQueryFlowConfigFromUser(profile string, standalone bool) error {
	return d.saveUrlAndAPIKey(profile, "queryflow", "QueryFlow", standalone)
}

// SaveStagingConfigFromUser asks the user for the values it wants to set for Discovery Staging's configuration properties for the given profile.
// It then writes the new configuration to the Discovery struct's Config Path.
// The standalone parameter is used to display the instructions in case this function is used by itself and not by the SaveConfigFromUser() function.
func (d discovery) SaveStagingConfigFromUser(profile string, standalone bool) error {
	return d.saveUrlAndAPIKey(profile, "staging", "Staging", standalone)
}

// SaveConfigFromUser asks the user for the URLs and API Keys of the Discovery's components to save them in a profile.
// It then writes the current configuration into the given file.
func (d discovery) SaveConfigFromUser(profile string) error {
	fmt.Fprintf(d.IOStreams().Out, SaveHeader, profile)

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
	return d.SaveStagingConfigFromUser(profile, false)
}

// PrintConfig is the auxiliary function to print a property's value to the user.
// It prints the property with the given profile and name.
// If the value of the property is sensitive, it is obfuscated.
func (d discovery) printConfig(profile, propertyName, property string, sensitive bool) error {
	v := d.Config()
	ios := d.IOStreams()

	if v.IsSet(fmt.Sprintf("%s.%s", profile, property)) {
		value := v.GetString(fmt.Sprintf("%s.%s", profile, property))
		if sensitive {
			value = obfuscate(value)
		}
		_, err := fmt.Fprintf(ios.Out, "%s: %q\n", propertyName, value)
		return err
	}
	return nil
}

// PrintURLAndAPIKey prints the URL and API key of a Discovery component to the Out IOStream.
func (d discovery) printURLAndAPIKey(profile, component, componentName string, standalone, sensitive bool) error {
	ios := d.IOStreams()
	var err error
	if standalone {
		fmt.Fprintf(ios.Out, PrintHeader, profile)
	}

	err = d.printConfig(profile, fmt.Sprintf("%s URL", componentName), fmt.Sprintf("%s_url", component), false)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not print %s's URL", componentName)
	}

	err = d.printConfig(profile, fmt.Sprintf("%s API Key", componentName), fmt.Sprintf("%s_key", component), sensitive)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not print %s's API key", componentName)
	}
	return nil
}

// PrintCoreConfigToUser prints the Discovery Core's configuration properties for the given profile.
// The caller of the function can determine if the API Key is sensitive so that it can be obfuscated.
// The standalone parameter is used to display the header information in case this function is used by itself and not by the PrintConfigToUser() function.
func (d discovery) PrintCoreConfigToUser(profile string, sensitive, standalone bool) error {
	return d.printURLAndAPIKey(profile, "core", "Core", standalone, sensitive)
}

// PrintIngestionConfigToUser prints the Discovery Ingestion's configuration properties for the given profile.
// The caller of the function can determine if the API Key is sensitive so that it can be obfuscated.
// The standalone parameter is used to display the header information in case this function is used by itself and not by the PrintConfigToUser() function.
func (d discovery) PrintIngestionConfigToUser(profile string, sensitive, standalone bool) error {
	return d.printURLAndAPIKey(profile, "ingestion", "Ingestion", standalone, sensitive)
}

// PrintQueryFlowConfigToUser prints the Discovery QueryFlow's configuration properties for the given profile.
// The caller of the function can determine if the API Key is sensitive so that it can be obfuscated.
// The standalone parameter is used to display the header information in case this function is used by itself and not by the PrintConfigToUser() function.
func (d discovery) PrintQueryFlowConfigToUser(profile string, sensitive, standalone bool) error {
	return d.printURLAndAPIKey(profile, "queryflow", "QueryFlow", standalone, sensitive)
}

// PrintStagingConfigToUser prints the Discovery Staging's configuration properties for the given profile.
// The caller of the function can determine if the API Key is sensitive so that it can be obfuscated.
// The standalone parameter is used to display the header information in case this function is used by itself and not by the PrintConfigToUser() function.
func (d discovery) PrintStagingConfigToUser(profile string, sensitive, standalone bool) error {
	return d.printURLAndAPIKey(profile, "staging", "Staging", standalone, sensitive)
}

// PrintConfigToUser prints the Discovery Components' configuration properties for the given profile.
// The caller of the function can determine if the API Keys are sensitive so that they can be obfuscated.
// The standalone parameter is used to display the header information in case this function is used by itself and not by the PrintConfigToUser() function.
func (d discovery) PrintConfigToUser(profile string, sensitive bool) error {
	fmt.Fprintf(d.IOStreams().Out, PrintHeader, profile)

	err := d.PrintCoreConfigToUser(profile, sensitive, false)
	if err != nil {
		return err
	}
	err = d.PrintIngestionConfigToUser(profile, sensitive, false)
	if err != nil {
		return err
	}
	err = d.PrintQueryFlowConfigToUser(profile, sensitive, false)
	if err != nil {
		return err
	}
	return d.PrintStagingConfigToUser(profile, sensitive, false)
}
