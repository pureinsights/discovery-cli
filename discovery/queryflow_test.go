package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_newQueryFlowProcessorsClient test the queryFlowProcessorsClient's constructor
func Test_newQueryFlowProcessorsClient(t *testing.T) {
	url := "http://localhost:8088/v2"
	apiKey := "Api Key"
	qpc := newQueryFlowProcessorsClient(url, apiKey)

	assert.Equal(t, apiKey, qpc.crud.client.ApiKey)
	assert.Equal(t, url+"/processor", qpc.crud.client.client.BaseURL)
	assert.Equal(t, apiKey, qpc.cloner.client.ApiKey)
	assert.Equal(t, url+"/processor", qpc.cloner.client.client.BaseURL)
}

// Test_newEndpointsClient tests the constructor of endpointsClients.
func Test_newEndpointsClient(t *testing.T) {
	url := "http://localhost:8088/v2"
	apiKey := "Api Key"
	qec := newEndpointsClient(url, apiKey)

	assert.Equal(t, apiKey, qec.crud.client.ApiKey)
	assert.Equal(t, url+"/endpoint", qec.crud.client.client.BaseURL)
	assert.Equal(t, apiKey, qec.cloner.client.ApiKey)
	assert.Equal(t, url+"/endpoint", qec.cloner.client.client.BaseURL)
}
