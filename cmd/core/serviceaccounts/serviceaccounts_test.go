package serviceaccounts

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/pureinsights/discovery-cli/v2/internal/iostreams"
	"github.com/pureinsights/discovery-cli/v2/internal/testutils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const accountId = "3d51beef-8b90-40aa-84b5-033241dc6239"

const account = `{
	"id": "3d51beef-8b90-40aa-84b5-033241dc6239",
	"name": "ci",
	"description": "Continuous Integration account",
	"role": "OPERATE",
	"labels": [{"key": "keyA", "value": "valueA"}, {"key": "keyB", "value": "valueB"}],
	"active": true,
	"creationTimestamp": "2025-10-17T22:37:57Z"
}`

// searchResponses returns the responses needed to find the "ci" service account by name.
func searchResponses() map[string]testutils.MockResponse {
	return map[string]testutils.MockResponse{
		"POST:/v2/account/service/search": {
			StatusCode:  http.StatusOK,
			ContentType: "application/json",
			Body:        `{"content":[{"source":` + account + `}],"totalSize":1,"totalPages":1,"empty":false,"numberOfElements":1,"pageNumber":0}`,
		},
		"GET:/v2/account/service/" + accountId: {
			StatusCode:  http.StatusOK,
			ContentType: "application/json",
			Body:        account,
		},
	}
}

// expectBody adds a response to the given path that asserts the request body is the expected JSON.
func expectBody(t *testing.T, responses map[string]testutils.MockResponse, path, expected string) {
	responses[path] = testutils.MockResponse{
		StatusCode:  http.StatusOK,
		ContentType: "application/json",
		Body:        expected,
		Assertions: func(t *testing.T, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			assert.JSONEq(t, expected, string(body))
		},
	}
}

// runCommand executes the given command against a mock Core server, using stdin as the user input.
func runCommand(t *testing.T, newCommand func(cli.Discovery) *cobra.Command, responses map[string]testutils.MockResponse, stdin string, args ...string) (string, error) {
	srv := httptest.NewServer(testutils.HttpMultiResponseHandler(t, responses))
	defer srv.Close()

	out := &bytes.Buffer{}
	ios := iostreams.IOStreams{In: strings.NewReader(stdin), Out: out, Err: &bytes.Buffer{}}

	vpr := viper.New()
	vpr.Set("profile", "default")
	vpr.Set("output", "json")
	vpr.Set("default.core_url", srv.URL)
	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	cmd := newCommand(d)
	cmd.SilenceUsage = true
	cmd.SetOut(out)
	cmd.SetErr(ios.Err)
	cmd.PersistentFlags().StringP("profile", "p", "default", "configuration profile to use")
	cmd.SetArgs(args)

	err := cmd.Execute()
	return out.String(), err
}

// TestLabels tests the parseLabels and formatLabels functions.
func TestLabels(t *testing.T) {
	parsed := parseLabels("keyA:valueA, keyB ,url:http://x,")
	assert.Equal(t, []map[string]string{
		{"key": "keyA", "value": "valueA"},
		{"key": "keyB"},
		{"key": "url", "value": "http://x"},
	}, parsed)
	assert.Equal(t, "keyA:valueA,keyB:valueB", formatLabels(gjson.Get(account, "labels")))
	assert.Equal(t, []map[string]string{}, parseLabels(""))
}

// TestNewCreateCommand tests the create command with flags and interactively.
func TestNewCreateCommand(t *testing.T) {
	expected := `{"name":"ci","description":"Continuous Integration account","role":"OPERATE","labels":[{"key":"keyA","value":"valueA"},{"key":"keyB","value":"valueB"}],"active":true}`

	t.Run("With flags", func(t *testing.T) {
		responses := map[string]testutils.MockResponse{}
		expectBody(t, responses, "POST:/v2/account/service", expected)
		out, err := runCommand(t, NewCreateCommand, responses, "", "ci", "--description", "Continuous Integration account", "--role", "OPERATE", "--labels", "keyA:valueA,keyB:valueB", "--enabled")
		require.NoError(t, err)
		assert.Contains(t, out, `"name":"ci"`)
	})

	t.Run("Interactive", func(t *testing.T) {
		responses := map[string]testutils.MockResponse{}
		expectBody(t, responses, "POST:/v2/account/service", expected)
		out, err := runCommand(t, NewCreateCommand, responses, "ci\nContinuous Integration account\nOPERATE\nkeyA:valueA,keyB:valueB\n\n")
		require.NoError(t, err)
		assert.Contains(t, out, "Name []: Description []: Role []: Labels []: Active [true]: ")
	})

	t.Run("Flags without a name fail", func(t *testing.T) {
		_, err := runCommand(t, NewCreateCommand, map[string]testutils.MockResponse{}, "", "--role", "OPERATE")
		assert.ErrorContains(t, err, "The service account name is required")
	})

	t.Run("Invalid active value fails", func(t *testing.T) {
		_, err := runCommand(t, NewCreateCommand, map[string]testutils.MockResponse{}, "ci\n\n\n\nmaybe\n")
		var cliErr cli.Error
		require.ErrorAs(t, err, &cliErr)
		assert.Contains(t, err.Error(), `Invalid value for active: "maybe"`)
	})
}

// TestNewUpdateCommand tests the update command with flags and interactively.
func TestNewUpdateCommand(t *testing.T) {
	t.Run("With flags only changes the given fields", func(t *testing.T) {
		responses := searchResponses()
		expected := strings.Replace(account, `"OPERATE"`, `"ADMIN"`, 1)
		expectBody(t, responses, "PUT:/v2/account/service/"+accountId, expected)
		_, err := runCommand(t, NewUpdateCommand, responses, "", "ci", "--role", "ADMIN")
		require.NoError(t, err)
	})

	t.Run("Interactive shows the current values", func(t *testing.T) {
		responses := searchResponses()
		expected := strings.Replace(strings.Replace(account, `"OPERATE"`, `"ADMIN"`, 1), `"description": "Continuous Integration account"`, `"description": ""`, 1)
		expectBody(t, responses, "PUT:/v2/account/service/"+accountId, expected)
		out, err := runCommand(t, NewUpdateCommand, responses, "\n \nADMIN\n\n\n", "ci")
		require.NoError(t, err)
		assert.Contains(t, out, "Name [ci]: Description [Continuous Integration account]: Role [OPERATE]: Labels [keyA:valueA,keyB:valueB]: Active [true]: ")
	})

	t.Run("Account not found fails", func(t *testing.T) {
		responses := map[string]testutils.MockResponse{
			"POST:/v2/account/service/search": {StatusCode: http.StatusOK, ContentType: "application/json", Body: `{"content":[],"totalSize":0,"totalPages":0,"empty":true}`},
		}
		_, err := runCommand(t, NewUpdateCommand, responses, "", "missing", "--role", "ADMIN")
		assert.ErrorContains(t, err, "Could not get service account ID.")
	})
}
