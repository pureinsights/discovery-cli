package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/fileutils"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_readConfigFile tests the readConfigFile() auxiliary function.
func Test_readConfigFile(t *testing.T) {
	tests := []struct {
		name                   string
		baseName               string
		config                 string
		expectedConfig         map[string]string
		expectedFileExistsBool bool
		err                    error
	}{
		// Working cases
		{
			name:     "Config file can be read",
			baseName: "config",
			config: `
[default]
core_url="http://localhost:12010"

[cn]
core_url="http://discovery.core.cn"	
`,
			expectedConfig: map[string]string{
				"default.core_url": "http://localhost:12010",
				"cn.core_url":      "http://discovery.core.cn",
			},
			expectedFileExistsBool: true,
			err:                    nil,
		},
		{
			name:                   "File does not exist",
			baseName:               "fail",
			config:                 ``,
			expectedConfig:         map[string]string{},
			expectedFileExistsBool: false,
			err:                    nil,
		},
		{
			name:     "Cannot Merge configuration",
			baseName: "config",
			config: `
{
  "default": {
    "core_key": ""
  },
  "cn": {
    "core_key": "discovery.key.core.cn"
  }
}
`,
			expectedConfig: map[string]string{
				"default.core_key": "",
				"cn.core_key":      "discovery.key.core.cn",
			},
			expectedFileExistsBool: true,
			err:                    errors.New("While parsing config: toml: invalid character at start of key: {"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.baseName != "fail" {
				_, err := fileutils.CreateTemporaryFile(dir, tc.baseName+".toml", tc.config)
				require.NoError(t, err)
			}

			errStream := &bytes.Buffer{}
			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: os.Stdout,
				Err: errStream,
			}

			viper := viper.New()
			exists, err := readConfigFile(tc.baseName, dir, viper, &ios)
			assert.Equal(t, tc.expectedFileExistsBool, exists)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				if tc.config != "" {
					for k, v := range tc.expectedConfig {
						assert.Equal(t, v, viper.GetString(k))
					}
				} else {
					got := errStream.String()
					assert.Contains(t, got,
						fmt.Sprintf("Configuration file %q not found under %q; using default values.\n",
							tc.baseName, filepath.Clean(dir)),
					)
				}
			}
		})
	}
}

// TestInitializeConfig tests the InitializeConfig() function
func TestInitializeConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         string
		credentials    string
		expectedConfig map[string]string
		err            error
	}{
		// Working cases
		{
			name: "There are config and credentials files",
			config: `
[default]
core_url="http://localhost:3000"

[cn]
core_url="http://discovery.core.cn"	
`,
			credentials: `
[default]
core_key="APIKey"

[cn]
core_key="discovery.key.core.cn"
`,
			expectedConfig: map[string]string{
				"default.core_url": "http://localhost:3000",
				"cn.core_url":      "http://discovery.core.cn",
				"default.core_key": "APIKey",
				"cn.core_key":      "discovery.key.core.cn",
			},
			err: nil,
		},
		{
			name: "There is only a config file",
			config: `
[default]
core_url="http://localhost:8083"

[cn]
core_url="http://discovery.core.cn"	
`,
			credentials: ``,
			expectedConfig: map[string]string{
				"default.core_url": "http://localhost:8083",
				"cn.core_url":      "http://discovery.core.cn",
				"default.core_key": "",
			},
			err: nil,
		},
		{
			name:   "There is only a credentials file",
			config: ``,
			credentials: `
[default]
core_key="APIKey"

[cn]
core_key="discovery.key.core.cn"
`,
			expectedConfig: map[string]string{
				"default.core_url": "http://localhost:12010",
				"default.core_key": "APIKey",
				"cn.core_key":      "discovery.key.core.cn",
			},
			err: nil,
		},
		{
			name:        "There are no config files",
			config:      ``,
			credentials: ``,
			expectedConfig: map[string]string{
				"default.core_url": "http://localhost:12010",
				"default.core_key": "",
			},
			err: nil,
		},
		{
			name: "Reading the config file fails",
			config: `
{
  "default": {
    "core_url": "http://localhost:12010"
  },
  "cn": {
    "core_url": "http://discovery.core.cn"
  }
}
`,
			credentials: ``,
			expectedConfig: map[string]string{
				"default.core_url": "http://localhost:12010",
				"cn.core_url":      "http://discovery.core.cn",
				"default.core_key": "",
			},
			err: NewErrorWithCause(ErrorExitCode, errors.New("While parsing config: toml: invalid character at start of key: {"), "Could not read the configuration file"),
		},
		{
			name:   "Reading the credentials file fails",
			config: ``,
			credentials: `
{
  "default": {
    "core_key": ""
  },
  "cn": {
    "core_key": "discovery.key.core.cn"
  }
}
`,
			expectedConfig: map[string]string{
				"default.core_url": "http://localhost:12010",
				"default.core_key": "",
				"cn.core_key":      "discovery.key.core.cn",
			},
			err: NewErrorWithCause(ErrorExitCode, errors.New("While parsing config: toml: invalid character at start of key: {"), "Could not read the credentials file"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: os.Stdout,
				Err: os.Stderr,
			}

			dir := t.TempDir()
			if tc.config != "" {
				_, err := fileutils.CreateTemporaryFile(dir, "config.toml", tc.config)
				require.NoError(t, err)
			}

			if tc.credentials != "" {
				_, err := fileutils.CreateTemporaryFile(dir, "credentials.toml", tc.credentials)
				require.NoError(t, err)
			}

			viper, err := InitializeConfig(ios, dir)
			if tc.err != nil {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				for k, v := range tc.expectedConfig {
					assert.Equal(t, v, viper.GetString(k))
				}
			}
		})
	}
}

// Test_obfuscate tests the obfuscate() function.
func Test_obfuscate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "ascii len 10 (60% masked -> 6 masked, 4 visible)",
			input:    "ABCDEFGHIJ",
			expected: "******GHIJ",
		},
		{
			name:     "ascii len 5 (60% masked -> 3 masked, 2 visible)",
			input:    "abcde",
			expected: "***de",
		},
		{
			name:     "ascii len 4 (60% masked -> 3 masked, 1 visible)",
			input:    "abcd",
			expected: "***d",
		},
		{
			name:     "ascii len 3 (60% masked -> 2 masked, 1 visible)",
			input:    "abc",
			expected: "**c",
		},

		{
			name:     "two characters (60% masked -> 2 masked, 0 visible)",
			input:    "xy",
			expected: "**",
		},
		{
			name:     "single character",
			input:    "x",
			expected: "*",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := obfuscate(tc.input)
			assert.Equal(t, tc.expected, got)
		})
	}
}

// Test_discovery_AskUserConfig tests the discovery.AskUser() function
func Test_discovery_askUserConfig(t *testing.T) {
	const profile = "cn"
	const prop = "core_url"
	const propName = "Core URL"
	const initial = "http://localhost:8080"

	tests := []struct {
		name           string
		input          string
		inReader       io.Reader
		err            error
		expectedResult string
		sensitive      bool
	}{
		{
			name:           "Keep value when user presses Enter",
			input:          "\n",
			err:            nil,
			expectedResult: initial,
			sensitive:      false,
		},
		{
			name:           "Set empty when user types single space",
			input:          " \n",
			err:            nil,
			expectedResult: "",
			sensitive:      false,
		},
		{
			name:           "Set new value",
			input:          "http://discovery.core.cn\n",
			err:            nil,
			expectedResult: "http://discovery.core.cn",
			sensitive:      true,
		},
		{
			name:           "Value without newline then EOF returns value, no error",
			input:          "http://discovery.core.cn",
			err:            nil,
			expectedResult: "http://discovery.core.cn",
			sensitive:      true,
		},
		{
			name:           "Space without newline then EOF sets empty, no error",
			input:          " ",
			err:            nil,
			expectedResult: "",
			sensitive:      false,
		},
		{
			name:           "Immediate EOF (empty reader) returns empty string, no error",
			input:          "",
			err:            nil,
			expectedResult: initial,
			sensitive:      false,
		},
		{
			name:           "CRLF line endings handled",
			input:          "http://discovery.core.cn\r\n",
			err:            nil,
			expectedResult: "http://discovery.core.cn",
			sensitive:      true,
		},
		{
			name:      "Reading from the In IOStream fails",
			inReader:  testutils.ErrReader{Err: errors.New("read failed")},
			err:       fmt.Errorf("read failed"),
			sensitive: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var in io.Reader
			if tc.inReader != nil {
				in = tc.inReader
			} else {
				in = strings.NewReader(tc.input)
			}

			var out bytes.Buffer

			ios := iostreams.IOStreams{
				In:  in,
				Out: &out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			vpr.Set(fmt.Sprintf("%s.%s", profile, prop), initial)

			d := NewDiscovery(&ios, vpr, "")

			err := d.askUserConfig(profile, propName, prop, tc.sensitive)

			if tc.err != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				got := d.Config().GetString(profile + "." + prop)
				require.Equal(t, tc.expectedResult, got, "property value mismatch")
			}
			if tc.sensitive {
				require.Contains(t, out.String(), propName+" ["+obfuscate(initial)+"]")
			} else {
				require.Contains(t, out.String(), propName+" ["+initial+"]")
			}

		})
	}
}

// Test_discovery_saveConfig tests the discovery.saveConfig() function.
func Test_discovery_saveConfig(t *testing.T) {
	tests := []struct {
		name                string
		config              map[string]string
		writePath           string
		expectedConfig      map[string]string
		expectedCredentials map[string]string
		err                 error
	}{
		// Working cases
		{
			name:      "Every value exists",
			writePath: t.TempDir(),
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.ingestion_url": "http://localhost:12030",
				"cn.queryflow_url": "http://localhost:12040",
				"cn.staging_url":   "http://localhost:12020",
				"cn.core_key":      "core321",
				"cn.ingestion_key": "ingestion432",
				"cn.queryflow_key": "queryflow123",
				"cn.staging_key":   "staging235",
			},
			expectedConfig: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.ingestion_url": "http://localhost:12030",
				"cn.queryflow_url": "http://localhost:12040",
				"cn.staging_url":   "http://localhost:12020",
			},
			expectedCredentials: map[string]string{
				"cn.core_key":      "core321",
				"cn.ingestion_key": "ingestion432",
				"cn.queryflow_key": "queryflow123",
				"cn.staging_key":   "staging235",
			},
			err: nil,
		},
		{
			name:      "No keys exist",
			writePath: t.TempDir(),
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.ingestion_url": "http://localhost:12030",
				"cn.queryflow_url": "http://localhost:12040",
				"cn.staging_url":   "http://localhost:12020",
			},
			expectedConfig: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.ingestion_url": "http://localhost:12030",
				"cn.queryflow_url": "http://localhost:12040",
				"cn.staging_url":   "http://localhost:12020",
			},
			expectedCredentials: map[string]string{},
			err:                 nil,
		},
		{
			name:      "Only keys exist",
			writePath: t.TempDir(),
			config: map[string]string{
				"cn.core_key":      "core321",
				"cn.ingestion_key": "ingestion432",
				"cn.queryflow_key": "queryflow123",
				"cn.staging_key":   "staging235",
			},
			expectedConfig: map[string]string{},
			expectedCredentials: map[string]string{
				"cn.core_key":      "core321",
				"cn.ingestion_key": "ingestion432",
				"cn.queryflow_key": "queryflow123",
				"cn.staging_key":   "staging235",
			},
			err: nil,
		},
		{
			name:      "There are keys with multiple periods in their viper keys",
			writePath: t.TempDir(),
			config: map[string]string{
				"cn.core_url":            "http://localhost:12010",
				"cn.ingestion_url":       "http://localhost:12030",
				"cn.queryflow_url":       "http://localhost:12040",
				"cn.staging_url":         "http://localhost:12020",
				"cn.core_key":            "core321",
				"cn.cn.ingestion_key":    "ingestion432",
				"cn.cn.cn.queryflow_key": "queryflow123",
				"cn.cn.cn.staging_key":   "staging235",
			},
			expectedConfig: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.ingestion_url": "http://localhost:12030",
				"cn.queryflow_url": "http://localhost:12040",
				"cn.staging_url":   "http://localhost:12020",
			},
			expectedCredentials: map[string]string{
				"cn.core_key":            "core321",
				"cn.cn.ingestion_key":    "ingestion432",
				"cn.cn.cn.queryflow_key": "queryflow123",
				"cn.cn.cn.staging_key":   "staging235",
			},
			err: nil,
		},

		// Error cases
		{
			name:      "Writing to config.toml fails",
			writePath: "doesnotexist",
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.ingestion_url": "http://localhost:12030",
				"cn.queryflow_url": "http://localhost:12040",
				"cn.staging_url":   "http://localhost:12020",
				"cn.core_key":      "core321",
				"cn.ingestion_key": "ingestion432",
				"cn.queryflow_key": "queryflow123",
				"cn.staging_key":   "staging235",
			},
			expectedConfig: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.ingestion_url": "http://localhost:12030",
				"cn.queryflow_url": "http://localhost:12040",
				"cn.staging_url":   "http://localhost:12020",
			},
			expectedCredentials: map[string]string{
				"cn.core_key":      "core321",
				"cn.ingestion_key": "ingestion432",
				"cn.queryflow_key": "queryflow123",
				"cn.staging_key":   "staging235",
			},
			err: fmt.Errorf("cannot find the path specified"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: os.Stdout,
				Err: os.Stderr,
			}

			d := NewDiscovery(&ios, vpr, tc.writePath)

			err := d.saveConfig()
			if tc.err != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, fs.ErrNotExist))
			} else {
				require.NoError(t, err)
				configVpr := viper.New()
				configVpr.SetConfigFile(filepath.Join(tc.writePath, "config.toml"))
				require.NoError(t, configVpr.ReadInConfig())

				for k, expected := range tc.expectedConfig {
					require.Equal(t, expected, configVpr.GetString(k))
				}

				credentialsVpr := viper.New()
				credentialsVpr.SetConfigFile(filepath.Join(tc.writePath, "credentials.toml"))
				require.NoError(t, credentialsVpr.ReadInConfig())

				for k, expected := range tc.expectedCredentials {
					require.Equal(t, expected, credentialsVpr.GetString(k))
				}
			}
		})
	}
}

// TestSetDiscoveryDir_MkDirAllFails tests the SetDiscoveryDir() function when the ~/.discovery directory could not be made
func TestSetDiscoveryDir_MkDirAllFails(t *testing.T) {
	tmp := t.TempDir()

	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	target := filepath.Join(tmp, ".discovery")

	require.NoError(t, os.WriteFile(target, []byte("MkDirAll will fail"), 0o600))

	_, err := SetDiscoveryDir()
	require.Error(t, err)
	var errStruct Error
	require.ErrorAs(t, err, &errStruct)
	if cliError, ok := err.(*Error); ok {
		cause := cliError.Cause
		isFileError := errors.Is(cause, fs.ErrNotExist) || errors.Is(cause, syscall.ENOTDIR)
		assert.True(t, isFileError)
		assert.Equal(t, "Could not create the /.discovery directory", cliError.Message)
		assert.Equal(t, ErrorExitCode, cliError.ExitCode)
	}
}

// TestSetDiscoveryDir_osUserHomeDirFails tests the SetDiscoveryDir() function when the environment variables are not defined.
func TestSetDiscoveryDir_osUserHomeDirFails(t *testing.T) {

	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	_, err := SetDiscoveryDir()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not defined")
}

// TestSetDiscoveryDir_Success tests the SetDiscoveryDir() function when the ~/.discovery directory could be created successfully
func TestSetDiscoveryDir_DiscoveryDirCreated(t *testing.T) {
	tmp := t.TempDir()

	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	configPath, err := SetDiscoveryDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmp, ".discovery"), configPath)
}

// TestSetDiscoveryDir_Success tests the SetDiscoveryDir() function when the ~/.discovery directory already exists
func TestSetDiscoveryDir_DiscoveryDirExists(t *testing.T) {
	tmp := t.TempDir()

	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	target := filepath.Join(tmp, ".discovery")

	require.NoError(t, os.Mkdir(target, 0o600))

	configPath, err := SetDiscoveryDir()
	require.NoError(t, err)
	assert.Equal(t, target, configPath)
}

// Test_discovery_SaveConfigFromUser_AllConfigPresent tests the discovery.SaveConfigFromUser() when there is a configuration for every possible URL and API Key
func Test_discovery_SaveConfigFromUser_AllConfigPresent(t *testing.T) {
	const profile = "cn"
	config := map[string]string{
		"cn.core_url":      "http://localhost:8080",
		"cn.ingestion_url": "http://localhost:8080",
		"cn.queryflow_url": "http://localhost:8088",
		"cn.staging_url":   "http://localhost:8081",
		"cn.core_key":      "core321",
		"cn.ingestion_key": "ingestion432",
		"cn.queryflow_key": "queryflow123",
		"cn.staging_key":   "staging235",
	}

	tests := []struct {
		name       string
		input      string
		inReader   io.Reader
		writePath  string
		err        error
		expectKeys map[string]string
	}{
		{
			name:      "Keep every existing value",
			input:     strings.Repeat("\n", 8),
			writePath: t.TempDir(),
			expectKeys: map[string]string{
				"core_url":      config["cn.core_url"],
				"core_key":      config["cn.core_key"],
				"ingestion_url": config["cn.ingestion_url"],
				"ingestion_key": config["cn.ingestion_key"],
			},
		},
		{
			name:      "Set Core URL and Staging URL to empty, keep the rest",
			input:     " \n" + strings.Repeat("\n", 5) + " \n\n",
			writePath: t.TempDir(),
			expectKeys: map[string]string{
				"core_url":      "",
				"core_key":      config["cn.core_key"],
				"ingestion_url": config["cn.ingestion_url"],
				"staging_url":   "",
			},
		},
		{
			name:      "Set Core URL to new value, keep the rest",
			input:     "http://discovery.core.cn\n" + strings.Repeat("\n", 7),
			writePath: t.TempDir(),
			expectKeys: map[string]string{
				"core_url":      "http://discovery.core.cn",
				"core_key":      config["cn.core_key"],
				"ingestion_url": config["cn.ingestion_url"],
			},
		},
		{
			name:      "The user writes an End Of File while inputting the values",
			input:     "http://discovery.core.cn\ncore123",
			writePath: t.TempDir(),
			expectKeys: map[string]string{
				"core_url": "http://discovery.core.cn",
				"core_key": "core123",
			},
		},
		{
			name:      "Reading from the Core Config In IOStream fails",
			inReader:  testutils.ErrReader{Err: errors.New("read failed")},
			writePath: t.TempDir(),
			err:       NewErrorWithCause(ErrorExitCode, fmt.Errorf("read failed"), "Failed to get Core's URL"),
		},
		{
			name:      "Reading from the Ingestion Config In IOStream fails",
			inReader:  io.MultiReader(strings.NewReader("http://discovery.core.cn\n\n"), testutils.ErrReader{Err: errors.New("read failed")}),
			writePath: t.TempDir(),
			err:       NewErrorWithCause(ErrorExitCode, fmt.Errorf("read failed"), "Failed to get Ingestion's URL"),
		},
		{
			name:      "Reading from the QueryFlow Config In IOStream fails",
			inReader:  io.MultiReader(strings.NewReader("http://discovery.core.cn\n\nhttp://discovery.ingestion.cn\n\n"), testutils.ErrReader{Err: errors.New("read failed")}),
			writePath: t.TempDir(),
			err:       NewErrorWithCause(ErrorExitCode, fmt.Errorf("read failed"), "Failed to get QueryFlow's URL"),
		},
		{
			name:      "Reading from the Staging Config In IOStream fails",
			inReader:  io.MultiReader(strings.NewReader("http://discovery.core.cn\n\nhttp://discovery.ingestion.cn\n\nhttp://discovery.queryflow.cn\n\n"), testutils.ErrReader{Err: errors.New("read failed")}),
			writePath: t.TempDir(),
			err:       NewErrorWithCause(ErrorExitCode, fmt.Errorf("read failed"), "Failed to get Staging's URL"),
		},
		{
			name:      "Invalid write location",
			input:     strings.Repeat("\n", 8),
			writePath: "doesnotexist",
			err:       NewErrorWithCause(ErrorExitCode, fmt.Errorf("open doesnotexist\\config.toml: The system cannot find the path specified."), "Failed to save Core's configuration"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var in io.Reader
			if tc.inReader != nil {
				in = tc.inReader
			} else {
				in = strings.NewReader(tc.input)
			}

			out := &bytes.Buffer{}

			ios := iostreams.IOStreams{
				In:  in,
				Out: out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			for k, v := range config {
				vpr.Set(k, v)
			}

			d := NewDiscovery(&ios, vpr, tc.writePath)

			err := d.SaveConfigFromUser(profile)
			if tc.err != nil {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				if cliError, ok := err.(*Error); ok {
					cause := cliError.Cause
					if !errors.Is(cause, fs.ErrNotExist) {
						assert.EqualError(t, err, tc.err.Error())
					} else {
						tcError, _ := err.(*Error)
						assert.Equal(t, tcError.Message, cliError.Message)
						assert.Equal(t, tcError.ExitCode, cliError.ExitCode)
					}
				}

			} else {
				require.NoError(t, err)
				vpr, err := InitializeConfig(ios, tc.writePath)
				require.NoError(t, err)

				for k, expected := range tc.expectKeys {
					gotVal := vpr.GetString(profile + "." + k)
					require.Equal(t, expected, gotVal)
				}

				assert.Contains(t, out.String(), fmt.Sprintf("Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile))
			}
		})
	}
}

// Test_discovery_SaveConfigFromUser_NotAllConfigPresent tests the discovery.SaveConfigFromUser() function when there are some properties with no explicit values set.
func Test_discovery_SaveConfigFromUser_NotAllConfigPresent(t *testing.T) {
	const profile = "cn"
	config := map[string]string{
		"cn.core_url":      "http://localhost:8080",
		"cn.staging_url":   "http://localhost:8081",
		"cn.core_key":      "core321",
		"cn.ingestion_key": "ingestion432",
	}

	dir := t.TempDir()

	in := strings.NewReader("\ncore123\nhttp://localhost:8080\n \n\nqueryflow123\nstaging.cn.aws.com\n\n")

	outStream := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: outStream,
		Err: os.Stderr,
	}

	vpr := viper.New()
	for k, v := range config {
		vpr.Set(k, v)
	}

	d := NewDiscovery(&ios, vpr, dir)

	err := d.SaveConfigFromUser(profile)
	require.NoError(t, err)
	got, err := InitializeConfig(ios, dir)
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:8080", got.Get("cn.core_url"))
	assert.Equal(t, "core123", got.Get("cn.core_key"))
	assert.Equal(t, "http://localhost:8080", got.Get("cn.ingestion_url"))
	assert.Equal(t, "", got.Get("cn.ingestion_key"))
	assert.True(t, got.IsSet("cn.ingestion_key"))
	assert.Nil(t, got.Get("cn.queryflow_url"))
	assert.False(t, got.IsSet("cn.queryflow_url"))
	assert.Equal(t, "queryflow123", got.Get("cn.queryflow_key"))
	assert.Equal(t, "staging.cn.aws.com", got.Get("cn.staging_url"))
	assert.Nil(t, got.Get("cn.staging_key"))
	assert.False(t, got.IsSet("cn.staging_key"))
}

func Test_discovery_saveUrlAndAPIKey(t *testing.T) {
	const profile = "cn"

	tests := []struct {
		name          string
		input         string
		config        map[string]string
		standalone    bool
		inReader      io.Reader
		writePath     string
		component     string
		componentName string
		err           error
		expectKeys    map[string]string
	}{
		{
			name:       "Keep every existing value",
			input:      "\n\n",
			writePath:  t.TempDir(),
			standalone: true,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
				"cn.core_key": "core321",
			},
			expectKeys: map[string]string{
				"core_url": "http://localhost:8080",
				"core_key": "core321",
			},
			component:     "core",
			componentName: "Core",
		},
		{
			name:       "Set URL to empty, keep Key",
			input:      " \n",
			writePath:  t.TempDir(),
			standalone: true,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
				"cn.core_key": "core321",
			},
			expectKeys: map[string]string{
				"core_url": "",
				"core_key": "core321",
			},
			component:     "core",
			componentName: "Core",
		},
		{
			name:       "Set URL to new value, keep Key",
			input:      "http://discovery.core.cn\n\n",
			writePath:  t.TempDir(),
			standalone: false,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
				"cn.core_key": "core321",
			},
			expectKeys: map[string]string{
				"core_url": "http://discovery.core.cn",
				"core_key": "core321",
			},
			component:     "core",
			componentName: "Core",
		},
		{
			name:       "The user writes an End Of File while inputting the values",
			input:      "http://discovery.core.cn\ncore123",
			writePath:  t.TempDir(),
			standalone: false,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
				"cn.core_key": "core321",
			},
			expectKeys: map[string]string{
				"core_url": "http://discovery.core.cn",
				"core_key": "core123",
			},
			component:     "core",
			componentName: "Core",
		},
		{
			name:       "Key is nil, keep Key",
			input:      "http://discovery.core.cn\n\n",
			standalone: true,
			writePath:  t.TempDir(),
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
			},
			expectKeys: map[string]string{
				"core_url": "http://discovery.core.cn",
				"core_key": "",
			},
			component:     "core",
			componentName: "Core",
		},
		{
			name:       "Key is nil, change Key",
			input:      "http://discovery.core.cn\ncore123\n",
			writePath:  t.TempDir(),
			standalone: true,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
			},
			expectKeys: map[string]string{
				"core_url": "http://discovery.core.cn",
				"core_key": "core123",
			},
			component:     "core",
			componentName: "Core",
		},
		{
			name:          "Reading from the In IOStream fails to get URL",
			inReader:      testutils.ErrReader{Err: errors.New("read failed")},
			standalone:    true,
			writePath:     t.TempDir(),
			err:           NewErrorWithCause(ErrorExitCode, fmt.Errorf("read failed"), "Failed to get Core's URL"),
			component:     "core",
			componentName: "Core",
		},
		{
			name:          "Reading from the In IOStream fails to get Key",
			inReader:      io.MultiReader(strings.NewReader("http://discovery.core.cn\n"), testutils.ErrReader{Err: errors.New("read failed")}),
			standalone:    true,
			writePath:     t.TempDir(),
			err:           NewErrorWithCause(ErrorExitCode, fmt.Errorf("read failed"), "Failed to get Core's API key"),
			component:     "core",
			componentName: "Core",
		},
		{
			name:          "Invalid write location",
			input:         strings.Repeat("\n", 8),
			standalone:    false,
			writePath:     "doesnotexist",
			err:           NewErrorWithCause(ErrorExitCode, fmt.Errorf("open doesnotexist\\config.toml: The system cannot find the path specified."), "Failed to save Core's configuration"),
			component:     "core",
			componentName: "Core",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var in io.Reader
			if tc.inReader != nil {
				in = tc.inReader
			} else {
				in = strings.NewReader(tc.input)
			}

			out := &bytes.Buffer{}

			ios := iostreams.IOStreams{
				In:  in,
				Out: out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}

			d := NewDiscovery(&ios, vpr, tc.writePath)

			err := d.saveUrlAndAPIKey(profile, tc.component, tc.componentName, tc.standalone)
			if tc.standalone {
				assert.Contains(t, out.String(), fmt.Sprintf("Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile))
			}
			if tc.err != nil {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				if cliError, ok := err.(*Error); ok {
					cause := cliError.Cause
					if !errors.Is(cause, fs.ErrNotExist) {
						assert.EqualError(t, err, tc.err.Error())
					} else {
						tcError, _ := err.(*Error)
						assert.Equal(t, tcError.Message, cliError.Message)
						assert.Equal(t, tcError.ExitCode, cliError.ExitCode)
					}
				}
			} else {
				require.NoError(t, err)
				vpr, err := InitializeConfig(ios, tc.writePath)
				require.NoError(t, err)

				for k, expected := range tc.expectKeys {
					gotVal := vpr.GetString(profile + "." + k)
					require.Equal(t, expected, gotVal)
				}
			}
		})
	}
}

// Test_discovery_SaveCoreConfigFromUser tests the discovery.SaveCoreConfigFromUser() function.
func Test_discovery_SaveCoreConfigFromUser(t *testing.T) {
	const profile = "cn"

	tests := []struct {
		name       string
		input      string
		config     map[string]string
		standalone bool
		inReader   io.Reader
		writePath  string
		err        error
		expectKeys map[string]string
	}{
		{
			name:       "saveURLAndAPIKey returns no error",
			input:      "\n\n",
			writePath:  t.TempDir(),
			standalone: true,
			config: map[string]string{
				"cn.core_url": "http://localhost:12010",
				"cn.core_key": "core321",
			},
			expectKeys: map[string]string{
				"core_url": "http://localhost:12010",
				"core_key": "core321",
			},
		},
		{
			name:       "saveURLAndAPIKey returns an error",
			inReader:   testutils.ErrReader{Err: errors.New("read failed")},
			standalone: false,
			writePath:  t.TempDir(),
			err:        NewErrorWithCause(ErrorExitCode, fmt.Errorf("read failed"), "Failed to get Core's URL"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var in io.Reader
			if tc.inReader != nil {
				in = tc.inReader
			} else {
				in = strings.NewReader(tc.input)
			}

			out := &bytes.Buffer{}

			ios := iostreams.IOStreams{
				In:  in,
				Out: out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}

			d := NewDiscovery(&ios, vpr, tc.writePath)

			err := d.SaveCoreConfigFromUser(profile, true)
			if tc.standalone {
				assert.Contains(t, out.String(), fmt.Sprintf("Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile))
			}
			if tc.err != nil {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				vpr, err := InitializeConfig(ios, tc.writePath)
				require.NoError(t, err)

				for k, expected := range tc.expectKeys {
					gotVal := vpr.GetString(profile + "." + k)
					require.Equal(t, expected, gotVal)
				}
			}
		})
	}
}

// Test_discovery_SaveIngestionConfigFromUser tests the discovery.SaveIngestionConfigFromUser() function.
func Test_discovery_SaveIngestionConfigFromUser(t *testing.T) {
	const profile = "cn"

	tests := []struct {
		name       string
		input      string
		config     map[string]string
		standalone bool
		inReader   io.Reader
		writePath  string
		err        error
		expectKeys map[string]string
	}{
		{
			name:       "saveURLAndAPIKey returns no error",
			input:      "\n\n",
			writePath:  t.TempDir(),
			standalone: true,
			config: map[string]string{
				"cn.ingestion_url": "http://localhost:12030",
				"cn.ingestion_key": "ingestion321",
			},
			expectKeys: map[string]string{
				"ingestion_url": "http://localhost:12030",
				"ingestion_key": "ingestion321",
			},
		},
		{
			name:       "saveURLAndAPIKey returns an error",
			inReader:   testutils.ErrReader{Err: errors.New("read failed")},
			standalone: false,
			writePath:  t.TempDir(),
			err:        NewErrorWithCause(ErrorExitCode, fmt.Errorf("read failed"), "Failed to get Ingestion's URL"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var in io.Reader
			if tc.inReader != nil {
				in = tc.inReader
			} else {
				in = strings.NewReader(tc.input)
			}

			out := &bytes.Buffer{}

			ios := iostreams.IOStreams{
				In:  in,
				Out: out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}

			d := NewDiscovery(&ios, vpr, tc.writePath)

			err := d.SaveIngestionConfigFromUser(profile, true)
			if tc.standalone {
				assert.Contains(t, out.String(), fmt.Sprintf("Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile))
			}
			if tc.err != nil {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				vpr, err := InitializeConfig(ios, tc.writePath)
				require.NoError(t, err)

				for k, expected := range tc.expectKeys {
					gotVal := vpr.GetString(profile + "." + k)
					require.Equal(t, expected, gotVal)
				}
			}
		})
	}
}

// Test_discovery_SaveQueryFlowConfigFromUser tests discovery.SaveQueryFlowConfigFromUser() function.
func Test_discovery_SaveQueryFlowConfigFromUser(t *testing.T) {
	const profile = "cn"

	tests := []struct {
		name       string
		input      string
		config     map[string]string
		standalone bool
		inReader   io.Reader
		writePath  string
		err        error
		expectKeys map[string]string
	}{
		{
			name:       "saveURLAndAPIKey returns no error",
			input:      "\n\n",
			writePath:  t.TempDir(),
			standalone: true,
			config: map[string]string{
				"cn.queryflow_url": "http://localhost:12040",
				"cn.queryflow_key": "queryflow321",
			},
			expectKeys: map[string]string{
				"queryflow_url": "http://localhost:12040",
				"queryflow_key": "queryflow321",
			},
		},
		{
			name:       "saveURLAndAPIKey returns an error",
			inReader:   testutils.ErrReader{Err: errors.New("read failed")},
			standalone: false,
			writePath:  t.TempDir(),
			err:        NewErrorWithCause(ErrorExitCode, fmt.Errorf("read failed"), "Failed to get QueryFlow's URL"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var in io.Reader
			if tc.inReader != nil {
				in = tc.inReader
			} else {
				in = strings.NewReader(tc.input)
			}

			out := &bytes.Buffer{}

			ios := iostreams.IOStreams{
				In:  in,
				Out: out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}

			d := NewDiscovery(&ios, vpr, tc.writePath)

			err := d.SaveQueryFlowConfigFromUser(profile, true)
			if tc.standalone {
				assert.Contains(t, out.String(), fmt.Sprintf("Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile))
			}
			if tc.err != nil {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				vpr, err := InitializeConfig(ios, tc.writePath)
				require.NoError(t, err)

				for k, expected := range tc.expectKeys {
					gotVal := vpr.GetString(profile + "." + k)
					require.Equal(t, expected, gotVal)
				}
			}
		})
	}
}

// Test_discovery_SaveStagingConfigFromUser tests the discovery.SaveStagingConfigFromUser() function.
func Test_discovery_SaveStagingConfigFromUser(t *testing.T) {
	const profile = "cn"

	tests := []struct {
		name       string
		input      string
		config     map[string]string
		standalone bool
		inReader   io.Reader
		writePath  string
		err        error
		expectKeys map[string]string
	}{
		{
			name:       "saveURLAndAPIKey returns no error",
			input:      "\n\n",
			writePath:  t.TempDir(),
			standalone: true,
			config: map[string]string{
				"cn.staging_url": "http://localhost:12020",
				"cn.staging_key": "staging321",
			},
			expectKeys: map[string]string{
				"staging_url": "http://localhost:12020",
				"staging_key": "staging321",
			},
		},
		{
			name:       "saveURLAndAPIKey returns an error",
			inReader:   testutils.ErrReader{Err: errors.New("read failed")},
			standalone: false,
			writePath:  t.TempDir(),
			err:        NewErrorWithCause(ErrorExitCode, fmt.Errorf("read failed"), "Failed to get Staging's URL"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var in io.Reader
			if tc.inReader != nil {
				in = tc.inReader
			} else {
				in = strings.NewReader(tc.input)
			}

			out := &bytes.Buffer{}

			ios := iostreams.IOStreams{
				In:  in,
				Out: out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}

			d := NewDiscovery(&ios, vpr, tc.writePath)

			err := d.SaveStagingConfigFromUser(profile, true)
			if tc.standalone {
				assert.Contains(t, out.String(), fmt.Sprintf("Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n", profile))
			}
			if tc.err != nil {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				vpr, err := InitializeConfig(ios, tc.writePath)
				require.NoError(t, err)

				for k, expected := range tc.expectKeys {
					gotVal := vpr.GetString(profile + "." + k)
					require.Equal(t, expected, gotVal)
				}
			}
		})
	}
}

// Test_discovery_printConfig tests the discovery.PrintConfig() function.
func Test_discovery_printConfig(t *testing.T) {
	tests := []struct {
		name           string
		profile        string
		propertyName   string
		property       string
		sensitive      bool
		config         map[string]string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		{
			name:         "Print not sensitive value ",
			profile:      "cn",
			propertyName: "Core URL",
			property:     "core_url",
			sensitive:    false,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n", "Core URL", "http://localhost:8080"),
			outWriter:      nil,
			err:            nil,
		},
		{
			name:         "Print sensitive value",
			profile:      "cn",
			propertyName: "Core API Key",
			property:     "core_key",
			sensitive:    true,
			config: map[string]string{
				"cn.core_key": "discovery.key.core.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n", "Core API Key", obfuscate("discovery.key.core.cn")),
			outWriter:      nil,
			err:            nil,
		},
		{
			name:         "Do not print nil value",
			profile:      "cn",
			propertyName: "Core URL",
			property:     "core_url",
			sensitive:    false,
			config: map[string]string{
				"cn.core_key": "discovery.key.core.cn",
			},
			expectedOutput: "",
			outWriter:      nil,
			err:            nil,
		},
		{
			name:         "Printing fails",
			profile:      "cn",
			propertyName: "Core URL",
			property:     "core_url",
			sensitive:    false,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n", "Core URL", "http://localhost:8080"),
			outWriter:      testutils.ErrWriter{Err: errors.New("write failed")},
			err:            errors.New("write failed"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			var out io.Writer
			if tc.outWriter != nil {
				out = tc.outWriter
			} else {
				out = buf
			}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}

			d := NewDiscovery(&ios, vpr, "")

			err := d.printConfig(tc.profile, tc.propertyName, tc.property, tc.sensitive)

			if tc.err != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}

// Test_discovery_PrintCoreConfigToUser tests the discovery.PrintCoreToUser() function.
func Test_discovery_printURLAndAPIKey(t *testing.T) {
	tests := []struct {
		name           string
		profile        string
		sensitive      bool
		standalone     bool
		config         map[string]string
		component      string
		componentName  string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		{
			name:       "Print not standalone and not sensitive values",
			profile:    "cn",
			sensitive:  false,
			standalone: false,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
				"cn.core_key": "discovery.key.core.cn",
			},
			component:      "core",
			componentName:  "Core",
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n", "Core URL", "http://localhost:8080", "Core API Key", "discovery.key.core.cn"),
			outWriter:      nil,
			err:            nil,
		},
		{
			name:       "Print not standalone and sensitive value",
			profile:    "cn",
			sensitive:  true,
			standalone: false,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
				"cn.core_key": "discovery.key.core.cn",
			},
			component:      "core",
			componentName:  "Core",
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n", "Core URL", "http://localhost:8080", "Core API Key", obfuscate("discovery.key.core.cn")),
			outWriter:      nil,
			err:            nil,
		},
		{
			name:       "Print standalone and sensitive value",
			profile:    "cn",
			sensitive:  true,
			standalone: true,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
				"cn.core_key": "discovery.key.core.cn",
			},
			component:      "core",
			componentName:  "Core",
			expectedOutput: fmt.Sprintf("Showing the configuration of profile %q:\n\n%s: %q\n%s: %q\n", "cn", "Core URL", "http://localhost:8080", "Core API Key", obfuscate("discovery.key.core.cn")),
			outWriter:      nil,
			err:            nil,
		},
		{
			name:       "Print not standalone and nil value",
			profile:    "cn",
			sensitive:  false,
			standalone: false,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
			},
			component:      "core",
			componentName:  "Core",
			expectedOutput: fmt.Sprintf("%s: %q\n", "Core URL", "http://localhost:8080"),
			outWriter:      nil,
			err:            nil,
		},
		{
			name:       "Printing fails for URL",
			profile:    "cn",
			sensitive:  false,
			standalone: false,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
				"cn.core_key": "discovery.key.core.cn",
			},
			component:      "core",
			componentName:  "Core",
			expectedOutput: fmt.Sprintf("%s: %q\n", "Core URL", "http://localhost:8080"),
			outWriter:      testutils.ErrWriter{Err: errors.New("write failed")},
			err:            NewErrorWithCause(ErrorExitCode, fmt.Errorf("write failed"), "Could not print Core's URL"),
		},
		{
			name:       "Printing fails for API Key",
			profile:    "cn",
			sensitive:  false,
			standalone: false,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
				"cn.core_key": "discovery.key.core.cn",
			},
			component:      "core",
			componentName:  "Core",
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n", "Core URL", "http://localhost:8080", "Core API Key", "discovery.key.core.cn"),
			outWriter:      &testutils.FailOnNWriter{Writer: &bytes.Buffer{}, N: 2},
			err:            NewErrorWithCause(ErrorExitCode, fmt.Errorf("write failed"), "Could not print Core's API key"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			var out io.Writer
			if tc.outWriter != nil {
				out = tc.outWriter
			} else {
				out = buf
			}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}

			d := NewDiscovery(&ios, vpr, "")

			err := d.printURLAndAPIKey(tc.profile, tc.component, tc.componentName, tc.standalone, tc.sensitive)

			if tc.err != nil {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				if cliError, ok := err.(*Error); ok {
					cause := cliError.Cause
					if !errors.Is(cause, fs.ErrNotExist) {
						assert.EqualError(t, err, tc.err.Error())
					} else {
						tcError, _ := err.(*Error)
						assert.Equal(t, tcError.Message, cliError.Message)
						assert.Equal(t, tcError.ExitCode, cliError.ExitCode)
					}
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}

// Test_discovery_PrintCoreConfigToUser tests the discovery.PrintCoreToUser() function.
func Test_discovery_PrintCoreConfigToUser(t *testing.T) {
	tests := []struct {
		name           string
		profile        string
		sensitive      bool
		standalone     bool
		config         map[string]string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		{
			name:       "printURLAndAPIKey returns no error",
			profile:    "cn",
			sensitive:  false,
			standalone: false,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
				"cn.core_key": "discovery.key.core.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n", "Core URL", "http://localhost:8080", "Core API Key", "discovery.key.core.cn"),
			outWriter:      nil,
			err:            nil,
		},
		{
			name:       "printURLAndAPIKey returns error",
			profile:    "cn",
			sensitive:  false,
			standalone: false,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
				"cn.core_key": "discovery.key.core.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n", "Core URL", "http://localhost:8080"),
			outWriter:      testutils.ErrWriter{Err: errors.New("write failed")},
			err:            NewErrorWithCause(ErrorExitCode, fmt.Errorf("write failed"), "Could not print Core's URL"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			var out io.Writer
			if tc.outWriter != nil {
				out = tc.outWriter
			} else {
				out = buf
			}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}

			d := NewDiscovery(&ios, vpr, "")

			err := d.PrintCoreConfigToUser(tc.profile, tc.sensitive, tc.standalone)

			if tc.err != nil {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}

// Test_discovery_PrintIngestionConfigToUser tests the discovery.PrintIngestionConfigToUser() function.
func Test_discovery_PrintIngestionConfigToUser(t *testing.T) {
	tests := []struct {
		name           string
		profile        string
		sensitive      bool
		standalone     bool
		config         map[string]string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		{
			name:       "printURLAndAPIKey returns no error",
			profile:    "cn",
			sensitive:  false,
			standalone: false,
			config: map[string]string{
				"cn.ingestion_url": "http://localhost:8080",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n", "Ingestion URL", "http://localhost:8080", "Ingestion API Key", "discovery.key.ingestion.cn"),
			outWriter:      nil,
			err:            nil,
		},
		{
			name:       "printURLAndAPIKey returns error",
			profile:    "cn",
			sensitive:  false,
			standalone: false,
			config: map[string]string{
				"cn.ingestion_url": "http://localhost:8080",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n", "Ingestion URL", "http://localhost:8080"),
			outWriter:      testutils.ErrWriter{Err: errors.New("write failed")},
			err:            NewErrorWithCause(ErrorExitCode, fmt.Errorf("write failed"), "Could not print Ingestion's URL"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			var out io.Writer
			if tc.outWriter != nil {
				out = tc.outWriter
			} else {
				out = buf
			}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}

			d := NewDiscovery(&ios, vpr, "")

			err := d.PrintIngestionConfigToUser(tc.profile, tc.sensitive, tc.standalone)

			if tc.err != nil {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}

// Test_discovery_PrintQueryFlowConfigToUser tests the discovery.PrintQueryFlowConfigToUser() function.
func Test_discovery_PrintQueryFlowConfigToUser(t *testing.T) {
	tests := []struct {
		name           string
		profile        string
		sensitive      bool
		standalone     bool
		config         map[string]string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		{
			name:       "printURLAndAPIKey returns no error",
			profile:    "cn",
			sensitive:  false,
			standalone: false,
			config: map[string]string{
				"cn.queryflow_url": "http://localhost:8080",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n", "QueryFlow URL", "http://localhost:8080", "QueryFlow API Key", "discovery.key.queryflow.cn"),
			outWriter:      nil,
			err:            nil,
		},
		{
			name:       "printURLAndAPIKey returns no error",
			profile:    "cn",
			sensitive:  false,
			standalone: false,
			config: map[string]string{
				"cn.queryflow_url": "http://localhost:8080",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n", "QueryFlow URL", "http://localhost:8080"),
			outWriter:      testutils.ErrWriter{Err: errors.New("write failed")},
			err:            NewErrorWithCause(ErrorExitCode, fmt.Errorf("write failed"), "Could not print QueryFlow's URL"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			var out io.Writer
			if tc.outWriter != nil {
				out = tc.outWriter
			} else {
				out = buf
			}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}

			d := NewDiscovery(&ios, vpr, "")

			err := d.PrintQueryFlowConfigToUser(tc.profile, tc.sensitive, tc.standalone)

			if tc.err != nil {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}

// Test_discovery_PrintStagingConfigToUser tests the discovery.PrintStagingConfigToUser() function.
func Test_discovery_PrintStagingConfigToUser(t *testing.T) {
	tests := []struct {
		name           string
		profile        string
		sensitive      bool
		standalone     bool
		config         map[string]string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		{
			name:       "printURLAndAPIKey returns no error",
			profile:    "cn",
			sensitive:  false,
			standalone: false,
			config: map[string]string{
				"cn.staging_url": "http://localhost:8080",
				"cn.staging_key": "discovery.key.queryflow.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n", "Staging URL", "http://localhost:8080", "Staging API Key", "discovery.key.queryflow.cn"),
			outWriter:      nil,
			err:            nil,
		},
		{
			name:       "printURLAndAPIKey returns no error",
			profile:    "cn",
			sensitive:  false,
			standalone: false,
			config: map[string]string{
				"cn.staging_url": "http://localhost:8080",
				"cn.staging_key": "discovery.key.staging.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n", "Staging URL", "http://localhost:8080"),
			outWriter:      testutils.ErrWriter{Err: errors.New("write failed")},
			err:            NewErrorWithCause(ErrorExitCode, fmt.Errorf("write failed"), "Could not print Staging's URL"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			var out io.Writer
			if tc.outWriter != nil {
				out = tc.outWriter
			} else {
				out = buf
			}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}

			d := NewDiscovery(&ios, vpr, "")

			err := d.PrintStagingConfigToUser(tc.profile, tc.sensitive, tc.standalone)

			if tc.err != nil {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}

// Test_discovery_PrintConfigToUser tests the discovery.PrintConfigToUser() function.
func Test_discovery_PrintConfigToUser(t *testing.T) {
	tests := []struct {
		name           string
		profile        string
		sensitive      bool
		config         map[string]string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		{
			name:      "Print all values not sensitive",
			profile:   "cn",
			sensitive: false,
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
				"cn.queryflow_url": "http://localhost:12030",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
				"cn.staging_url":   "http://localhost:12040",
				"cn.staging_key":   "discovery.key.staging.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n", "Core URL", "http://localhost:12010", "Core API Key", "discovery.key.core.cn", "Ingestion URL", "http://localhost:12020", "Ingestion API Key", "discovery.key.ingestion.cn", "QueryFlow URL", "http://localhost:12030", "QueryFlow API Key", "discovery.key.queryflow.cn", "Staging URL", "http://localhost:12040", "Staging API Key", "discovery.key.staging.cn"),
			outWriter:      nil,
			err:            nil,
		},
		{
			name:      "Print all values sensitive",
			profile:   "cn",
			sensitive: true,
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
				"cn.queryflow_url": "http://localhost:12030",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
				"cn.staging_url":   "http://localhost:12040",
				"cn.staging_key":   "discovery.key.staging.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n", "Core URL", "http://localhost:12010", "Core API Key", obfuscate("discovery.key.core.cn"), "Ingestion URL", "http://localhost:12020", "Ingestion API Key", obfuscate("discovery.key.ingestion.cn"), "QueryFlow URL", "http://localhost:12030", "QueryFlow API Key", obfuscate("discovery.key.queryflow.cn"), "Staging URL", "http://localhost:12040", "Staging API Key", obfuscate("discovery.key.staging.cn")),
			outWriter:      nil,
			err:            nil,
		},
		{
			name:      "Print some values",
			profile:   "cn",
			sensitive: false,
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n%s: %q\n%s: %q\n", "Core URL", "http://localhost:12010", "Core API Key", "discovery.key.core.cn", "Ingestion URL", "http://localhost:12020", "QueryFlow API Key", "discovery.key.queryflow.cn"),
			outWriter:      nil,
			err:            nil,
		},
		{
			name:      "Print Fail on Printing Core Config",
			profile:   "cn",
			sensitive: false,
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
				"cn.queryflow_url": "http://localhost:12030",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
				"cn.staging_url":   "http://localhost:12040",
				"cn.staging_key":   "discovery.key.staging.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n", "Core URL", "http://localhost:12010", "Core API Key", "discovery.key.core.cn", "Ingestion URL", "http://localhost:12020", "Ingestion API Key", "discovery.key.ingestion.cn", "QueryFlow URL", "http://localhost:12030", "QueryFlow API Key", "discovery.key.queryflow.cn", "Staging URL", "http://localhost:12040", "Staging API Key", "discovery.key.staging.cn"),
			outWriter:      &testutils.FailOnNWriter{Writer: &bytes.Buffer{}, N: 2},
			err:            NewErrorWithCause(ErrorExitCode, fmt.Errorf("write failed"), "Could not print Core's URL"),
		},
		{
			name:      "Print Fail on Printing Ingestion Config",
			profile:   "cn",
			sensitive: false,
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
				"cn.queryflow_url": "http://localhost:12030",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
				"cn.staging_url":   "http://localhost:12040",
				"cn.staging_key":   "discovery.key.staging.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n", "Core URL", "http://localhost:12010", "Core API Key", "discovery.key.core.cn", "Ingestion URL", "http://localhost:12020", "Ingestion API Key", "discovery.key.ingestion.cn", "QueryFlow URL", "http://localhost:12030", "QueryFlow API Key", "discovery.key.queryflow.cn", "Staging URL", "http://localhost:12040", "Staging API Key", "discovery.key.staging.cn"),
			outWriter:      &testutils.FailOnNWriter{Writer: &bytes.Buffer{}, N: 4},
			err:            NewErrorWithCause(ErrorExitCode, fmt.Errorf("write failed"), "Could not print Ingestion's URL"),
		},
		{
			name:      "Print Fail on Printing QueryFlow Config",
			profile:   "cn",
			sensitive: false,
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
				"cn.queryflow_url": "http://localhost:12030",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
				"cn.staging_url":   "http://localhost:12040",
				"cn.staging_key":   "discovery.key.staging.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n", "Core URL", "http://localhost:12010", "Core API Key", "discovery.key.core.cn", "Ingestion URL", "http://localhost:12020", "Ingestion API Key", "discovery.key.ingestion.cn", "QueryFlow URL", "http://localhost:12030", "QueryFlow API Key", "discovery.key.queryflow.cn", "Staging URL", "http://localhost:12040", "Staging API Key", "discovery.key.staging.cn"),
			outWriter:      &testutils.FailOnNWriter{Writer: &bytes.Buffer{}, N: 6},
			err:            NewErrorWithCause(ErrorExitCode, fmt.Errorf("write failed"), "Could not print QueryFlow's URL"),
		},
		{
			name:      "Print Fail on Printing Staging Config",
			profile:   "cn",
			sensitive: false,
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
				"cn.queryflow_url": "http://localhost:12030",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
				"cn.staging_url":   "http://localhost:12040",
				"cn.staging_key":   "discovery.key.staging.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n", "Core URL", "http://localhost:12010", "Core API Key", "discovery.key.core.cn", "Ingestion URL", "http://localhost:12020", "Ingestion API Key", "discovery.key.ingestion.cn", "QueryFlow URL", "http://localhost:12030", "QueryFlow API Key", "discovery.key.queryflow.cn", "Staging URL", "http://localhost:12040", "Staging API Key", "discovery.key.staging.cn"),
			outWriter:      &testutils.FailOnNWriter{Writer: &bytes.Buffer{}, N: 8},
			err:            NewErrorWithCause(ErrorExitCode, fmt.Errorf("write failed"), "Could not print Staging's URL"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			var out io.Writer
			if tc.outWriter != nil {
				out = tc.outWriter
			} else {
				out = buf
			}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}

			d := NewDiscovery(&ios, vpr, "")

			err := d.PrintConfigToUser(tc.profile, tc.sensitive)

			if tc.err != nil {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Contains(t, buf.String(), fmt.Sprintf("Showing the configuration of profile %q:\n\n", tc.profile))
				require.Contains(t, buf.String(), tc.expectedOutput)
			}
		})
	}
}
