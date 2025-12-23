package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Test_printJsonObject tests the printJsonObject() function.
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
}`,
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
			expectedPrint: `{"active":true,"content":{"mechanism":"SCRAM-SHA-1","password":"password","username":"user"},"name":"test-secret"}`,
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

// Test_printArrayObject tests the printArrayObject() function.
func Test_printArrayObject(t *testing.T) {
	tests := []struct {
		name          string
		pretty        bool
		array         []gjson.Result
		expectedPrint string
		err           error
		outWriter     io.Writer
	}{
		{
			name:   "True pretty with a working JSON array",
			pretty: true,
			array: gjson.Parse(`[{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","labels":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"MongoDB text processor 4","type":"mongo"},     
			{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"5f125024-1e5e-4591-9fee-365dc20eeeed","labels":[],"lastUpdatedTimestamp":"2025-08-18T20:55:43Z","name":"MongoDB text processor","type":"mongo"},       
			{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"86e7f920-a4e4-4b64-be84-5437a7673db8","labels":[],"lastUpdatedTimestamp":"2025-08-14T18:02:38Z","name":"Script processor","type":"script"}
			]`).Array(),
			expectedPrint: `[
{
  "active": true,
  "creationTimestamp": "2025-08-21T17:57:16Z",
  "id": "3393f6d9-94c1-4b70-ba02-5f582727d998",
  "labels": [],
  "lastUpdatedTimestamp": "2025-08-21T17:57:16Z",
  "name": "MongoDB text processor 4",
  "type": "mongo"
},
{
  "active": true,
  "creationTimestamp": "2025-08-14T18:02:38Z",
  "id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
  "labels": [],
  "lastUpdatedTimestamp": "2025-08-18T20:55:43Z",
  "name": "MongoDB text processor",
  "type": "mongo"
},
{
  "active": true,
  "creationTimestamp": "2025-08-14T18:02:38Z",
  "id": "86e7f920-a4e4-4b64-be84-5437a7673db8",
  "labels": [],
  "lastUpdatedTimestamp": "2025-08-14T18:02:38Z",
  "name": "Script processor",
  "type": "script"
}
]` + "\n",
			err: nil,
		},
		{
			name:   "True pretty with a string array",
			pretty: true,
			array:  gjson.Parse(`["test1", "test2", "test3"]`).Array(),
			expectedPrint: `[
"test1",
"test2",
"test3"
]` + "\n",
			err: nil,
		},
		{
			name:   "False pretty with a working JSON array",
			pretty: false,
			array: gjson.Parse(`[
				{
				"type": "mongo",
				"name": "MongoDB text processor 4",
				"labels": [],
				"active": true,
				"id": "3393f6d9-94c1-4b70-ba02-5f582727d998",
				"creationTimestamp": "2025-08-21T17:57:16Z",
				"lastUpdatedTimestamp": "2025-08-21T17:57:16Z"
				},
				{
				"type": "mongo",
				"name": "MongoDB text processor",
				"labels": [],
				"active": true,
				"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
				},
				{
				"type": "script",
				"name": "Script processor",
				"labels": [],
				"active": true,
				"id": "86e7f920-a4e4-4b64-be84-5437a7673db8",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-14T18:02:38Z"
				}
			]`).Array(),
			expectedPrint: `{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","labels":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"MongoDB text processor 4","type":"mongo"}
{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"5f125024-1e5e-4591-9fee-365dc20eeeed","labels":[],"lastUpdatedTimestamp":"2025-08-18T20:55:43Z","name":"MongoDB text processor","type":"mongo"}
{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"86e7f920-a4e4-4b64-be84-5437a7673db8","labels":[],"lastUpdatedTimestamp":"2025-08-14T18:02:38Z","name":"Script processor","type":"script"}
`,
			err: nil,
		},
		{
			name:   "Working JSON Array, but failing ios.Out print",
			pretty: true,
			array: gjson.Parse(`[{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","labels":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"MongoDB text processor 4","type":"mongo"},     
			{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"5f125024-1e5e-4591-9fee-365dc20eeeed","labels":[],"lastUpdatedTimestamp":"2025-08-18T20:55:43Z","name":"MongoDB text processor","type":"mongo"},       
			{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"86e7f920-a4e4-4b64-be84-5437a7673db8","labels":[],"lastUpdatedTimestamp":"2025-08-14T18:02:38Z","name":"Script processor","type":"script"}
			]`).Array(),
			expectedPrint: ``,
			err:           errors.New("write failed"),
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
		},
		{
			name:   "Working JSON Array, but fail to print \",\" to ios.Out",
			pretty: true,
			array: gjson.Parse(`[{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","labels":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"MongoDB text processor 4","type":"mongo"},     
			{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"5f125024-1e5e-4591-9fee-365dc20eeeed","labels":[],"lastUpdatedTimestamp":"2025-08-18T20:55:43Z","name":"MongoDB text processor","type":"mongo"},       
			{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"86e7f920-a4e4-4b64-be84-5437a7673db8","labels":[],"lastUpdatedTimestamp":"2025-08-14T18:02:38Z","name":"Script processor","type":"script"}
			]`).Array(),
			expectedPrint: ``,
			err:           errors.New("write failed"),
			outWriter:     &testutils.FailOnNWriter{Writer: &bytes.Buffer{}, N: 3},
		},
		{
			name:   "Working JSON Array, but fail to print \"\n\" to ios.Out",
			pretty: false,
			array: gjson.Parse(`[{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","labels":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"MongoDB text processor 4","type":"mongo"},     
			{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"5f125024-1e5e-4591-9fee-365dc20eeeed","labels":[],"lastUpdatedTimestamp":"2025-08-18T20:55:43Z","name":"MongoDB text processor","type":"mongo"},       
			{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"86e7f920-a4e4-4b64-be84-5437a7673db8","labels":[],"lastUpdatedTimestamp":"2025-08-14T18:02:38Z","name":"Script processor","type":"script"}
			]`).Array(),
			expectedPrint: ``,
			err:           errors.New("write failed"),
			outWriter:     &testutils.FailOnNWriter{Writer: &bytes.Buffer{}, N: 2},
		},
		{
			name:   "Working JSON Array, but fail to print \"]\" to ios.Out",
			pretty: true,
			array: gjson.Parse(`[{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","labels":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"MongoDB text processor 4","type":"mongo"},     
			{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"5f125024-1e5e-4591-9fee-365dc20eeeed","labels":[],"lastUpdatedTimestamp":"2025-08-18T20:55:43Z","name":"MongoDB text processor","type":"mongo"},       
			{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"86e7f920-a4e4-4b64-be84-5437a7673db8","labels":[],"lastUpdatedTimestamp":"2025-08-14T18:02:38Z","name":"Script processor","type":"script"}
			]`).Array(),
			expectedPrint: ``,
			err:           errors.New("write failed"),
			outWriter:     &testutils.FailOnNWriter{Writer: &bytes.Buffer{}, N: 10},
		},
		{
			name:   "Failing JSON, unmarshal fails",
			pretty: true,
			array: gjson.Parse(`[{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","labels":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"MongoDB text processor 4","type":"mongo"},     
			{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"5f125024-1e5e-4591-9fee-365dc20eeeed","labels":[],"lastUpdatedTimestamp":"2025-08-18T20:55:43Z","name":"MongoDB text processor","type":"mongo",       
			{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"86e7f920-a4e4-4b64-be84-5437a7673db8","labels":[],"lastUpdatedTimestamp":"2025-08-14T18:02:38Z","name":"Script processor","type":"script"}
			]`).Array(),
			expectedPrint: ``,
			err:           errors.New("invalid character '{' looking for beginning of object key string"),
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

			err := printArrayObject(ios, tc.pretty, tc.array...)

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

// TestJsonObjectPrinter tests the JsonObjectPrinter() function.
func TestJsonObjectPrinter(t *testing.T) {
	tests := []struct {
		name          string
		array         []gjson.Result
		pretty        bool
		expectedPrint string
		err           error
		outWriter     io.Writer
	}{
		{
			name:   "The array has only one element",
			pretty: false,
			array: gjson.Parse(`[{
		"name": "test-secret",
		"active": true,
		"content": {
			"mechanism": "SCRAM-SHA-1", 
			"username": "user",
			"password": "password"
		}
	}]`).Array(),
			expectedPrint: `{"active":true,"content":{"mechanism":"SCRAM-SHA-1","password":"password","username":"user"},"name":"test-secret"}` + "\n",
			err:           nil,
		},
		{
			name:   "The array does not have only one element",
			pretty: false,
			array: gjson.Parse(`[{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","labels":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"MongoDB text processor 4","type":"mongo"},     
			{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"5f125024-1e5e-4591-9fee-365dc20eeeed","labels":[],"lastUpdatedTimestamp":"2025-08-18T20:55:43Z","name":"MongoDB text processor","type":"mongo"},       
			{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"86e7f920-a4e4-4b64-be84-5437a7673db8","labels":[],"lastUpdatedTimestamp":"2025-08-14T18:02:38Z","name":"Script processor","type":"script"}
			]`).Array(),
			expectedPrint: "",
			err:           NewError(ErrorExitCode, "JsonObjectPrinter only works with a single JSON object"),
		},
		{
			name:          "Working JSON Array, but failing ios.Out print",
			pretty:        false,
			array:         gjson.Parse(`[{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","labels":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"MongoDB text processor 4","type":"mongo"}]`).Array(),
			expectedPrint: ``,
			err:           NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
		},
		{
			name:          "Working JSON Array, but failing to print \"\n\"t",
			pretty:        false,
			array:         gjson.Parse(`[{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","labels":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"MongoDB text processor 4","type":"mongo"}]`).Array(),
			expectedPrint: ``,
			err:           NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
			outWriter:     &testutils.FailOnNWriter{Writer: &bytes.Buffer{}, N: 2},
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

			jsonPrinter := JsonObjectPrinter(tc.pretty)
			err := jsonPrinter(ios, tc.array...)

			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedPrint, buf.String())
			}
		})
	}
}

// TestJsonArrayPrinter tests the JsonArrayPrinter() function.
func TestJsonArrayPrinter(t *testing.T) {
	tests := []struct {
		name          string
		array         []gjson.Result
		pretty        bool
		expectedPrint string
		err           error
		outWriter     io.Writer
	}{
		{
			name:   "The array could be printed correctly",
			pretty: true,
			array: gjson.Parse(`[{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","labels":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"MongoDB text processor 4","type":"mongo"},     
			{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"5f125024-1e5e-4591-9fee-365dc20eeeed","labels":[],"lastUpdatedTimestamp":"2025-08-18T20:55:43Z","name":"MongoDB text processor","type":"mongo"},       
			{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"86e7f920-a4e4-4b64-be84-5437a7673db8","labels":[],"lastUpdatedTimestamp":"2025-08-14T18:02:38Z","name":"Script processor","type":"script"}
			]`).Array(),
			expectedPrint: `[
{
  "active": true,
  "creationTimestamp": "2025-08-21T17:57:16Z",
  "id": "3393f6d9-94c1-4b70-ba02-5f582727d998",
  "labels": [],
  "lastUpdatedTimestamp": "2025-08-21T17:57:16Z",
  "name": "MongoDB text processor 4",
  "type": "mongo"
},
{
  "active": true,
  "creationTimestamp": "2025-08-14T18:02:38Z",
  "id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
  "labels": [],
  "lastUpdatedTimestamp": "2025-08-18T20:55:43Z",
  "name": "MongoDB text processor",
  "type": "mongo"
},
{
  "active": true,
  "creationTimestamp": "2025-08-14T18:02:38Z",
  "id": "86e7f920-a4e4-4b64-be84-5437a7673db8",
  "labels": [],
  "lastUpdatedTimestamp": "2025-08-14T18:02:38Z",
  "name": "Script processor",
  "type": "script"
}
]` + "\n",
			err: nil,
		},
		{
			name:          "Working JSON Array, but failing ios.Out print",
			pretty:        false,
			array:         gjson.Parse(`[{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","labels":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"MongoDB text processor 4","type":"mongo"}]`).Array(),
			expectedPrint: ``,
			err:           NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON Array"),
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
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

			jsonPrinter := JsonArrayPrinter(tc.pretty)
			err := jsonPrinter(ios, tc.array...)

			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedPrint, buf.String())
			}
		})
	}
}

// TestGetObjectPrinter tests the GetObjectPrinter() function.
func TestGetObjectPrinter(t *testing.T) {
	tests := []struct {
		name            string
		printerName     string
		expectedPrinter Printer
		input           []gjson.Result
		expectedOutput  string
	}{
		{
			name:            "The switch returns the JSON Printer",
			printerName:     "json",
			expectedPrinter: JsonObjectPrinter(false),
			input: gjson.Parse(`[{
		"name": "test-secret",
		"active": true,
		"content": {
			"mechanism": "SCRAM-SHA-1", 
			"username": "user",
			"password": "password"
		}
	}]`).Array(),
			expectedOutput: `{"active":true,"content":{"mechanism":"SCRAM-SHA-1","password":"password","username":"user"},"name":"test-secret"}` + "\n",
		},
		{
			name:            "The switch returns the JSON Pretty Printer",
			printerName:     "pretty-json",
			expectedPrinter: JsonObjectPrinter(true),
			input: gjson.Parse(`[{
		"name": "test-secret",
		"active": true,
		"content": {
			"mechanism": "SCRAM-SHA-1", 
			"username": "user",
			"password": "password"
		}
	}]`).Array(),
			expectedOutput: "{\n  \"active\": true,\n  \"content\": {\n    \"mechanism\": \"SCRAM-SHA-1\",\n    \"password\": \"password\",\n    \"username\": \"user\"\n  },\n  \"name\": \"test-secret\"\n}\n",
		},
		{
			name:            "The switch returns the default case",
			printerName:     "doesnotexist",
			expectedPrinter: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			printer := GetObjectPrinter(tc.printerName)

			if tc.expectedPrinter != nil {
				buf := &bytes.Buffer{}

				ios := iostreams.IOStreams{
					In:  os.Stdin,
					Out: buf,
					Err: os.Stderr,
				}

				err := printer(ios, tc.input...)
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			} else {
				assert.Nil(t, printer)
			}
		})
	}
}

// TestGetArrayPrinter tests the GetArrayPrinter() function.
func TestGetArrayPrinter(t *testing.T) {
	tests := []struct {
		name            string
		printerName     string
		expectedPrinter Printer
		input           []gjson.Result
		expectedOutput  string
	}{
		{
			name:            "The switch returns the JSON Printer",
			printerName:     "json",
			expectedPrinter: JsonArrayPrinter(false),
			input: gjson.Parse(`[
				{
				"type": "mongo",
				"name": "MongoDB text processor 4",
				"labels": [],
				"active": true,
				"id": "3393f6d9-94c1-4b70-ba02-5f582727d998",
				"creationTimestamp": "2025-08-21T17:57:16Z",
				"lastUpdatedTimestamp": "2025-08-21T17:57:16Z"
				},
				{
				"type": "mongo",
				"name": "MongoDB text processor",
				"labels": [],
				"active": true,
				"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
				},
				{
				"type": "script",
				"name": "Script processor",
				"labels": [],
				"active": true,
				"id": "86e7f920-a4e4-4b64-be84-5437a7673db8",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-14T18:02:38Z"
				}
			]`).Array(),
			expectedOutput: `{"active":true,"creationTimestamp":"2025-08-21T17:57:16Z","id":"3393f6d9-94c1-4b70-ba02-5f582727d998","labels":[],"lastUpdatedTimestamp":"2025-08-21T17:57:16Z","name":"MongoDB text processor 4","type":"mongo"}
{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"5f125024-1e5e-4591-9fee-365dc20eeeed","labels":[],"lastUpdatedTimestamp":"2025-08-18T20:55:43Z","name":"MongoDB text processor","type":"mongo"}
{"active":true,"creationTimestamp":"2025-08-14T18:02:38Z","id":"86e7f920-a4e4-4b64-be84-5437a7673db8","labels":[],"lastUpdatedTimestamp":"2025-08-14T18:02:38Z","name":"Script processor","type":"script"}
`,
		},
		{
			name:            "The switch returns the JSON Pretty Printer",
			printerName:     "pretty-json",
			expectedPrinter: JsonArrayPrinter(true),
			input: gjson.Parse(`[
				{
				"type": "mongo",
				"name": "MongoDB text processor 4",
				"labels": [],
				"active": true,
				"id": "3393f6d9-94c1-4b70-ba02-5f582727d998",
				"creationTimestamp": "2025-08-21T17:57:16Z",
				"lastUpdatedTimestamp": "2025-08-21T17:57:16Z"
				},
				{
				"type": "mongo",
				"name": "MongoDB text processor",
				"labels": [],
				"active": true,
				"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
				},
				{
				"type": "script",
				"name": "Script processor",
				"labels": [],
				"active": true,
				"id": "86e7f920-a4e4-4b64-be84-5437a7673db8",
				"creationTimestamp": "2025-08-14T18:02:38Z",
				"lastUpdatedTimestamp": "2025-08-14T18:02:38Z"
				}
			]`).Array(),
			expectedOutput: "[\n{\n  \"active\": true,\n  \"creationTimestamp\": \"2025-08-21T17:57:16Z\",\n  \"id\": \"3393f6d9-94c1-4b70-ba02-5f582727d998\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-21T17:57:16Z\",\n  \"name\": \"MongoDB text processor 4\",\n  \"type\": \"mongo\"\n},\n{\n  \"active\": true,\n  \"creationTimestamp\": \"2025-08-14T18:02:38Z\",\n  \"id\": \"5f125024-1e5e-4591-9fee-365dc20eeeed\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-18T20:55:43Z\",\n  \"name\": \"MongoDB text processor\",\n  \"type\": \"mongo\"\n},\n{\n  \"active\": true,\n  \"creationTimestamp\": \"2025-08-14T18:02:38Z\",\n  \"id\": \"86e7f920-a4e4-4b64-be84-5437a7673db8\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-14T18:02:38Z\",\n  \"name\": \"Script processor\",\n  \"type\": \"script\"\n}\n]\n",
		},
		{
			name:            "The switch returns the default case",
			printerName:     "doesnotexist",
			expectedPrinter: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			printer := GetArrayPrinter(tc.printerName)

			if tc.expectedPrinter != nil {
				buf := &bytes.Buffer{}

				ios := iostreams.IOStreams{
					In:  os.Stdin,
					Out: buf,
					Err: os.Stderr,
				}

				err := printer(ios, tc.input...)
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			} else {
				assert.Nil(t, printer)
			}
		})
	}
}
