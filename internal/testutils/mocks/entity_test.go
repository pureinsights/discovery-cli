package mocks

import (
	"net/http"
	"testing"

	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestWorkingContentController_Count tests the WorkingContentController.Count() mock.
func TestWorkingContentController_Count(t *testing.T) {
	c := new(WorkingContentController)
	result, err := c.Count(gjson.Result{})

	require.NoError(t, err)
	assert.Equal(t, gjson.Parse(`{"total": 10}`), result)
}

// TestFailingContentController_Count tests the FailingContentController.Count() mock.
func TestFailingContentController_Count(t *testing.T) {
	c := new(FailingContentController)
	_, err := c.Count(gjson.Result{})

	require.Error(t, err)
	var discoveryErr discoveryPackage.Error
	require.ErrorAs(t, err, &discoveryErr)
	assert.Equal(t, http.StatusInternalServerError, discoveryErr.Status)
}
