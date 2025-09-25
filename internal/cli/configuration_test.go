package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/fileutils"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_readConfigFile tests the readConfigFile() auxiliary function.
func Test_readConfigFile(t *testing.T) {
	tests := []struct {
		name           string
		baseName       string
		config         string
		expectedConfig map[string]string
		expectedBool   bool
		err            error
	}{
		// Working cases
		{
			name:     "Config file can be read",
			baseName: "config",
			config: `
[default]
core_url="http://localhost:8080"

[cn]
core_url="http://discovery.core.cn"	
`,
			expectedConfig: map[string]string{
				"default.core_url": "http://localhost:8080",
				"cn.core_url":      "http://discovery.core.cn",
			},
			expectedBool: true,
			err:          nil,
		},
		{
			name:           "File does not exist",
			baseName:       "fail",
			config:         ``,
			expectedConfig: map[string]string{},
			expectedBool:   false,
			err:            nil,
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
			expectedBool: true,
			err:          errors.New("While parsing config: toml: invalid character at start of key: {"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.baseName != "fail" {
				_, err := fileutils.CreateTemporaryFile(dir, tc.baseName+".toml", tc.config)
				if err != nil {
					t.Fatalf("Could not create temporary file: %s", err.Error())
				}
			}

			errStream := &bytes.Buffer{}
			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: os.Stdout,
				Err: errStream,
			}

			viper := viper.New()
			exists, err := readConfigFile(tc.baseName, dir, viper, &ios)
			assert.Equal(t, tc.expectedBool, exists)
			if tc.err != nil {
				assert.Contains(t, err.Error(), tc.err.Error())
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
				"default.core_url": "http://localhost:8080",
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
				"default.core_url": "http://localhost:8080",
				"default.core_key": "",
			},
			err: nil,
		},
		{
			name: "Reading the config file fails",
			config: `
{
  "default": {
    "core_url": "http://localhost:8080"
  },
  "cn": {
    "core_url": "http://discovery.core.cn"
  }
}
`,
			credentials: ``,
			expectedConfig: map[string]string{
				"default.core_url": "http://localhost:8080",
				"cn.core_url":      "http://discovery.core.cn",
				"default.core_key": "",
			},
			err: errors.New("While parsing config: toml: invalid character at start of key: {"),
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
				"default.core_url": "http://localhost:8080",
				"default.core_key": "",
				"cn.core_key":      "discovery.key.core.cn",
			},
			err: errors.New("While parsing config: toml: invalid character at start of key: {"),
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
				if err != nil {
					t.Fatalf("Could not create temporary file: %s", err.Error())
				}
			}

			if tc.credentials != "" {
				_, err := fileutils.CreateTemporaryFile(dir, "credentials.toml", tc.credentials)
				if err != nil {
					t.Fatalf("Could not create temporary file: %s", err.Error())
				}
			}

			viper, err := InitializeConfig(ios, dir)
			if tc.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err.Error())
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

// ErrReader is a struct that fails on any Read. It is used to mock when reading from an in IOStream fails.
type errReader struct{ err error }

// Read fails on any reading operation.
func (r errReader) Read(p []byte) (int, error) { return 0, r.err }

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
			inReader:  errReader{err: errors.New("read failed")},
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

	dir := t.TempDir()

	tests := []struct {
		name       string
		input      string
		inReader   io.Reader
		outWriter  io.Writer
		writePath  string
		err        error
		expectKeys map[string]string
	}{
		{
			name:      "Keep every existing value",
			input:     strings.Repeat("\n", 8),
			writePath: filepath.Join(dir, "test1.toml"),
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
			writePath: filepath.Join(dir, "test2.toml"),
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
			writePath: filepath.Join(dir, "test3.toml"),
			expectKeys: map[string]string{
				"core_url":      "http://discovery.core.cn",
				"core_key":      config["cn.core_key"],
				"ingestion_url": config["cn.ingestion_url"],
			},
		},
		{
			name:      "The user writes an End Of File while inputting the values",
			input:     "http://discovery.core.cn\ncore123",
			writePath: filepath.Join(dir, "test5.toml"),
			expectKeys: map[string]string{
				"core_url": "http://discovery.core.cn",
				"core_key": "core123",
			},
		},
		{
			name:      "Reading from the Core Config In IOStream fails",
			inReader:  errReader{err: errors.New("read failed")},
			writePath: filepath.Join(dir, "test7.toml"),
			err:       fmt.Errorf("read failed"),
		},
		{
			name:      "Reading from the Ingestion Config In IOStream fails",
			inReader:  io.MultiReader(strings.NewReader("http://discovery.core.cn\n\n"), errReader{err: errors.New("read failed")}),
			writePath: filepath.Join(dir, "test7.toml"),
			err:       fmt.Errorf("read failed"),
		},
		{
			name:      "Reading from the QueryFlow Config In IOStream fails",
			inReader:  io.MultiReader(strings.NewReader("http://discovery.core.cn\n\nhttp://discovery.ingestion.cn\n\n"), errReader{err: errors.New("read failed")}),
			writePath: filepath.Join(dir, "test7.toml"),
			err:       fmt.Errorf("read failed"),
		},
		{
			name:      "Reading from the Staging Config In IOStream fails",
			inReader:  io.MultiReader(strings.NewReader("http://discovery.core.cn\n\nhttp://discovery.ingestion.cn\n\nhttp://discovery.queryflow.cn\n\n"), errReader{err: errors.New("read failed")}),
			writePath: filepath.Join(dir, "test7.toml"),
			err:       fmt.Errorf("read failed"),
		},
		{
			name:      "Invalid write location",
			input:     strings.Repeat("\n", 8),
			writePath: filepath.Join(dir, "nonexistent", "config.toml"),
			err:       fmt.Errorf("cannot find the path specified"),
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
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err.Error())
				return
			} else {
				require.NoError(t, err)
				got := viper.New()
				got.SetConfigFile(tc.writePath)
				require.NoError(t, got.ReadInConfig())

				for k, want := range tc.expectKeys {
					gotVal := got.GetString(profile + "." + k)
					require.Equalf(t, want, gotVal, "key %q mismatch", k)
				}
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

	path := filepath.Join(dir, "config.toml")
	d := NewDiscovery(&ios, vpr, path)

	err := d.SaveConfigFromUser(profile)
	require.NoError(t, err)
	got := viper.New()
	got.SetConfigFile(path)
	require.NoError(t, got.ReadInConfig())

	assert.Equal(t, "http://localhost:8080", got.Get("cn.core_url"))
	assert.Equal(t, "core123", got.Get("cn.core_key"))
	assert.Equal(t, "http://localhost:8080", got.Get("cn.ingestion_url"))
	assert.Equal(t, "", got.Get("cn.ingestion_key"))
	assert.Nil(t, got.Get("cn.queryflow_url"))
	assert.Equal(t, "queryflow123", got.Get("cn.queryflow_key"))
	assert.Equal(t, "staging.cn.aws.com", got.Get("cn.staging_url"))
	assert.Nil(t, got.Get("cn.staging_key"))
}

// Test_discovery_SaveCoreConfigFromUser tests the discovery.SaveCoreConfigFromUser() function.
func Test_discovery_SaveCoreConfigFromUser(t *testing.T) {
	const profile = "cn"

	dir := t.TempDir()

	tests := []struct {
		name       string
		input      string
		config     map[string]string
		standalone bool
		inReader   io.Reader
		outWriter  io.Writer
		writePath  string
		err        error
		expectKeys map[string]string
	}{
		{
			name:       "Keep every existing value",
			input:      "\n\n",
			writePath:  filepath.Join(dir, "test1.toml"),
			standalone: true,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
				"cn.core_key": "core321",
			},
			expectKeys: map[string]string{
				"core_url": "http://localhost:8080",
				"core_key": "core321",
			},
		},
		{
			name:       "Set Core URL to empty, keep Core Key",
			input:      " \n",
			writePath:  filepath.Join(dir, "test2.toml"),
			standalone: true,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
				"cn.core_key": "core321",
			},
			expectKeys: map[string]string{
				"core_url": "",
				"core_key": "core321",
			},
		},
		{
			name:       "Set Core URL to new value, keep Core Key",
			input:      "http://discovery.core.cn\n\n",
			writePath:  filepath.Join(dir, "test3.toml"),
			standalone: false,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
				"cn.core_key": "core321",
			},
			expectKeys: map[string]string{
				"core_url": "http://discovery.core.cn",
				"core_key": "core321",
			},
		},
		{
			name:       "The user writes an End Of File while inputting the values",
			input:      "http://discovery.core.cn\ncore123",
			writePath:  filepath.Join(dir, "test5.toml"),
			standalone: false,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
				"cn.core_key": "core321",
			},
			expectKeys: map[string]string{
				"core_url": "http://discovery.core.cn",
				"core_key": "core123",
			},
		},
		{
			name:       "Core Key is nil, keep Core Key",
			input:      "http://discovery.core.cn\n\n",
			standalone: true,
			writePath:  filepath.Join(dir, "test3.toml"),
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
			},
			expectKeys: map[string]string{
				"core_url": "http://discovery.core.cn",
				"core_key": "",
			},
		},
		{
			name:       "Core Key is nil, change Core Key",
			input:      "http://discovery.core.cn\ncore123\n",
			writePath:  filepath.Join(dir, "test3.toml"),
			standalone: true,
			config: map[string]string{
				"cn.core_url": "http://localhost:8080",
			},
			expectKeys: map[string]string{
				"core_url": "http://discovery.core.cn",
				"core_key": "core123",
			},
		},
		{
			name:       "Reading from the In IOStream fails in Core URL",
			inReader:   errReader{err: errors.New("read failed")},
			standalone: true,
			writePath:  filepath.Join(dir, "test7.toml"),
			err:        fmt.Errorf("read failed"),
		},
		{
			name:       "Reading from the In IOStream fails in Core Key",
			inReader:   io.MultiReader(strings.NewReader("http://discovery.core.cn\n"), errReader{err: errors.New("read failed")}),
			standalone: true,
			writePath:  filepath.Join(dir, "test9.toml"),
			err:        fmt.Errorf("read failed"),
		},
		{
			name:       "Invalid write location",
			input:      strings.Repeat("\n", 8),
			standalone: false,
			writePath:  filepath.Join(dir, "nonexistent", "config.toml"),
			err:        fmt.Errorf("cannot find the path specified"),
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
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err.Error())
				return
			} else {
				require.NoError(t, err)
				got := viper.New()
				got.SetConfigFile(tc.writePath)
				require.NoError(t, got.ReadInConfig())

				for k, want := range tc.expectKeys {
					gotVal := got.GetString(profile + "." + k)
					require.Equalf(t, want, gotVal, "key %q mismatch", k)
				}
			}
		})
	}
}

// Test_discovery_SaveIngestionConfigFromUser tests the discovery.SaveIngestionConfigFromUser() function.
func Test_discovery_SaveIngestionConfigFromUser(t *testing.T) {
	const profile = "cn"

	dir := t.TempDir()

	tests := []struct {
		name       string
		input      string
		config     map[string]string
		standalone bool
		inReader   io.Reader
		outWriter  io.Writer
		writePath  string
		err        error
		expectKeys map[string]string
	}{
		{
			name:       "Keep every existing value",
			input:      "\n\n",
			writePath:  filepath.Join(dir, "test1.toml"),
			standalone: true,
			config: map[string]string{
				"cn.ingestion_url": "http://localhost:8080",
				"cn.ingestion_key": "ingestion321",
			},
			expectKeys: map[string]string{
				"ingestion_url": "http://localhost:8080",
				"ingestion_key": "ingestion321",
			},
		},
		{
			name:       "Set Ingestion URL to empty, keep Ingestion Key",
			input:      " \n",
			writePath:  filepath.Join(dir, "test2.toml"),
			standalone: true,
			config: map[string]string{
				"cn.ingestion_url": "http://localhost:8080",
				"cn.ingestion_key": "ingestion321",
			},
			expectKeys: map[string]string{
				"ingestion_url": "",
				"ingestion_key": "ingestion321",
			},
		},
		{
			name:       "Set Ingestion URL to new value, keep Ingestion Key",
			input:      "http://discovery.ingestion.cn\n\n",
			writePath:  filepath.Join(dir, "test3.toml"),
			standalone: false,
			config: map[string]string{
				"cn.ingestion_url": "http://localhost:8080",
				"cn.ingestion_key": "ingestion321",
			},
			expectKeys: map[string]string{
				"ingestion_url": "http://discovery.ingestion.cn",
				"ingestion_key": "ingestion321",
			},
		},
		{
			name:       "The user writes an End Of File while inputting the values",
			input:      "http://discovery.ingestion.cn\ningestion123",
			writePath:  filepath.Join(dir, "test5.toml"),
			standalone: false,
			config: map[string]string{
				"cn.ingestion_url": "http://localhost:8080",
				"cn.ingestion_key": "ingestion321",
			},
			expectKeys: map[string]string{
				"ingestion_url": "http://discovery.ingestion.cn",
				"ingestion_key": "ingestion123",
			},
		},
		{
			name:       "Ingestion Key is nil, keep Ingestion Key",
			input:      "http://discovery.ingestion.cn\n\n",
			standalone: true,
			writePath:  filepath.Join(dir, "test3.toml"),
			config: map[string]string{
				"cn.ingestion_url": "http://localhost:8080",
			},
			expectKeys: map[string]string{
				"ingestion_url": "http://discovery.ingestion.cn",
				"ingestion_key": "",
			},
		},
		{
			name:       "Ingestion Key is nil, change Ingestion Key",
			input:      "http://discovery.ingestion.cn\ningestion123\n",
			writePath:  filepath.Join(dir, "test3.toml"),
			standalone: true,
			config: map[string]string{
				"cn.ingestion_url": "http://localhost:8080",
			},
			expectKeys: map[string]string{
				"ingestion_url": "http://discovery.ingestion.cn",
				"ingestion_key": "ingestion123",
			},
		},
		{
			name:       "Reading from the In IOStream fails in Ingestion URL",
			inReader:   errReader{err: errors.New("read failed")},
			standalone: true,
			writePath:  filepath.Join(dir, "test7.toml"),
			err:        fmt.Errorf("read failed"),
		},
		{
			name:       "Reading from the In IOStream fails in Ingestion Key",
			inReader:   io.MultiReader(strings.NewReader("http://discovery.ingestion.cn\n"), errReader{err: errors.New("read failed")}),
			standalone: true,
			writePath:  filepath.Join(dir, "test9.toml"),
			err:        fmt.Errorf("read failed"),
		},
		{
			name:       "Invalid write location",
			input:      strings.Repeat("\n", 8),
			standalone: false,
			writePath:  filepath.Join(dir, "nonexistent", "config.toml"),
			err:        fmt.Errorf("cannot find the path specified"),
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
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err.Error())
				return
			} else {
				require.NoError(t, err)
				got := viper.New()
				got.SetConfigFile(tc.writePath)
				require.NoError(t, got.ReadInConfig())

				for k, want := range tc.expectKeys {
					gotVal := got.GetString(profile + "." + k)
					require.Equalf(t, want, gotVal, "key %q mismatch", k)
				}
			}
		})
	}
}

// Test_discovery_SaveQueryFlowConfigFromUser tests discovery.SaveQueryFlowConfigFromUser() function.
func Test_discovery_SaveQueryFlowConfigFromUser(t *testing.T) {
	const profile = "cn"

	dir := t.TempDir()

	tests := []struct {
		name       string
		input      string
		config     map[string]string
		standalone bool
		inReader   io.Reader
		outWriter  io.Writer
		writePath  string
		err        error
		expectKeys map[string]string
	}{
		{
			name:       "Keep every existing value",
			input:      "\n\n",
			writePath:  filepath.Join(dir, "test1.toml"),
			standalone: true,
			config: map[string]string{
				"cn.queryflow_url": "http://localhost:8080",
				"cn.queryflow_key": "queryflow321",
			},
			expectKeys: map[string]string{
				"queryflow_url": "http://localhost:8080",
				"queryflow_key": "queryflow321",
			},
		},
		{
			name:       "Set QueryFlow URL to empty, keep QueryFlow Key",
			input:      " \n",
			writePath:  filepath.Join(dir, "test2.toml"),
			standalone: true,
			config: map[string]string{
				"cn.queryflow_url": "http://localhost:8080",
				"cn.queryflow_key": "queryflow321",
			},
			expectKeys: map[string]string{
				"queryflow_url": "",
				"queryflow_key": "queryflow321",
			},
		},
		{
			name:       "Set QueryFlow URL to new value, keep QueryFlow Key",
			input:      "http://discovery.queryflow.cn\n\n",
			writePath:  filepath.Join(dir, "test3.toml"),
			standalone: false,
			config: map[string]string{
				"cn.queryflow_url": "http://localhost:8080",
				"cn.queryflow_key": "queryflow321",
			},
			expectKeys: map[string]string{
				"queryflow_url": "http://discovery.queryflow.cn",
				"queryflow_key": "queryflow321",
			},
		},
		{
			name:       "The user writes an End Of File while inputting the values",
			input:      "http://discovery.queryflow.cn\nqueryflow123",
			writePath:  filepath.Join(dir, "test5.toml"),
			standalone: false,
			config: map[string]string{
				"cn.queryflow_url": "http://localhost:8080",
				"cn.queryflow_key": "queryflow321",
			},
			expectKeys: map[string]string{
				"queryflow_url": "http://discovery.queryflow.cn",
				"queryflow_key": "queryflow123",
			},
		},
		{
			name:       "QueryFlow Key is nil, keep QueryFlow Key",
			input:      "http://discovery.queryflow.cn\n\n",
			standalone: true,
			writePath:  filepath.Join(dir, "test3.toml"),
			config: map[string]string{
				"cn.queryflow_url": "http://localhost:8080",
			},
			expectKeys: map[string]string{
				"queryflow_url": "http://discovery.queryflow.cn",
				"queryflow_key": "",
			},
		},
		{
			name:       "QueryFlow Key is nil, change QueryFlow Key",
			input:      "http://discovery.queryflow.cn\nqueryflow123\n",
			writePath:  filepath.Join(dir, "test3.toml"),
			standalone: true,
			config: map[string]string{
				"cn.queryflow_url": "http://localhost:8080",
			},
			expectKeys: map[string]string{
				"queryflow_url": "http://discovery.queryflow.cn",
				"queryflow_key": "queryflow123",
			},
		},
		{
			name:       "Reading from the In IOStream fails in QueryFlow URL",
			inReader:   errReader{err: errors.New("read failed")},
			standalone: true,
			writePath:  filepath.Join(dir, "test7.toml"),
			err:        fmt.Errorf("read failed"),
		},
		{
			name:       "Reading from the In IOStream fails in QueryFlow Key",
			inReader:   io.MultiReader(strings.NewReader("http://discovery.queryflow.cn\n"), errReader{err: errors.New("read failed")}),
			standalone: true,
			writePath:  filepath.Join(dir, "test9.toml"),
			err:        fmt.Errorf("read failed"),
		},
		{
			name:       "Invalid write location",
			input:      strings.Repeat("\n", 8),
			standalone: false,
			writePath:  filepath.Join(dir, "nonexistent", "config.toml"),
			err:        fmt.Errorf("cannot find the path specified"),
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
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err.Error())
				return
			} else {
				require.NoError(t, err)
				got := viper.New()
				got.SetConfigFile(tc.writePath)
				require.NoError(t, got.ReadInConfig())

				for k, want := range tc.expectKeys {
					gotVal := got.GetString(profile + "." + k)
					require.Equalf(t, want, gotVal, "key %q mismatch", k)
				}
			}
		})
	}
}

// Test_discovery_SaveStagingConfigFromUser tests the discovery.SaveStagingConfigFromUser() function.
func Test_discovery_SaveStagingConfigFromUser(t *testing.T) {
	const profile = "cn"

	dir := t.TempDir()

	tests := []struct {
		name       string
		input      string
		config     map[string]string
		standalone bool
		inReader   io.Reader
		outWriter  io.Writer
		writePath  string
		err        error
		expectKeys map[string]string
	}{
		{
			name:       "Keep every existing value",
			input:      "\n\n",
			writePath:  filepath.Join(dir, "test1.toml"),
			standalone: true,
			config: map[string]string{
				"cn.staging_url": "http://localhost:8080",
				"cn.staging_key": "staging321",
			},
			expectKeys: map[string]string{
				"staging_url": "http://localhost:8080",
				"staging_key": "staging321",
			},
		},
		{
			name:       "Set Staging URL to empty, keep Staging Key",
			input:      " \n",
			writePath:  filepath.Join(dir, "test2.toml"),
			standalone: true,
			config: map[string]string{
				"cn.staging_url": "http://localhost:8080",
				"cn.staging_key": "staging321",
			},
			expectKeys: map[string]string{
				"staging_url": "",
				"staging_key": "staging321",
			},
		},
		{
			name:       "Set Staging URL to new value, keep Staging Key",
			input:      "http://discovery.staging.cn\n\n",
			writePath:  filepath.Join(dir, "test3.toml"),
			standalone: false,
			config: map[string]string{
				"cn.staging_url": "http://localhost:8080",
				"cn.staging_key": "staging321",
			},
			expectKeys: map[string]string{
				"staging_url": "http://discovery.staging.cn",
				"staging_key": "staging321",
			},
		},
		{
			name:       "The user writes an End Of File while inputting the values",
			input:      "http://discovery.staging.cn\nstaging123",
			writePath:  filepath.Join(dir, "test5.toml"),
			standalone: false,
			config: map[string]string{
				"cn.staging_url": "http://localhost:8080",
				"cn.staging_key": "staging321",
			},
			expectKeys: map[string]string{
				"staging_url": "http://discovery.staging.cn",
				"staging_key": "staging123",
			},
		},
		{
			name:       "Staging Key is nil, keep Staging Key",
			input:      "http://discovery.staging.cn\n\n",
			standalone: true,
			writePath:  filepath.Join(dir, "test3.toml"),
			config: map[string]string{
				"cn.staging_url": "http://localhost:8080",
			},
			expectKeys: map[string]string{
				"staging_url": "http://discovery.staging.cn",
				"staging_key": "",
			},
		},
		{
			name:       "Staging Key is nil, change Staging Key",
			input:      "http://discovery.staging.cn\nstaging123\n",
			writePath:  filepath.Join(dir, "test3.toml"),
			standalone: true,
			config: map[string]string{
				"cn.staging_url": "http://localhost:8080",
			},
			expectKeys: map[string]string{
				"staging_url": "http://discovery.staging.cn",
				"staging_key": "staging123",
			},
		},
		{
			name:       "Reading from the In IOStream fails in Staging URL",
			inReader:   errReader{err: errors.New("read failed")},
			standalone: true,
			writePath:  filepath.Join(dir, "test7.toml"),
			err:        fmt.Errorf("read failed"),
		},
		{
			name:       "Reading from the In IOStream fails in Staging Key",
			inReader:   io.MultiReader(strings.NewReader("http://discovery.staging.cn\n"), errReader{err: errors.New("read failed")}),
			standalone: true,
			writePath:  filepath.Join(dir, "test9.toml"),
			err:        fmt.Errorf("read failed"),
		},
		{
			name:       "Invalid write location",
			input:      strings.Repeat("\n", 8),
			standalone: false,
			writePath:  filepath.Join(dir, "nonexistent", "config.toml"),
			err:        fmt.Errorf("cannot find the path specified"),
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
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err.Error())
				return
			} else {
				require.NoError(t, err)
				got := viper.New()
				got.SetConfigFile(tc.writePath)
				require.NoError(t, got.ReadInConfig())

				for k, want := range tc.expectKeys {
					gotVal := got.GetString(profile + "." + k)
					require.Equalf(t, want, gotVal, "key %q mismatch", k)
				}
			}
		})
	}
}
