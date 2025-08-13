package discovery

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestError_ErrorString_StringBody(t *testing.T) {
	body := gjson.Parse(`"connection refused"`) // gjson.Result.String() => connection refused
	e := Error{
		Status: 500,
		Body:   body,
	}

	real := e.Error()
	expected := fmt.Sprintf("Status: %d, Body: %s", 500, body.String())

	require.NotEmpty(t, real)
	assert.Equal(t, expected, real)
}
