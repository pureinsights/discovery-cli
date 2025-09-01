package discovery

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

// TestError_ErrorString_StringBody tests how Error() behaves with multiple types of string bodies.
func TestError_ErrorString_StringBody(t *testing.T) {
	tests := []struct {
		name   string
		status int
		json   string
	}{
		{
			name:   "JSON object body",
			status: 418,
			json:   `{"error":"request failed","status":500}`,
		},
		{
			name:   "JSON array body",
			status: 400,
			json:   `["error1","error2","error3"]`,
		},
		{
			name:   "JSON string body",
			status: 400,
			json:   `"connection refused"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := gjson.Parse(tc.json)
			e := Error{Status: tc.status, Body: body}

			assert.EqualError(t, e, fmt.Sprintf("status: %d, body: %s", tc.status, body.String()))
		})
	}
}
