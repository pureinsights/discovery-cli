package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// PrintJsonObject prints the given JSON object to the Out IOStream.
// If the pretty boolean is true, it prints the object with spacing and indentation
// If not, it prints it in a compact format.
func Test_printJsonObject(t *testing.T) {
	tests := []struct {
		name          string
		pretty        bool
		json          gjson.Result
		expectedPrint string
		err           error
		outWriter     io.Writer
	}{
		{
			name:   "True pretty with a working JSON",
			pretty: true,
			json:   gjson.Parse(`{"name":"test-secret","active":true,"content":{"mechanism":"SCRAM-SHA-1","username":"user","password":"password"}}`),
			expectedPrint: `{
  "active": true,
  "content": {
    "mechanism": "SCRAM-SHA-1",
    "password": "password",
    "username": "user"
  },
  "name": "test-secret"
}` + "\n",
			err: nil,
		},
		{
			name:   "False pretty with a working JSON",
			pretty: false,
			json: gjson.Parse(`{
		"name": "test-secret",
		"active": true,
		"content": {
			"mechanism": "SCRAM-SHA-1", 
			"username": "user",
			"password": "password"
		}
	}`),
			expectedPrint: `{"active":true,"content":{"mechanism":"SCRAM-SHA-1","password":"password","username":"user"},"name":"test-secret"}` + "\n",
			err:           nil,
		},
		{
			name:          "Working JSON, but failing ios.Out print",
			pretty:        true,
			json:          gjson.Parse(`{"name":"test-secret","active":true,"content":{"mechanism":"SCRAM-SHA-1","username":"user","password":"password"}}`),
			expectedPrint: ``,
			err:           errors.New("write failed"),
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
		},
		{
			name:   "Failing JSON, unmarshal fails",
			pretty: true,
			json: gjson.Parse(`{
				"name": "test-secret",
				"active": true,
				"content": {
					"mechanism": "SCRAM-SHA-1", 
					"username": "user",
				
			}`),
			expectedPrint: ``,
			err:           errors.New("invalid character '}' looking for beginning of object key string"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			var out io.Writer
			if tc.outWriter != nil {
				out = tc.outWriter
			} else {
				out = buf
			}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: out,
				Err: os.Stderr,
			}

			err := printJsonObject(ios, tc.pretty, tc.json)

			if tc.err != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedPrint, buf.String())
			}
		})
	}
}

// PrintJsonArray prints the given JSON array to the Out IOStream.
// If the pretty boolean is true, it prints the objects in the array with spacing and indentation and adds brackets at each end.
// If not, it prints the objects in a compact format.
// func Test_printArrayObject(t *testing.T) {
// 	if pretty {
// 		_, err := fmt.Fprint(ios.Out, "[\n")
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	for _, object := range array {
// 		err := printJsonObject(ios, pretty, object)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	if pretty {
// 		_, err := fmt.Fprint(ios.Out, "]\n")
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// // JsonObjectPrinter receives the pretty boolean and returns the function that prints the JSON object with that boolean as a parameter.
// func TestJsonObjectPrinter(t *testing.T) {
// 	return func(ios iostreams.IOStreams, objects ...gjson.Result) error {
// 		if len(objects) != 1 {
// 			return NewError(ErrorExitCode, "JsonObjectPrinter only works with a single JSON object")
// 		}
// 		err := printJsonObject(ios, pretty, objects[0])
// 		if err != nil {
// 			return NewErrorWithCause(ErrorExitCode, err, "Could not print JSON object")
// 		}
// 		return nil
// 	}
// }

// // JsonArrayPrinter receives the pretty boolean and returns the function that prints the JSON array with that boolean as a parameter.
// func TestJsonArrayPrinter(t *testing.T) {
// 	return func(ios iostreams.IOStreams, objects ...gjson.Result) error {
// 		err := printArrayObject(ios, pretty, objects...)
// 		if err != nil {
// 			return NewErrorWithCause(ErrorExitCode, err, "Could not print JSON Array")
// 		}
// 		return nil
// 	}
// }

// // GetObjectPrinter chooses the most appropiate printer depending on the given printer name.
// func TestGetObjectPrinter(t *testing.T) {
// 	switch name {
// 	case "json":
// 		return JsonObjectPrinter(false)
// 	default:
// 		return nil
// 	}
// }

// // GetArrayPrinter chooses the most appropiate printer depending on the given printer name.
// func TestGetArrayPrinter(t *testing.T) {
// 	switch name {
// 	case "json":
// 		return JsonArrayPrinter(false)
// 	default:
// 		return nil
// 	}
// }
