package discovery

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Test_statusChecker_StatusCheck_ProductOnline tests the statusChecker.StatusCheck() function when the Discovery product is up.
func Test_statusChecker_StatusCheck_ProductOnline(t *testing.T) {
	const statusUp = `{
    "status": "UP"
}`
	srv := httptest.NewServer(testutils.HttpHandler(t, http.StatusOK, "application/json", statusUp, func(t *testing.T, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/health", r.URL.Path)
	}))
	defer srv.Close()

	sc := statusChecker{
		client: newClient(srv.URL, ""),
	}

	status, err := sc.StatusCheck()
	require.NoError(t, err)

	assert.Equal(t, gjson.Parse(statusUp), status)
}

// Test_statusChecker_StatusCheck_ProductOffline tests the statusChecker.StatusCheck() function when the Discovery product is down.
func Test_statusChecker_StatusCheck_ProductOffline(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	base := srv.URL
	srv.Close()

	sc := statusChecker{
		client: newClient(srv.URL, ""),
	}

	status, err := sc.StatusCheck()
	require.Error(t, err)
	assert.Equal(t, gjson.Result{}, status)
	assert.Contains(t, err.Error(), base+"/health")
}
