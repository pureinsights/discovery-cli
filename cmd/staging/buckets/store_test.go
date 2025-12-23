package buckets

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestNewStoreCommand tests the NewStoreCommand() function.
func TestNewStoreCommand(t *testing.T) {
	tests := []struct {
		name      string
		url       bool
		apiKey    string
		outGolden string
		errGolden string
		outBytes  []byte
		errBytes  []byte
		data      string
		file      string
		responses map[string]testutils.MockResponse
		err       error
	}{
		// Working case
		{
			name:      "Store receives the bucket config through data flag",
			url:       true,
			apiKey:    "",
			outGolden: "NewStoreCommand_Out_StoreDataFlag",
			errGolden: "NewStoreCommand_Err_StoreDataFlag",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_StoreDataFlag"),
			errBytes:  []byte(nil),
			data: `{
	"indices": [
		{
			"name": "myIndexA",
			"fields": [
				{
				"fieldName": "ASC"
				}
			],
			"unique": false
			},
			{
			"name": "myIndexB",
			"fields": [
				{
				"fieldName2": "DESC"
				}
			],
			"unique": false
		}
	],
	"config": {}
}`,
			file: "",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/bucket/my-bucket": {
					StatusCode: http.StatusOK,
					Body: `{
  "acknowledged": true
}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/bucket/my-bucket", r.URL.Path)
					},
				},
				"GET:/v2/bucket/my-bucket": {
					StatusCode: http.StatusOK,
					Body: `{
  "name": "my-bucket",
  "documentCount": {},
  "indices": [
    {
		"name": "myIndexA",
		"fields": [
			{
			"fieldName": "ASC"
			}
		],
		"unique": false
	},
	{
		"name": "myIndexB",
		"fields": [
			{
			"fieldName2": "DESC"
			}
		],
		"unique": false
	}
  ]
}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/bucket/my-bucket", r.URL.Path)
					},
				},
			},
			err: nil,
		},
		{
			name:      "Store receives the bucket config through the argument",
			url:       true,
			apiKey:    "",
			outGolden: "NewStoreCommand_Out_StoreFileArg",
			errGolden: "NewStoreCommand_Err_StoreFileArg",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_StoreFileArg"),
			errBytes:  []byte(nil),
			data:      ``,
			file:      "testdata/StoreCommand_BucketConfig.json",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/bucket/my-bucket": {
					StatusCode: http.StatusOK,
					Body: `{
  "acknowledged": true
}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/bucket/my-bucket", r.URL.Path)
					},
				},
				"GET:/v2/bucket/my-bucket": {
					StatusCode: http.StatusOK,
					Body: `{
  "name": "my-bucket",
  "documentCount": {},
  "indices": [
    {
		"name": "myIndexA",
		"fields": [
			{
			"fieldName": "ASC"
			}
		],
		"unique": false
	},
	{
		"name": "myIndexB",
		"fields": [
			{
			"fieldName2": "DESC"
			}
		],
		"unique": false
	}
  ]
}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/bucket/my-bucket", r.URL.Path)
					},
				},
			},
			err: nil,
		},

		// Error case
		{
			name:      "No URL",
			outGolden: "NewStoreCommand_Out_NoURL",
			errGolden: "NewStoreCommand_Err_NoURL",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_NoURL"),
			errBytes:  testutils.Read(t, "NewStoreCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			data: `{
	"indices": [
		{
			"name": "myIndexA",
			"fields": [
				{
				"fieldName": "ASC"
				}
			],
			"unique": false
			},
			{
			"name": "myIndexB",
			"fields": [
				{
				"fieldName2": "DESC"
				}
			],
			"unique": false
		}
	],
	"config": {}
}`,
			file: "",
			err:  cli.NewError(cli.ErrorExitCode, "The Discovery Staging URL is missing for profile \"default\".\nTo set the URL for the Discovery Staging API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery staging config --profile \"default\""),
		},
		{
			name:      "Store receives the data flag and bucket config file",
			url:       true,
			apiKey:    "",
			outGolden: "NewStoreCommand_Out_DataFlagAndFile",
			errGolden: "NewStoreCommand_Err_DataFlagAndFile",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_DataFlagAndFile"),
			errBytes:  testutils.Read(t, "NewStoreCommand_Err_DataFlagAndFile"),
			data: `{
	"indices": [
		{
			"name": "myIndexA",
			"fields": [
				{
				"fieldName": "ASC"
				}
			],
			"unique": false
			},
			{
			"name": "myIndexB",
			"fields": [
				{
				"fieldName2": "DESC"
				}
			],
			"unique": false
		}
	],
	"config": {}
}`,
			file:      "testdata/StoreCommand_BucketConfig.json",
			responses: map[string]testutils.MockResponse{},
			err:       cli.NewError(cli.ErrorExitCode, "The data flag can only have the bucket name argument."),
		},
		{
			name:      "The bucket already exists",
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewStoreCommand_Out_BucketExists",
			errGolden: "NewStoreCommand_Err_BucketExists",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_BucketExists"),
			errBytes:  testutils.Read(t, "NewStoreCommand_Err_BucketExists"),
			data: `{
	"config": {}
}`,
			file: "",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/bucket/my-bucket": {
					StatusCode: http.StatusConflict,
					Body: `{
  "acknowledged": false
}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/bucket/my-bucket", r.URL.Path)
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
  "acknowledged": false
}`)}, "Could not create bucket with name \"my-bucket\"."),
		},
		{
			name:      "StoreCommand tries to read a file that does not exist",
			url:       true,
			apiKey:    "apiKey123",
			outGolden: "NewStoreCommand_Out_FileNotExists",
			errGolden: "NewStoreCommand_Err_FileNotExists",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_FileNotExists"),
			data:      "",
			file:      "doesnotexist",
			errBytes:  testutils.Read(t, "NewStoreCommand_Err_FileNotExists"),
			err:       cli.NewErrorWithCause(cli.ErrorExitCode, fmt.Errorf("file does not exist: %s", "doesnotexist"), "Could not read file \"doesnotexist\""),
		},
		{
			name:      "Printing JSON Object fails",
			outGolden: "NewStoreCommand_Out_PrintJSONFails",
			errGolden: "NewStoreCommand_Err_PrintJSONFails",
			outBytes:  testutils.Read(t, "NewStoreCommand_Out_PrintJSONFails"),
			errBytes:  testutils.Read(t, "NewStoreCommand_Err_PrintJSONFails"),
			url:       true,
			apiKey:    "apiKey123",
			data:      ``,
			file:      "testdata/StoreCommand_BucketConfig.json",
			responses: map[string]testutils.MockResponse{
				"POST:/v2/bucket/my-bucket": {
					StatusCode: http.StatusOK,
					Body: `{
  "acknowledged": true
}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodPost, r.Method)
						assert.Equal(t, "/v2/bucket/my-bucket", r.URL.Path)
					},
				},
				"GET:/v2/bucket/my-bucket": {
					StatusCode: http.StatusOK,
					Body: `{
  "name": "my-bucket",
  "documentCount": {},
  "indices": [
    {
		"name": "myIndexA",
		"fields": [
			{
			"fieldName": "ASC
			}
		],
		"unique": false
	},
	{
		"name": "myIndexB",
		"fields": [
			{
			"fieldName2": "DESC"
			}
		],
		"unique": false
	}
  ]
}`,
					ContentType: "application/json",
					Assertions: func(t *testing.T, r *http.Request) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, "/v2/bucket/my-bucket", r.URL.Path)
					},
				},
			},
			err: cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid character '\\n' in string literal"), "Could not print JSON object"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(testutils.HttpMultiResponseHandler(t, tc.responses))

			defer srv.Close()

			in := strings.NewReader("")
			out := &bytes.Buffer{}

			errBuf := &bytes.Buffer{}
			ios := iostreams.IOStreams{
				In:  in,
				Out: out,
				Err: errBuf,
			}

			vpr := viper.New()
			vpr.Set("profile", "default")
			vpr.Set("output", "pretty-json")
			if tc.url {
				vpr.Set("default.staging_url", srv.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.staging_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, t.TempDir())

			storeCmd := NewStoreCommand(d)

			storeCmd.SetIn(ios.In)
			storeCmd.SetOut(ios.Out)
			storeCmd.SetErr(ios.Err)

			storeCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			args := []string{"my-bucket"}
			if tc.data != "" || tc.file == "" {
				args = append(args, "--data")
				args = append(args, tc.data)
			}

			if tc.file != "" {
				args = append(args, tc.file)
			}

			storeCmd.SetArgs(args)

			err := storeCmd.Execute()
			if tc.err != nil {
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				cliError, _ := tc.err.(cli.Error)
				if !errors.Is(cliError.Cause, fs.ErrNotExist) {
					assert.EqualError(t, err, tc.err.Error())
					testutils.CompareBytes(t, tc.errGolden, tc.errBytes, errBuf.Bytes())
				} else {
					assert.Equal(t, cliError.ExitCode, errStruct.ExitCode)
					assert.Equal(t, cliError.Message, errStruct.Message)
				}
			} else {
				require.NoError(t, err)
			}

			testutils.CompareBytes(t, tc.outGolden, tc.outBytes, out.Bytes())
		})
	}
}

// TestNewStoreCommand_NoProfileFlag tests the NewStoreCommand when the profile flag was not defined.
func TestNewStoreCommand_NoProfileFlag(t *testing.T) {
	in := strings.NewReader("")
	out := &bytes.Buffer{}

	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: out,
		Err: errBuf,
	}

	vpr := viper.New()
	vpr.Set("profile", "default")
	vpr.Set("output", "json")

	vpr.Set("default.staging_url", "test")
	vpr.Set("default.staging_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	storeCmd := NewStoreCommand(d)

	storeCmd.SetIn(ios.In)
	storeCmd.SetOut(ios.Out)
	storeCmd.SetErr(ios.Err)

	storeCmd.SetArgs([]string{"my-bucket", "testdata/StoreCommand_BucketConfig.json"})

	err := storeCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewStoreCommand_Out_NoProfile", testutils.Read(t, "NewStoreCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewStoreCommand_Err_NoProfile", testutils.Read(t, "NewStoreCommand_Err_NoProfile"), errBuf.Bytes())
}
