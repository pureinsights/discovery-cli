package mocks

import (
	"net/http"
	"testing"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestWorkingStagingContentController_Count tests the WorkingStagingContentController.Count() mock.
func TestWorkingStagingContentController_Count(t *testing.T) {
	c := new(WorkingStagingContentController)
	result, err := c.Count(gjson.Result{})

	require.NoError(t, err)
	assert.Equal(t, gjson.Parse(`{"total": 2}`), result)
}

// TestWorkingStagingContentControllerNoContent_Count tests the WorkingStagingContentControllerNoContent.Count() mock.
func TestWorkingStagingContentControllerNoContent_Count(t *testing.T) {
	c := new(WorkingStagingContentControllerNoContent)
	result, err := c.Count(gjson.Result{})

	require.NoError(t, err)
	assert.Equal(t, gjson.Parse(`{"total": 0}`), result)
}

// TestFailingStagingContentController_Count tests the FailingStagingContentController.Count() mock.
func TestFailingStagingContentController_Count(t *testing.T) {
	c := new(FailingStagingContentController)
	_, err := c.Count(gjson.Result{})

	require.Error(t, err)
	var discoveryErr discoveryPackage.Error
	require.ErrorAs(t, err, &discoveryErr)
	assert.Equal(t, http.StatusNotFound, discoveryErr.Status)
}
