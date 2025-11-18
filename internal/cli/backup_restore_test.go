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

var coreBytes, _ = os.ReadFile("testdata/core-export.zip")
var ingestionBytes, _ = os.ReadFile("testdata/ingestion-export.zip")
var queryflowBytes, _ = os.ReadFile("testdata/queryflow-export.zip")
var coreImport, _ = os.ReadFile("testdata/core-import.json")
var ingestionImport, _ = os.ReadFile("testdata/ingestion-import.json")
var queryflowImport, _ = os.ReadFile("testdata/queryflow-import.json")

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
	return gjson.ParseBytes(coreImport), nil
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
	return gjson.ParseBytes(ingestionImport), nil
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
	return gjson.ParseBytes(queryflowImport), nil
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
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
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

// TestImportEntitiesFromClient tests the ImportEntitiesFromClient() function
func TestImportEntitiesFromClient(t *testing.T) {
	tests := []struct {
		name           string
		client         BackupRestore
		path           string
		onConfict      discoveryPackage.OnConflict
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "ImportEntitiesFromClient correctly prints import results with the pretty printer",
			client:         new(WorkingCoreBackupRestore),
			path:           filepath.Join("testdata", "core-export.zip"),
			onConfict:      discoveryPackage.OnConflictUpdate,
			expectedOutput: "{\n  \"Credential\": [\n    {\n      \"id\": \"3b32e410-2f33-412d-9fb8-17970131921c\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"458d245a-6ed2-4c2b-a73f-5540d550a479\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"46cb4fff-28be-4901-b059-1dd618e74ee4\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"4957145b-6192-4862-a5da-e97853974e9f\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"5c09589e-b643-41aa-a766-3b7fc3660473\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"6dd2177f-0196-42d8-9468-0053a5c1127a\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"822b2d33-20a2-4df4-aebf-a1cee5acdce7\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"837196a6-1ac5-4b0c-a24a-4b9d092e6260\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"84f66cd4-a28b-4e66-94e1-a3dc9f083bbd\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"8c243a1d-9384-421d-8f99-4ef28d4e0ab0\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"9be0e625-a510-46c5-8130-438823f849c2\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"9d438628-5981-49c5-9426-0d328fd16370\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"b4d9ee85-9775-49fa-8dfb-b3e5ce2f619e\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"f643fe55-18db-48e4-9d3f-335d0f5f5348\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"f64a5451-3716-45c4-8158-350f30e1cbdb\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"f6c4585b-4e65-4359-9aee-e995ba09f69e\",\n      \"status\": 200\n    }\n  ],\n  \"Server\": [\n    {\n      \"id\": \"21029da3-041c-43b5-a67e-870251f2f6a6\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"226e8a0b-5016-4ebe-9963-1461edd39d0a\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"2b839453-ddad-4ced-8e13-2c7860af60a7\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"3ab2e3c0-5459-4f19-9e89-f8282d111eba\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"3edc9c72-a875-49d7-8929-af09f3e9c01c\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"6f2ddfd5-154a-4534-8f29-b1569ac23b8a\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"6ffc7784-481e-4da8-8ee3-6817d15a757c\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"74160a12-bcf6-4778-8944-4a4b2a7c4be1\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"741df47e-208f-47c1-812f-53cc62c726af\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"7cd191c0-d8ab-44f7-923f-2e32d044ced2\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"8f14c11c-bb66-49d3-aa2a-dedff4608c17\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"986ce864-af76-4fcb-8b4f-f4e4c6ab0951\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"a798cd5b-aa7a-4fc5-9292-1de6fe8e8b7f\",\n      \"status\": 200\n    },\n    {\n      \"id\": \"f6950327-3175-4a98-a570-658df852424a\",\n      \"status\": 200\n    }\n  ]\n}\n",
			printer:        JsonObjectPrinter(true),
			err:            nil,
		},
		{
			name:           "ImportEntitiesFromClient correctly prints import results with the ugly printer",
			client:         new(WorkingQueryFlowBackupRestore),
			path:           filepath.Join("testdata", "queryflow-export.zip"),
			onConfict:      discoveryPackage.OnConflictFail,
			expectedOutput: "{\"Endpoint\":[{\"id\":\"2fee5e27-4147-48de-ba1e-d7f32476a4a2\",\"status\":204},{\"id\":\"4ef9da31-2ba6-442c-86bb-1c9566dac4c2\",\"status\":204},{\"id\":\"c4ffddc0-9e80-4809-ad4d-f01c4e0dba71\",\"status\":204},{\"id\":\"cf56470f-0ab4-4754-b05c-f760669315af\",\"status\":204}],\"Processor\":[{\"id\":\"019ecd8e-76c9-41ee-b047-299b8aa14aba\",\"status\":204},{\"id\":\"0a7caa9b-99aa-4a63-aa6d-a1e40941984d\",\"status\":204},{\"id\":\"3393f6d9-94c1-4b70-ba02-5f582727d998\",\"status\":204},{\"id\":\"5f125024-1e5e-4591-9fee-365dc20eeeed\",\"status\":204},{\"id\":\"628d4b24-84cc-4070-8eed-c3155cf40fe9\",\"status\":204},{\"id\":\"746ba681-246a-4dba-aac0-58848ac97725\",\"status\":204},{\"id\":\"86e7f920-a4e4-4b64-be84-5437a7673db8\",\"status\":204},{\"id\":\"88022257-f5bc-4705-968e-81dae0c486d3\",\"status\":204},{\"id\":\"8a399b1c-95fc-406c-a220-7d321aaa7b0e\",\"status\":204},{\"id\":\"8e9ce4af-0f0b-44c7-bff7-c3c4f546e577\",\"status\":204},{\"id\":\"90b9e14f-1ba2-47cb-be42-77c4081e78a2\",\"status\":204},{\"id\":\"a5ee116b-bd95-474e-9d50-db7be988b196\",\"status\":204},{\"id\":\"aa02b328-76aa-4fcb-9eb1-6086d845adbd\",\"status\":204},{\"id\":\"b5c25cd3-e7c9-4fd2-b7e6-2bcf6e2caf89\",\"status\":204},{\"id\":\"c80758d7-7989-4c23-8f8f-b92497e3ab90\",\"status\":204},{\"id\":\"eb9499c3-e134-4f4a-8aaf-288da68e68f0\",\"status\":204},{\"id\":\"f3d696ca-3c5d-4cdd-a569-b2619f7a1470\",\"status\":204},{\"id\":\"fd2d5f86-bdf0-44f7-ad70-fdd636c52c4e\",\"status\":204}]}\n",
			printer:        nil,
			err:            nil,
		},
		// Error case
		{
			name:   "Import fails",
			client: new(FailingBackupRestore),
			path:   filepath.Join(t.TempDir(), "core-export.zip"),
			err:    NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}, "Could not import entities"),
		},
		{
			name:      "Printing fails",
			client:    new(WorkingQueryFlowBackupRestore),
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
			err := d.ImportEntitiesToClient(tc.client, tc.path, tc.onConfict, tc.printer)

			bufString := buf.String()
			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, bufString)
			}
		})
	}
}

// createZipSlipPayload creates a malicious zip to test the zip slip detection.
func createZipSlipPayload(t *testing.T) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	hdr := &zip.FileHeader{
		Name:   "../../core-export.txt",
		Method: zip.Deflate,
	}
	f, err := zw.CreateHeader(hdr)
	require.NoError(t, err)

	_, err = f.Write([]byte("malicious zip file"))
	require.NoError(t, err)

	require.NoError(t, zw.Close())
	return buf.Bytes()
}

// TestUnzipExportsToTemp tests the UnzipExportsToTemp() function.
func TestUnzipExportsToTemp(t *testing.T) {
	correctZip, err := os.ReadFile("testdata/discovery.zip")
	require.NoError(t, err)
	fourFilesZip, err := os.ReadFile("testdata/4-files.zip")
	require.NoError(t, err)
	directoryZip, err := os.ReadFile("testdata/directory.zip")
	require.NoError(t, err)
	coreQueryFlowZip, err := os.ReadFile("testdata/OnlyCoreQueryFlow.zip")
	require.NoError(t, err)

	tests := []struct {
		name             string
		zipBytes         []byte
		expectedPrefixes []string
		err              error
	}{
		// Working cases
		{
			name:             "Receives a zip file with Core, QueryFlow, and Ingestion exports",
			zipBytes:         correctZip,
			expectedPrefixes: []string{"core", "ingestion", "queryflow"},
			err:              nil,
		},
		{
			name:             "Receives a file with only Core and QueryFlow exports",
			zipBytes:         coreQueryFlowZip,
			expectedPrefixes: []string{"core", "queryflow"},
			err:              nil,
		},
		// Error cases
		{
			name:     "Receives a zip file with too many files",
			zipBytes: fourFilesZip,
			err:      NewError(ErrorExitCode, "The sent file should only contain the Core, Ingestion, or QueryFlow export files."),
		},
		{
			name:     "Receives an invalid zip",
			zipBytes: []byte("this is not a valid zip"),
			err:      NewErrorWithCause(ErrorExitCode, errors.New("zip: not a valid zip file"), "Could not read the file with the entities"),
		},
		{
			name:     "Receives a zip file with a directory entry",
			zipBytes: directoryZip,
			err:      NewErrorWithCause(ErrorExitCode, err, "The sent file should only contain the Core, Ingestion, or QueryFlow export files."),
		},
		{
			name:     "Receives a malicious zip file with a zip slip",
			zipBytes: createZipSlipPayload(t),
			err:      NewError(ErrorExitCode, "The sent file contains malicious entries."),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualTmpDir, actualPaths, actualErr := UnzipExportsToTemp(tc.zipBytes)

			if tc.err != nil {
				require.Error(t, actualErr)
				var errStruct Error
				require.ErrorAs(t, actualErr, &errStruct)
				assert.EqualError(t, actualErr, tc.err.Error())

				assert.Equal(t, "", actualTmpDir)
				assert.Equal(t, map[string]string(nil), actualPaths)
				return
			} else {
				require.NoError(t, actualErr)
				defer os.RemoveAll(actualTmpDir)

				for _, prefix := range tc.expectedPrefixes {
					actualPath, ok := actualPaths[prefix]
					assert.True(t, ok)

					fileInfo, statErr := os.Stat(actualPath)
					require.NoError(t, statErr)
					assert.False(t, fileInfo.IsDir())
				}
			}

		})
	}
}

// TestImportEntitiesFromClients tests the ImportEntitiesFromClients() function.
func TestImportEntitiesFromClients(t *testing.T) {
	tests := []struct {
		name        string
		clients     []BackupRestoreClientEntry
		path        string
		printer     Printer
		goldenFile  string
		goldenBytes []byte
		outWriter   io.Writer
		err         error
	}{
		// Working cases
		{
			name:        "ImportEntitiesFromClients correctly prints with pretty printer when one of the imports fails",
			clients:     []BackupRestoreClientEntry{{Name: "core", Client: new(WorkingCoreBackupRestore)}, {Name: "ingestion", Client: new(FailingBackupRestore)}, {Name: "queryflow", Client: new(WorkingQueryFlowBackupRestore)}},
			path:        "testdata/discovery.zip",
			printer:     JsonObjectPrinter(false),
			goldenFile:  "FailingIngestionImport",
			goldenBytes: testutils.Read(t, "FailingIngestionImport"),
			err:         nil,
		},
		{
			name:        "ImportEntitiesFromClients correctly prints the results with ugly printer when the imports succeed",
			clients:     []BackupRestoreClientEntry{{Name: "core", Client: new(WorkingCoreBackupRestore)}, {Name: "ingestion", Client: new(WorkingIngestionBackupRestore)}, {Name: "queryflow", Client: new(WorkingQueryFlowBackupRestore)}},
			path:        "testdata/discovery.zip",
			printer:     nil,
			goldenFile:  "UglyImport",
			goldenBytes: testutils.Read(t, "UglyImport"),
			err:         nil,
		},
		{
			name:        "ImportEntitiesFromClients correctly prints with pretty printer when the imports and it only prints the results of the received products",
			clients:     []BackupRestoreClientEntry{{Name: "core", Client: new(WorkingCoreBackupRestore)}, {Name: "ingestion", Client: new(FailingBackupRestore)}, {Name: "queryflow", Client: new(WorkingQueryFlowBackupRestore)}},
			path:        "testdata/OnlyCoreQueryFlow.zip",
			printer:     JsonObjectPrinter(true),
			goldenFile:  "PrettyImport",
			goldenBytes: testutils.Read(t, "PrettyImport"),
			err:         nil,
		},
		// Error cases
		{
			name:    "The given file does not exist",
			clients: []BackupRestoreClientEntry{{Name: "core", Client: new(WorkingCoreBackupRestore)}, {Name: "ingestion", Client: new(WorkingIngestionBackupRestore)}, {Name: "queryflow", Client: new(WorkingQueryFlowBackupRestore)}},
			path:    filepath.Join("doesnotexist", "export.zip"),
			err:     NewErrorWithCause(ErrorExitCode, fs.ErrNotExist, "Could not open the file with the entities"),
		},
		{
			name:    "UnzipExportsToTemp fails",
			clients: []BackupRestoreClientEntry{{Name: "core", Client: new(WorkingCoreBackupRestore)}, {Name: "ingestion", Client: new(WorkingIngestionBackupRestore)}, {Name: "queryflow", Client: new(WorkingQueryFlowBackupRestore)}},
			path:    "testdata/4-files.zip",
			err:     NewError(ErrorExitCode, "The sent file should only contain the Core, Ingestion, or QueryFlow export files."),
		},
		{
			name:      "Printing fails",
			clients:   []BackupRestoreClientEntry{{Name: "core", Client: new(WorkingCoreBackupRestore)}, {Name: "ingestion", Client: new(WorkingIngestionBackupRestore)}, {Name: "queryflow", Client: new(WorkingQueryFlowBackupRestore)}},
			path:      "testdata/discovery.zip",
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
			err := d.ImportEntitiesToClients(tc.clients, tc.path, discoveryPackage.OnConflictUpdate, tc.printer)

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
				testutils.CompareBytes(t, tc.goldenFile, tc.goldenBytes, buf.Bytes())
			}
		})
	}
}
