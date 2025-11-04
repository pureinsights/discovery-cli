package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetCommandConfig tests the GetCommandConfig() function.
func TestGetCommandConfig(t *testing.T) {
	profile := "default"
	output := "json"
	componentName := "Core"
	url := "core_url"
	key := "core_key"

	commandConfig := GetCommandConfig(profile, output, componentName, url, key)
	assert.Equal(t, profile, commandConfig.profile)
	assert.Equal(t, output, commandConfig.output)
	assert.Equal(t, componentName, commandConfig.componentName)
	assert.Equal(t, url, commandConfig.url)
	assert.Equal(t, key, commandConfig.apiKey)
}

// TestCheckCredentials tests the CheckCredentials function.
func TestCheckCredentials(t *testing.T) {
	tests := []struct {
		name          string
		profile       string
		componentName string
		url           string
		key           string
		config        map[string]string
		err           error
	}{
		// Working case
		{
			name:          "All the properties are set",
			profile:       "default",
			url:           "core_url",
			key:           "core_key",
			componentName: "Core",
			config: map[string]string{
				"default.core_url": "http://localhost:12010/v2",
				"default.core_key": "http://discovery.core.cn",
			},
			err: nil,
		},

		// Error cases
		{
			name:          "URL is not set",
			profile:       "default",
			url:           "core_url",
			key:           "core_key",
			componentName: "Core",
			config: map[string]string{
				"default.core_key": "http://discovery.core.cn",
			},
			err: cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery core config --profile {profile}"),
		},
		{
			name:          "API Key is not set",
			profile:       "default",
			url:           "core_url",
			key:           "core_key",
			componentName: "Core",
			config: map[string]string{
				"default.core_url": "http://discovery.core.cn",
			},
			err: cli.NewError(cli.ErrorExitCode, "The Discovery Core API key is missing for profile \"default\".\nTo set the API key for the Discovery Core API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery core config --profile {profile}"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			io := iostreams.IOStreams{
				In:  strings.NewReader("Test Reader"),
				Out: &bytes.Buffer{},
				Err: &bytes.Buffer{},
			}

			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}
			d := cli.NewDiscovery(&io, vpr, "")
			err := CheckCredentials(d, tc.profile, tc.componentName, tc.url, tc.key)
			if tc.err != nil {
				require.Error(t, err)
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
