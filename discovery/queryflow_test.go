package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConstructors tests all of the constructors of the subclient structs.
func Test_newQueryFlowProcessorsClient(t *testing.T) {
	c := newClient("http://localhost:8088/v2/", "Api Key")
	qpc := newQueryFlowProcessorsClient(c)

	assert.Equal(t, c.ApiKey, qpc.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/processor", qpc.client.client.BaseURL)
}

// Test_newEndpointsClient tests the Ping method.
func Test_newEndpointsClient(t *testing.T) {
	c := newClient("http://localhost:8088/v2/", "Api Key")
	qec := newEndpointsClient(c)

	assert.Equal(t, c.ApiKey, qec.crud.ApiKey)
	assert.Equal(t, c.client.BaseURL+"/endpoint", qec.crud.client.client.BaseURL)
}
