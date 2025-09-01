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
	assert.Equal(t, apiKey, qec.enabler.client.ApiKey)
	assert.Equal(t, url+"/endpoint", qec.enabler.client.client.BaseURL)
}

// Test_queryFlow_Processors tests the queryFlow.Processors() function
func Test_queryFlow_Processors(t *testing.T) {
	q := NewQueryFlow("http://localhost:8080/v2", "Api Key")
	qpc := q.Processors()

	assert.Equal(t, q.ApiKey, qpc.crud.client.ApiKey)
	assert.Equal(t, q.Url+"/processor", qpc.crud.client.client.BaseURL)
	assert.Equal(t, q.ApiKey, qpc.cloner.client.ApiKey)
	assert.Equal(t, q.Url+"/processor", qpc.cloner.client.client.BaseURL)
}

// Test_queryFlow_Endpoints tests the queryFlow.Endpoints() function
func Test_queryFlow_Endpoints(t *testing.T) {
	q := NewQueryFlow("http://localhost:8080/v2", "Api Key")
	qec := q.Endpoints()

	assert.Equal(t, q.ApiKey, qec.crud.client.ApiKey)
	assert.Equal(t, q.Url+"/endpoint", qec.crud.client.client.BaseURL)
	assert.Equal(t, q.ApiKey, qec.cloner.client.ApiKey)
	assert.Equal(t, q.Url+"/endpoint", qec.cloner.client.client.BaseURL)
	assert.Equal(t, q.ApiKey, qec.enabler.client.ApiKey)
	assert.Equal(t, q.Url+"/endpoint", qec.enabler.client.client.BaseURL)
}

// Test_queryFlow_BackupRestore tests the core.BackupRestore() function
func Test_queryFlow_BackupRestore(t *testing.T) {
	q := NewQueryFlow("http://localhost:8088/v2", "Api Key")
	bc := q.BackupRestore()

	assert.Equal(t, q.ApiKey, bc.ApiKey)
	assert.Equal(t, q.Url, bc.client.client.BaseURL)
}
