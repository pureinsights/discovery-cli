package cli

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
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

// TestRenderExportStatus tests the RenderExportStatus() function
func TestRenderExportStatus(t *testing.T) {
	tests := []struct {
		name                string
		err                 error
		expectedAcknowledge gjson.Result
		expectedErr         error
	}{
		{
			name:                "Render export status returns acknowledged true if no error",
			err:                 nil,
			expectedAcknowledge: gjson.Parse(`{"acknowledged": true}`),
			expectedErr:         nil,
		},
		{
			name:                "Render export status returns acknowledged false if there is an error",
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

var coreBytes, _ = os.ReadFile("testdata/core-export.zip")
var ingestionBytes, _ = os.ReadFile("testdata/ingestion-export.zip")
var queryflowBytes, _ = os.ReadFile("testdata/queryflow-export.zip")

// WorkingCoreBackupRestore mocks a working backup restore
type WorkingCoreBackupRestore struct {
	mock.Mock
}

// Get returns zip bytes as if the request worked successfully.
func (g *WorkingCoreBackupRestore) Export() ([]byte, string, error) {
	return coreBytes, "export-20251110T1455.zip", nil
}

// Import implements the interface
func (g *WorkingCoreBackupRestore) Import(discoveryPackage.OnConflict, string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// WorkingIngestionBackupRestore mocks a working backup restore
type WorkingIngestionBackupRestore struct {
	mock.Mock
}

// Get returns zip bytes as if the request worked successfully.
func (g *WorkingIngestionBackupRestore) Export() ([]byte, string, error) {
	return ingestionBytes, "export-20251110T1455.zip", nil
}

// Import implements the interface
func (g *WorkingIngestionBackupRestore) Import(discoveryPackage.OnConflict, string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// WorkingQueryFlowBackupRestore mocks a working backup restore
type WorkingQueryFlowBackupRestore struct {
	mock.Mock
}

// Get returns zip bytes as if the request worked successfully.
func (g *WorkingQueryFlowBackupRestore) Export() ([]byte, string, error) {
	return queryflowBytes, "export-20251110T1455.zip", nil
}

// Import implements the interface
func (g *WorkingQueryFlowBackupRestore) Import(discoveryPackage.OnConflict, string) (gjson.Result, error) {
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

// TestWriteExport tests the WriteExport() function
func TestWriteExport(t *testing.T) {
	testutils.ChangeDirectoryHelper(t)
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
			client:              new(WorkingIngestionBackupRestore),
			path:                filepath.Join(dir1, "export.zip"),
			expectedPath:        filepath.Join(dir1, "export.zip"),
			expectedAcknowledge: gjson.Parse(`{"acknowledged": true}`),
			err:                 nil,
		},
		{
			name:                "WriteExport correctly writes the file with no path",
			client:              new(WorkingIngestionBackupRestore),
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
				readBytes, err := os.ReadFile(tc.expectedPath)
				require.NoError(t, err)
				assert.Equal(t, readBytes, ingestionBytes)
			}
		})
	}
}

// TestExportEntitiesFromClient tests the ExportEntitiesFromClient() function
func TestExportEntitiesFromClient(t *testing.T) {
	testutils.ChangeDirectoryHelper(t)
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
			client:         new(WorkingIngestionBackupRestore),
			path:           filepath.Join(dir1, "export.zip"),
			expectedPath:   filepath.Join(dir1, "export.zip"),
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			printer:        JsonObjectPrinter(true),
			err:            nil,
		},
		{
			name:           "ExportEntitiesFromClient correctly writes the file with no path and prints with the ugly printer",
			client:         new(WorkingIngestionBackupRestore),
			path:           "",
			expectedPath:   filepath.Join(".", "export-20251110T1455.zip"),
			expectedOutput: "{\"acknowledged\":true}\n",
			printer:        nil,
			err:            nil,
		},
		// Error case
		{
			name:   "Export fails",
			client: new(FailingBackupRestore),
			path:   filepath.Join(t.TempDir(), "export.zip"),
			err:    NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}, "Could not export entities"),
		},
		{
			name:      "Printing fails",
			client:    new(WorkingIngestionBackupRestore),
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

// TestWriteExportsIntoFile tests the WriteExportsIntoFile() function.
func TestWriteExportsIntoFile(t *testing.T) {
	tests := []struct {
		name           string
		clients        []BackupRestoreClientEntry
		path           string
		expectedOutput string
		err            error
	}{
		// Working case
		{
			name:           "WriteExportsIntoFile correctly writes the zip files.",
			clients:        []BackupRestoreClientEntry{{Name: "core", Client: new(WorkingCoreBackupRestore)}, {Name: "ingestion", Client: new(WorkingIngestionBackupRestore)}, {Name: "queryflow", Client: new(WorkingQueryFlowBackupRestore)}},
			path:           filepath.Join(t.TempDir(), "export.zip"),
			expectedOutput: `{"core":{"acknowledged": true},"ingestion":{"acknowledged": true},"queryflow":{"acknowledged": true}}`,
			err:            nil,
		},
		{
			name:           "Export works for core and queryflow but fails for ingestion",
			clients:        []BackupRestoreClientEntry{{Name: "core", Client: new(WorkingCoreBackupRestore)}, {Name: "ingestion", Client: new(FailingBackupRestore)}, {Name: "queryflow", Client: new(WorkingQueryFlowBackupRestore)}},
			path:           filepath.Join(t.TempDir(), "export.zip"),
			expectedOutput: "{\"core\":{\"acknowledged\": true},\"ingestion\":{\"acknowledged\": false,\"error\":\"status: 401, body: {\\\"error\\\":\\\"unauthorized\\\"}\\n\"},\"queryflow\":{\"acknowledged\": true}}",
			err:            nil,
		},
		// Error cases
		{
			name:           "WriteExportsIntoFile receives an invalid path.",
			clients:        []BackupRestoreClientEntry{{Name: "core", Client: new(WorkingCoreBackupRestore)}, {Name: "ingestion", Client: new(WorkingIngestionBackupRestore)}, {Name: "queryflow", Client: new(WorkingQueryFlowBackupRestore)}},
			path:           filepath.Join("doesnotexist", "export.zip"),
			expectedOutput: "",
			err:            fs.ErrNotExist,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			acknowledge, err := WriteExportsIntoFile(tc.path, tc.clients)

			assert.Equal(t, tc.expectedOutput, acknowledge)
			if tc.err != nil {
				require.Error(t, err)
				if !errors.Is(tc.err, fs.ErrNotExist) {
					var errStruct Error
					require.ErrorAs(t, err, &errStruct)
					assert.EqualError(t, err, tc.err.Error())
				}
			} else {
				require.NoError(t, err)
				zipFile, err := os.ReadFile(tc.path)
				require.NoError(t, err)
				zipReader, err := zip.NewReader(bytes.NewReader(zipFile), int64(len(zipFile)))
				require.NoError(t, err)
				files := make(map[string]*zip.File, len(zipReader.File))
				for _, f := range zipReader.File {
					files[f.Name] = f
				}
				for _, client := range tc.clients {
					exportBytes, filename, err := client.Client.Export()
					if err == nil {
						exportedFile, ok := files[fmt.Sprintf("%s-%s", client.Name, filename)]
						require.True(t, ok)
						fileContent, err := exportedFile.Open()
						require.NoError(t, err)
						gotBytes, err := io.ReadAll(fileContent)
						require.NoError(t, err)
						fileContent.Close()
						assert.Equal(t, exportBytes, gotBytes)
					}
				}
			}
		})
	}
}

// TestExportEntitiesFromClients tests the TestExportEntitiesFromClients() function.
func TestExportEntitiesFromClients(t *testing.T) {
	changeDirectoryHelper(t)
	dir1 := t.TempDir()
	tests := []struct {
		name           string
		clients        []BackupRestoreClientEntry
		path           string
		expectedPath   string
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working cases
		{
			name:           "ExportEntitiesFromClients correctly prints the results with ugly printer when all the exports succeeded",
			clients:        []BackupRestoreClientEntry{{Name: "core", Client: new(WorkingCoreBackupRestore)}, {Name: "ingestion", Client: new(WorkingIngestionBackupRestore)}, {Name: "queryflow", Client: new(WorkingQueryFlowBackupRestore)}},
			path:           filepath.Join(dir1, "export.zip"),
			expectedPath:   filepath.Join(dir1, "export.zip"),
			printer:        nil,
			expectedOutput: "{\"core\":{\"acknowledged\":true},\"ingestion\":{\"acknowledged\":true},\"queryflow\":{\"acknowledged\":true}}\n",
			err:            nil,
		},
		{
			name:           "ExportEntitiesFromClients correctly prints with pretty printer when export works for core and queryflow but fails for ingestion",
			clients:        []BackupRestoreClientEntry{{Name: "core", Client: new(WorkingCoreBackupRestore)}, {Name: "ingestion", Client: new(FailingBackupRestore)}, {Name: "queryflow", Client: new(WorkingQueryFlowBackupRestore)}},
			path:           "",
			expectedPath:   filepath.Join(".", "discovery.zip"),
			printer:        JsonObjectPrinter(true),
			expectedOutput: "{\n  \"core\": {\n    \"acknowledged\": true\n  },\n  \"ingestion\": {\n    \"acknowledged\": false,\n    \"error\": \"status: 401, body: {\\\"error\\\":\\\"unauthorized\\\"}\\n\"\n  },\n  \"queryflow\": {\n    \"acknowledged\": true\n  }\n}\n",
			err:            nil,
		},
		// Error cases
		{
			name:           "WriteExportsIntoFile receives an invalid path.",
			clients:        []BackupRestoreClientEntry{{Name: "core", Client: new(WorkingCoreBackupRestore)}, {Name: "ingestion", Client: new(WorkingIngestionBackupRestore)}, {Name: "queryflow", Client: new(WorkingQueryFlowBackupRestore)}},
			path:           filepath.Join("doesnotexist", "export.zip"),
			expectedOutput: "",
			err:            NewErrorWithCause(ErrorExitCode, fs.ErrNotExist, "Could not export entities"),
		},
		{
			name:      "Printing fails",
			clients:   []BackupRestoreClientEntry{{Name: "core", Client: new(WorkingCoreBackupRestore)}, {Name: "ingestion", Client: new(WorkingIngestionBackupRestore)}, {Name: "queryflow", Client: new(WorkingQueryFlowBackupRestore)}},
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
			err := d.ExportEntitiesFromClients(tc.clients, tc.path, tc.printer)

			if tc.err != nil {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				cliError, _ := tc.err.(Error)
				if !errors.Is(cliError.Cause, fs.ErrNotExist) {
					assert.EqualError(t, err, tc.err.Error())
				} else {
					assert.Equal(t, cliError.ExitCode, errStruct.ExitCode)
					assert.Equal(t, cliError.Message, errStruct.Message)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}
