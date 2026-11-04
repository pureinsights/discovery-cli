package serviceaccounts

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/pureinsights/discovery-cli/v2/internal/iostreams"
	"github.com/pureinsights/discovery-cli/v2/internal/testutils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestParseLabels tests the parseLabels() function.
func TestParseLabels(t *testing.T) {
	tests := []struct {
		name     string
		labels   string
		expected []map[string]string
	}{
		{name: "Empty string returns no labels", labels: "", expected: []map[string]string{}},
		{name: "Key and value", labels: "keyA:valueA", expected: []map[string]string{{"key": "keyA", "value": "valueA"}}},
		{name: "Value can contain colons", labels: "url:http://host:8080", expected: []map[string]string{{"key": "url", "value": "http://host:8080"}}},
		{
			name:   "Multiple labels ignore spaces and empty entries",
			labels: " keyA:valueA , keyB:valueB,,",
			expected: []map[string]string{
				{"key": "keyA", "value": "valueA"},
				{"key": "keyB", "value": "valueB"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, parseLabels(tc.labels))
		})
	}
}

// TestFormatLabels tests the formatLabels() function.
func TestFormatLabels(t *testing.T) {
	tests := []struct {
		name     string
		labels   string
		expected string
	}{
		{name: "No labels", labels: `[]`, expected: ""},
		{name: "Missing labels", labels: ``, expected: ""},
		{name: "Key and value", labels: `[{"key":"keyA","value":"valueA"}]`, expected: "keyA:valueA"},
		{name: "Key without value", labels: `[{"key":"keyA"}]`, expected: "keyA"},
		{name: "Multiple labels", labels: `[{"key":"keyA","value":"valueA"},{"key":"keyB"}]`, expected: "keyA:valueA,keyB"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, formatLabels(gjson.Parse(tc.labels)))
		})
	}
}

// TestSetField tests the setField() function.
func TestSetField(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    string
		expected string
		err      string
	}{
		{name: "String field", field: "role", value: "ADMIN", expected: `{"name":"ci","role":"ADMIN"}`},
		{name: "Empty string field", field: "description", value: "", expected: `{"name":"ci","description":""}`},
		{name: "Labels are parsed", field: "labels", value: "keyA:valueA", expected: `{"name":"ci","labels":[{"key":"keyA","value":"valueA"}]}`},
		{name: "Empty labels clear the labels", field: "labels", value: "", expected: `{"name":"ci","labels":[]}`},
		{name: "Active is a boolean", field: "active", value: "false", expected: `{"name":"ci","active":false}`},
		{name: "Invalid active fails", field: "active", value: "maybe", err: `Invalid value for active: "maybe"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config, err := setField(`{"name":"ci"}`, tc.field, tc.value)
			if tc.err != "" {
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				assert.ErrorContains(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			assert.JSONEq(t, tc.expected, config)
		})
	}
}

// TestApplyFlags tests the anyFlagChanged() and applyFlags() functions.
func TestApplyFlags(t *testing.T) {
	base := `{"name":"ci","description":"CI","role":"OPERATE","labels":[{"key":"keyA","value":"valueA"}],"active":true}`

	tests := []struct {
		name     string
		args     []string
		changed  bool
		expected string
	}{
		{name: "No flags keep the configuration", args: []string{}, changed: false, expected: base},
		{name: "Only the changed flag is applied", args: []string{"--role", "ADMIN"}, changed: true, expected: `{"name":"ci","description":"CI","role":"ADMIN","labels":[{"key":"keyA","value":"valueA"}],"active":true}`},
		{name: "Enabled false sets active", args: []string{"--enabled=false"}, changed: true, expected: `{"name":"ci","description":"CI","role":"OPERATE","labels":[{"key":"keyA","value":"valueA"}],"active":false}`},
		{name: "Enabled without value sets active to true", args: []string{"--enabled"}, changed: true, expected: base},
		{name: "Empty labels clear the labels", args: []string{"--labels", ""}, changed: true, expected: `{"name":"ci","description":"CI","role":"OPERATE","labels":[],"active":true}`},
		{
			name:     "Every flag is applied",
			args:     []string{"--description", "New", "--role", "ADMIN", "--labels", "keyB:valueB", "--enabled=false"},
			changed:  true,
			expected: `{"name":"ci","description":"New","role":"ADMIN","labels":[{"key":"keyB","value":"valueB"}],"active":false}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var flags accountFlags
			cmd := &cobra.Command{}
			addAccountFlags(cmd, &flags)
			require.NoError(t, cmd.ParseFlags(tc.args))

			assert.Equal(t, tc.changed, anyFlagChanged(cmd))

			config, err := applyFlags(cmd, base, flags)
			require.NoError(t, err)
			assert.JSONEq(t, tc.expected, config)
		})
	}
}

// TestAskAccount tests the askAccount() function.
func TestAskAccount(t *testing.T) {
	base := `{"name":"ci","description":"CI","role":"OPERATE","labels":[{"key":"keyA","value":"valueA"}],"active":true}`

	tests := []struct {
		name     string
		stdin    string
		expected string
		prompts  string
	}{
		{
			name:     "Empty answers keep the current values",
			stdin:    "\n\n\n\n\n",
			expected: base,
			prompts:  "Name [ci]: Description [CI]: Role [OPERATE]: Labels [keyA:valueA]: Active [true]: ",
		},
		{
			name:     "A space clears the value",
			stdin:    "\n \n\n \n\n",
			expected: `{"name":"ci","description":"","role":"OPERATE","labels":[],"active":true}`,
		},
		{
			name:     "New answers replace the values",
			stdin:    "ci2\nNew\nADMIN\nkeyB:valueB,keyC:valueC\nfalse\n",
			expected: `{"name":"ci2","description":"New","role":"ADMIN","labels":[{"key":"keyB","value":"valueB"},{"key":"keyC","value":"valueC"}],"active":false}`,
		},
		{
			name:     "Input without a final newline is read",
			stdin:    "\n\nADMIN",
			expected: `{"name":"ci","description":"CI","role":"ADMIN","labels":[{"key":"keyA","value":"valueA"}],"active":true}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			ios := &iostreams.IOStreams{In: strings.NewReader(tc.stdin), Out: out, Err: &bytes.Buffer{}}

			config, err := askAccount(ios, base)
			require.NoError(t, err)
			assert.JSONEq(t, tc.expected, config)
			if tc.prompts != "" {
				assert.Equal(t, tc.prompts, out.String())
			}
		})
	}
}

// TestAskAccount_Errors tests the askAccount() function when it fails.
func TestAskAccount_Errors(t *testing.T) {
	readErr := errors.New("read failed")
	writeErr := errors.New("write failed")

	tests := []struct {
		name     string
		ios      *iostreams.IOStreams
		err      error
		contains string
	}{
		{
			name: "Reading the answer fails",
			ios:  &iostreams.IOStreams{In: testutils.ErrReader{Err: readErr}, Out: &bytes.Buffer{}, Err: &bytes.Buffer{}},
			err:  cli.NewErrorWithCause(cli.ErrorExitCode, readErr, "Could not read the %s", "name"),
		},
		{
			name: "Writing the prompt fails",
			ios:  &iostreams.IOStreams{In: strings.NewReader("\n"), Out: testutils.ErrWriter{Err: writeErr}, Err: &bytes.Buffer{}},
			err:  cli.NewErrorWithCause(cli.ErrorExitCode, writeErr, "Could not read the %s", "name"),
		},
		{
			name:     "Invalid active answer",
			ios:      &iostreams.IOStreams{In: strings.NewReader("\n\n\n\nmaybe\n"), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}},
			contains: `Invalid value for active: "maybe"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := askAccount(tc.ios, `{"name":"ci","active":true}`)
			var errStruct cli.Error
			require.ErrorAs(t, err, &errStruct)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.ErrorContains(t, err, tc.contains)
			}
		})
	}
}
