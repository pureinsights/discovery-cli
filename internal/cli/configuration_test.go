package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/stretchr/testify/assert"
)

func Test_InitializeConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         string
		credentials    string
		expectedConfig map[string]string
		err            error
	}{
		// Working cases
		{
			name: "There config and credentials files",
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

			dir := os.TempDir()
			if tc.config != "" {
				config, err := os.Create(filepath.Join(dir, "config.toml"))
				if err != nil {
					t.Fatalf("Could not create temporary file: %s", err.Error())
				}
				defer config.Close()
				if _, err := config.Write([]byte(tc.config)); err != nil {
					t.Fatalf("Could not write to file: %s", err.Error())
				}
			}

			if tc.credentials != "" {
				credentials, err := os.Create(filepath.Join(dir, "credentials.toml"))
				if err != nil {
					t.Fatalf("Could not create temporary file: %s", err.Error())
				}
				defer credentials.Close()
				if _, err := credentials.Write([]byte(tc.credentials)); err != nil {
					t.Fatalf("Could not write to file: %s", err.Error())
				}
			}

			viper, err := InitializeConfig(ios, dir)
			if tc.err != nil {
				assert.Contains(t, err.Error(), tc.err.Error())
			} else {
				for k, v := range tc.expectedConfig {
					assert.Equal(t, v, viper.GetString(k))
				}
			}
		})
	}
}
