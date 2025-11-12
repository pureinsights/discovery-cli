package cli

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestRenderExportStatus tests the RenderExportStatus() function()
func TestRenderExportStatus(t *testing.T) {
	tests := []struct {
		name                string
		err                 error
		expectedAcknowledge gjson.Result
		expectedErr         error
	}{
		{
			name:                "Render export status returns acknoledged true if no error",
			err:                 nil,
			expectedAcknowledge: gjson.Parse(`{"acknowledged": true}`),
			expectedErr:         nil,
		},
		{
			name:                "Render export status returns acknoledged false if no error",
			err:                 errors.New("write failed"),
			expectedAcknowledge: gjson.Parse(`{"acknowledged": false,"error":"write failed"}`),
			expectedErr:         NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not export entities"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			acknowledge, err := RenderExportStatus(tc.err)
			assert.Equal(t, tc.expectedAcknowledge, acknowledge)
			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// WorkingBackupRestore mocks a working backup restore
type WorkingBackupRestore struct {
	mock.Mock
}

var zipBytes, _ = os.ReadFile("testdata/test-export.zip")

// Get returns zip bytes as if the request worked successfully.
func (g *WorkingBackupRestore) Export() ([]byte, string, error) {
	return zipBytes, "export-20251110T1455.zip", nil
}

// Import implements the interface
func (g *WorkingBackupRestore) Import(discoveryPackage.OnConflict, string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// FailingBackupRestore mocks a failing backup restore
type FailingBackupRestore struct {
	mock.Mock
}

// Get returns an error as if the request failed.
func (g *FailingBackupRestore) Export() ([]byte, string, error) {
	return []byte(nil), "discovery.zip", discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// Import implements the interface
func (g *FailingBackupRestore) Import(discoveryPackage.OnConflict, string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// changeDirectoryHelper changes the working directory to t.TempDir()
func changeDirectoryHelper(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(wd) })
	return tmp
}

// TestWriteExport tests the WriteExport() function
func TestWriteExport(t *testing.T) {
	changeDirectoryHelper(t)
	dir1 := t.TempDir()
	tests := []struct {
		name                string
		client              BackupRestore
		path                string
		expectedPath        string
		expectedAcknowledge gjson.Result
		err                 error
	}{
		// Working case
		{
			name:                "WriteExport correctly writes the file",
			client:              new(WorkingBackupRestore),
			path:                filepath.Join(dir1, "export.zip"),
			expectedPath:        filepath.Join(dir1, "export.zip"),
			expectedAcknowledge: gjson.Parse(`{"acknowledged": true}`),
			err:                 nil,
		},
		{
			name:                "WriteExport correctly writes the file with no path",
			client:              new(WorkingBackupRestore),
			path:                "",
			expectedPath:        filepath.Join(".", "export-20251110T1455.zip"),
			expectedAcknowledge: gjson.Parse(`{"acknowledged": true}`),
			err:                 nil,
		},
		{
			name:                "Export fails",
			client:              new(FailingBackupRestore),
			path:                filepath.Join(t.TempDir(), "export.zip"),
			expectedAcknowledge: gjson.Parse("{\"acknowledged\": false,\"error\":\"status: 401, body: {\\\"error\\\":\\\"unauthorized\\\"}\\n\"}"),
			err:                 NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}, "Could not export entities"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			acknowledge, err := WriteExport(tc.client, tc.path)

			assert.Equal(t, tc.expectedAcknowledge, acknowledge)
			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				_, err := os.Stat(tc.expectedPath)
				require.NoError(t, err)
			}
		})
	}
}

// TestExportEntitiesFromClient tests the ExportEntitiesFromClient() function
func TestExportEntitiesFromClient(t *testing.T) {
	changeDirectoryHelper(t)
	dir1 := t.TempDir()
	tests := []struct {
		name           string
		client         BackupRestore
		path           string
		expectedPath   string
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "ExportEntitiesFromClient correctly prints acknowledged true with pretty printer",
			client:         new(WorkingBackupRestore),
			path:           filepath.Join(dir1, "export.zip"),
			expectedPath:   filepath.Join(dir1, "export.zip"),
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			printer:        JsonObjectPrinter(true),
			err:            nil,
		},
		{
			name:           "ExportEntitiesFromClient correctly writes the file with no path and prints with the ugly printer",
			client:         new(WorkingBackupRestore),
			path:           "",
			expectedPath:   filepath.Join(".", "export-20251110T1455.zip"),
			expectedOutput: "{\"acknowledged\":true}\n",
			printer:        nil,
			err:            nil,
		},
		{
			name:   "Export fails",
			client: new(FailingBackupRestore),
			path:   filepath.Join(t.TempDir(), "export.zip"),
			err:    NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}, "Could not export entities"),
		},
		{
			name:      "Printing fails",
			client:    new(WorkingBackupRestore),
			printer:   nil,
			outWriter: testutils.ErrWriter{Err: errors.New("write failed")},
			err:       NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
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

			d := NewDiscovery(&ios, viper.New(), "")
			err := d.ExportEntitiesFromClient(tc.client, tc.path, tc.printer)

			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}
