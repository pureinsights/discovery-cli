package backuprestore

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

// TestNewImportCommand_ProfileFlag tests the TestNewImportCommand() function when there is a profile flag.
func TestNewImportCommand_ProfileFlag(t *testing.T) {
	importJson, err := os.ReadFile("testdata/ingestion-import.json")
	require.NoError(t, err)
	tests := []struct {
		name           string
		url            bool
		apiKey         string
		outGolden      string
		errGolden      string
		outBytes       []byte
		errBytes       []byte
		method         string
		path           string
		statusCode     int
		response       string
		file           string
		err            error
		compareOptions []testutils.CompareBytesOption
	}{
		// Working case
		{
			name:       "Import works",
			url:        true,
			apiKey:     "",
			outGolden:  "NewImportCommand_Out_ImportWorks",
			errGolden:  "NewImportCommand_Err_ImportWorks",
			outBytes:   testutils.Read(t, "NewImportCommand_Out_ImportWorks"),
			errBytes:   []byte(nil),
			method:     http.MethodPost,
			path:       "/v2/import",
			statusCode: http.StatusOK,
			response:   string(importJson),
			file:       filepath.Join("testdata", "ingestion-export.zip"),
			err:        nil,
		},

		// Error case
		{
			name:      "No URL",
			outGolden: "NewImportCommand_Out_NoURL",
			errGolden: "NewImportCommand_Err_NoURL",
			outBytes:  testutils.Read(t, "NewImportCommand_Out_NoURL"),
			errBytes:  testutils.Read(t, "NewImportCommand_Err_NoURL"),
			url:       false,
			apiKey:    "apiKey123",
			file:      filepath.Join("testdata", "ingestion-export.zip"),
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Ingestion URL is missing for profile \"default\".\nTo set the URL for the Discovery Ingestion API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery ingestion config --profile \"default\""),
		},
		{
			name:           "Import Fails because the sent file does not exist",
			url:            true,
			apiKey:         "",
			outGolden:      "NewImportCommand_Out_FileDoesNotExist",
			errGolden:      "NewImportCommand_Err_FileDoesNotExist",
			outBytes:       testutils.Read(t, "NewImportCommand_Out_FileDoesNotExist"),
			errBytes:       testutils.Read(t, "NewImportCommand_Err_FileDoesNotExist"),
			method:         http.MethodPost,
			path:           "/v2/import",
			response:       "",
			file:           filepath.Join("doesnotexist", "ingestion-export.zip"),
			err:            cli.NewErrorWithCause(cli.ErrorExitCode, fmt.Errorf("file does not exist: %s", filepath.Join("doesnotexist", "ingestion-export.zip")), "Could not import entities"),
			compareOptions: []testutils.CompareBytesOption{testutils.WithNormalizePaths()},
		},
		{
			name:       "Import fails",
			url:        true,
			apiKey:     "",
			outGolden:  "NewImportCommand_Out_ImportFails",
			errGolden:  "NewImportCommand_Err_ImportFails",
			outBytes:   testutils.Read(t, "NewImportCommand_Out_ImportFails"),
			errBytes:   testutils.Read(t, "NewImportCommand_Err_ImportFails"),
			method:     http.MethodPost,
			path:       "/v2/import",
			statusCode: http.StatusUnauthorized,
			response:   `{"error":"unauthorized"}`,
			file:       filepath.Join("testdata", "ingestion-export.zip"),
			err:        cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}, "Could not import entities"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				testutils.HttpHandler(t, tc.statusCode, "application/json", tc.response, func(t *testing.T, r *http.Request) {
					assert.Equal(t, tc.method, r.Method)
					assert.Equal(t, tc.path, r.URL.Path)
				}))

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
			vpr.Set("output", "json")
			if tc.url {
				vpr.Set("default.ingestion_url", srv.URL)
			}
			if tc.apiKey != "" {
				vpr.Set("default.ingestion_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, t.TempDir())

			importCmd := NewImportCommand(d)

			importCmd.SetIn(ios.In)
			importCmd.SetOut(ios.Out)
			importCmd.SetErr(ios.Err)

			importCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			args := []string{}
			args = append(args, "--file")
			args = append(args, tc.file)

			args = append(args, "--on-conflict")
			args = append(args, string(discoveryPackage.OnConflictFail))

			importCmd.SetArgs(args)

			err := importCmd.Execute()
			if tc.err != nil {
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
				testutils.CompareBytes(t, tc.errGolden, tc.errBytes, errBuf.Bytes(), tc.compareOptions...)
			} else {
				require.NoError(t, err)
			}

			testutils.CompareBytes(t, tc.outGolden, tc.outBytes, out.Bytes())
		})
	}
}

// TestNewImportCommand_NoProfileFlag tests the NewImportCommand when the profile flag was not defined.
func TestNewImportCommand_NoProfileFlag(t *testing.T) {
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

	vpr.Set("default.ingestion_url", "test")
	vpr.Set("default.ingestion_key", "test")

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	getCmd := NewImportCommand(d)

	getCmd.SetIn(ios.In)
	getCmd.SetOut(ios.Out)
	getCmd.SetErr(ios.Err)

	getCmd.SetArgs([]string{"--file", "testdata/ingestion-export.zip"})

	err := getCmd.Execute()
	require.Error(t, err)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewImportCommand_Out_NoProfile", testutils.Read(t, "NewImportCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewImportCommand_Err_NoProfile", testutils.Read(t, "NewImportCommand_Err_NoProfile"), errBuf.Bytes())
}
