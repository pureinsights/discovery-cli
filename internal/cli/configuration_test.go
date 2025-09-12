package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/fileutils"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_readConfigFile(t *testing.T) {
	tests := []struct {
		name           string
		baseName       string
		config         string
		expectedConfig map[string]string
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
			err: nil,
		},
		{
			name:           "File does not exist",
			baseName:       "fail",
			config:         ``,
			expectedConfig: map[string]string{},
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
			err: errors.New("While parsing config: toml: invalid character at start of key: {"),
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
			err = readConfigFile(tc.baseName, dir, viper, &ios)
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
core_url="http://localhost:8080"

[cn]
core_url="http://discovery.core.cn"	
`,
			credentials: `
[default]
core_key=""

[cn]
core_key="discovery.key.core.cn"
`,
			expectedConfig: map[string]string{
				"default.core_url": "http://localhost:8080",
				"cn.core_url":      "http://discovery.core.cn",
				"default.core_key": "",
				"cn.core_key":      "discovery.key.core.cn",
			},
			err: nil,
		},
		{
			name: "There is only a config file",
			config: `
[default]
core_url="http://localhost:8080"

[cn]
core_url="http://discovery.core.cn"	
`,
			credentials: ``,
			expectedConfig: map[string]string{
				"default.core_url": "http://localhost:8080",
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
core_key=""

[cn]
core_key="discovery.key.core.cn"
`,
			expectedConfig: map[string]string{
				"default.core_url": "http://localhost:8080",
				"default.core_key": "",
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
