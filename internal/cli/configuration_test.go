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
			dir, err := os.MkdirTemp("", "temp-")
			if err != nil {
				t.Fatal(err)
			}
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

			dir, err := os.MkdirTemp("", "temp-")
			if err != nil {
				t.Fatal(err)
			}
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

// ErrWriter is a struct that fails on any Write. It is used to mock when writing to an out IOStream fails.
type errWriter struct{ err error }

// Write fails on any writing operation.
func (w errWriter) Write(p []byte) (int, error) { return 0, w.err }

// ErrReader is a struct that fails on any Read. It is used to mock when reading from an in IOStream fails.
type errReader struct{ err error }

// Read fails on any reading operation.
func (r errReader) Read(p []byte) (int, error) { return 0, r.err }

// TestSaveConfigFromUser_AllConfigPresent tests the SaveConfigFromUser when there is a configuration for every possible URL and API Key
func TestSaveConfigFromUser_AllConfigPresent(t *testing.T) {
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

	dir, err := os.MkdirTemp("", "temp-")
	if err != nil {
		t.Fatal(err)
	}

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
			name:      "Set Core URL to empty, keep the rest",
			input:     " \n" + strings.Repeat("\n", 7),
			writePath: filepath.Join(dir, "test2.toml"),
			expectKeys: map[string]string{
				"core_url":      "",
				"core_key":      config["cn.core_key"],
				"ingestion_url": config["cn.ingestion_url"],
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
			input:     "http://discovery.core.cn\ncore123 ",
			writePath: filepath.Join(dir, "test5.toml"),
			expectKeys: map[string]string{
				"core_url": "http://discovery.core.cn",
				"core_key": config["cn.core_key"],
			},
		},
		{
			name:      "Printing to the Out IOStream fails",
			input:     "\n",
			outWriter: errWriter{err: errors.New("write failed")},
			writePath: filepath.Join(dir, "test6.toml"),
			err:       fmt.Errorf("write failed"),
		},
		{
			name:      "Reading from the In IOStream fails",
			inReader:  errReader{err: errors.New("read failed")},
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

			var out bytes.Buffer
			var outW io.Writer = &out
			if tc.outWriter != nil {
				outW = tc.outWriter
			}

			ios := iostreams.IOStreams{
				In:  in,
				Out: outW,
				Err: os.Stderr,
			}

			vpr := viper.New()
			for k, v := range config {
				vpr.Set(k, v)
			}

			d := NewDiscovery(ios, vpr)

			err = d.SaveConfigFromUser(ios, profile, tc.writePath)
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

// TestSaveConfigFromUser_NotAllConfigPresent tests the SaveConfigFromUser function when there are some properties with no explicit values set.
func TestSaveConfigFromUser_NotAllConfigPresent(t *testing.T) {
	const profile = "cn"
	config := map[string]string{
		"cn.core_url":      "http://localhost:8080",
		"cn.staging_url":   "http://localhost:8081",
		"cn.core_key":      "core321",
		"cn.ingestion_key": "ingestion432",
	}

	dir, err := os.MkdirTemp("", "temp-")
	if err != nil {
		t.Fatal(err)
	}

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

	d := NewDiscovery(ios, vpr)
	path := filepath.Join(dir, "config.toml")
	err = d.SaveConfigFromUser(ios, profile, path)
	require.NoError(t, err)
	got := viper.New()
	got.SetConfigFile(path)
	require.NoError(t, got.ReadInConfig())

	assert.Equal(t, "http://localhost:8080", got.Get("cn.core_url"))
	assert.Equal(t, "core123", got.Get("cn.core_key"))
	assert.Equal(t, "http://localhost:8080", got.Get("cn.ingestion_url"))
	assert.Equal(t, "", got.Get("cn.ingestion_key"))
	assert.Nil(t, got.Get("cn.queryflow_url"))
	assert.Contains(t, outStream.String(), "There is no value set for this property")
	assert.Equal(t, "queryflow123", got.Get("cn.queryflow_key"))
	assert.Equal(t, "staging.cn.aws.com", got.Get("cn.staging_url"))
	assert.Nil(t, got.Get("cn.staging_key"))
}
