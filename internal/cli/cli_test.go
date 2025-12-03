package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// Test_discovery_IOStreams tests the discover.IOStreams() function.
func Test_discovery_IOStreams(t *testing.T) {
	io := iostreams.IOStreams{
		In:  strings.NewReader("Test Reader"),
		Out: &bytes.Buffer{},
		Err: &bytes.Buffer{},
	}
	vpr := viper.New()
	discovery := NewDiscovery(&io, vpr, "")

	assert.Equal(t, &io, discovery.IOStreams())
}

// Test_discovery_Config tests the discovery.Config() function
func Test_discovery_Config(t *testing.T) {
	io := iostreams.IOStreams{
		In:  strings.NewReader("Test Reader"),
		Out: &bytes.Buffer{},
		Err: &bytes.Buffer{},
	}
	vpr := viper.New()
	discovery := NewDiscovery(&io, vpr, "")

	assert.Equal(t, vpr, discovery.Config())
}

// Test_discovery_ConfigPath tests the discovery.ConfigPath() function
func Test_discovery_ConfigPath(t *testing.T) {
	io := iostreams.IOStreams{
		In:  strings.NewReader("Test Reader"),
		Out: &bytes.Buffer{},
		Err: &bytes.Buffer{},
	}
	vpr := viper.New()
	configPath := "testFiles/configtest.toml"
	discovery := NewDiscovery(&io, vpr, configPath)

	assert.Equal(t, configPath, discovery.ConfigPath())
}

// TestNewDiscovery tests the NewDiscovery() constructor.
func TestNewDiscovery(t *testing.T) {
	io := iostreams.IOStreams{
		In:  strings.NewReader("Test Reader"),
		Out: &bytes.Buffer{},
		Err: &bytes.Buffer{},
	}
	vpr := viper.New()
	configPath := "testFiles/configtest.toml"
	discovery := NewDiscovery(&io, vpr, configPath)

	assert.Equal(t, &io, discovery.IOStreams())
	assert.Equal(t, vpr, discovery.Config())
	assert.Equal(t, configPath, discovery.ConfigPath())
}
