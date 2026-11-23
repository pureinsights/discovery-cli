package tokens

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_setField_scopes tests that setField() sends the comma-separated scopes as a JSON array.
func Test_setField_scopes(t *testing.T) {
	tests := []struct {
		name, input, expected string
	}{
		{name: "single scope", input: "core:*", expected: `{"scopes":["core:*"]}`},
		{name: "multiple scopes with spaces", input: "core:server:read, ingestion:seed:write", expected: `{"scopes":["core:server:read","ingestion:seed:write"]}`},
		{name: "empty values are dropped", input: "core:*,,", expected: `{"scopes":["core:*"]}`},
		{name: "empty input clears the scopes", input: "", expected: `{"scopes":[]}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config, err := setField(`{}`, "scopes", tc.input)
			require.NoError(t, err)
			assert.JSONEq(t, tc.expected, config)
		})
	}
}

// Test_setField_editScopes tests that setField() adds and removes scopes from the current ones of the token.
func Test_setField_editScopes(t *testing.T) {
	const current = `{"name":"t1","scopes":["core:server:read","queryflow:*"]}`
	tests := []struct {
		name, field, input, expected string
	}{
		{name: "add new scopes", field: "addScopes", input: "staging:bucket:create, staging:content:*", expected: `{"name":"t1","scopes":["core:server:read","queryflow:*","staging:bucket:create","staging:content:*"]}`},
		{name: "add an existing scope keeps it once", field: "addScopes", input: "queryflow:*", expected: `{"name":"t1","scopes":["core:server:read","queryflow:*"]}`},
		{name: "remove a scope", field: "removeScopes", input: "core:server:read", expected: `{"name":"t1","scopes":["queryflow:*"]}`},
		{name: "remove a missing scope does nothing", field: "removeScopes", input: "core:*", expected: current},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config, err := setField(current, tc.field, tc.input)
			require.NoError(t, err)
			assert.JSONEq(t, tc.expected, config)
		})
	}
}
