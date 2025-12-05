package cli

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
)

// BackupRestore defines methods to backup and restore entities in Discovery.
type BackupRestore interface {
	Export() ([]byte, string, error)
	Import(discoveryPackage.OnConflict, string) (gjson.Result, error)
}

// RenderExportStatus returns a JSON with the acknowledgment correct value depending on the given error.
func RenderExportStatus(err error) (gjson.Result, error) {
	if err != nil {
		acknowledged, _ := sjson.Set(`{"acknowledged": false}`, "error", err.Error())
		return gjson.Parse(acknowledged), NewErrorWithCause(ErrorExitCode, err, "Could not export entities")
	}
	return gjson.Parse(`{"acknowledged": true}`), nil
}

// WriteExport calls the Export endpoint and writes the results into a file in the given path.
func WriteExport(client BackupRestore, path string) (gjson.Result, error) {
	zipBytes, name, err := client.Export()
	if err != nil {
		return RenderExportStatus(err)
	}

	if path == "" {
		path = filepath.Join(".", name)
	}

	err = os.WriteFile(path, zipBytes, 0o644)
	return RenderExportStatus(err)
}

// ExportEntitiesFromClient exports the entities from a single Discovery product and prints the acknowledgement message
func (d discovery) ExportEntitiesFromClient(client BackupRestore, path string, printer Printer) error {
	result, err := WriteExport(client, path)
	if err != nil {
		return err
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.iostreams, result)
}

// BackupRestoreClientEntry is used to easily store the different backup and restore structs of the Discovery products.
type BackupRestoreClientEntry struct {
	Name   string
	Client BackupRestore
}

// WriteExportsIntoFile calls the export endpoints and writes the information into a file.
func WriteExportsIntoFile(path string, clients []BackupRestoreClientEntry) (string, error) {
	zipFile, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0o644,
	)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()
	result := `{}`

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	appendStatus := func(apiName string, err error) (string, error) {
		exportResult, _ := RenderExportStatus(err)
		return sjson.SetRaw(result, apiName, exportResult.Raw)
	}

	for _, entry := range clients {
		apiName := entry.Name
		client := entry.Client
		zipBytes, name, err := client.Export()
		if err != nil {
			result, err = appendStatus(apiName, err)
			if err != nil {
				return "", err
			}
			continue
		}

		h := &zip.FileHeader{
			Name:   fmt.Sprintf("%s-%s", apiName, name),
			Method: zip.Store,
		}

		fw, err := zipWriter.CreateHeader(h)
		if err != nil {
			result, err = appendStatus(apiName, err)
			if err != nil {
				return "", err
			}
			continue
		}

		_, err = fw.Write(zipBytes)
		result, err = appendStatus(apiName, err)
		if err != nil {
			return "", err
		}
	}

	return result, nil
}

// ExportEntitiesFromClients exports the entities from Discovery Core, Ingestion, and QueryFlow, writes the export files into the given path, and prints out the results.
func (d discovery) ExportEntitiesFromClients(clients []BackupRestoreClientEntry, path string, printer Printer) error {
	if path == "" {
		path = filepath.Join(".", "discovery.zip")
	}

	result, err := WriteExportsIntoFile(path, clients)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not export entities")
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.iostreams, gjson.Parse(result))
}

// ImportEntitiesToClient imports the entities to a Discovery product by reading them from the given path and using the given conflict resolution strategy.
// It then prints out the results.
func (d discovery) ImportEntitiesToClient(client BackupRestore, path string, onConflict discoveryPackage.OnConflict, printer Printer) error {
	results, err := client.Import(onConflict, path)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not import entities")
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.iostreams, results)
}

// copyImportEntitiesToTempFile copies the information of the inner zip file to the temporary file to be used in the import endpoint call.
func copyImportEntitiesToTempFile(file *zip.File, path string) error {
	if file.FileInfo().IsDir() {
		return NewError(ErrorExitCode, "The sent file should only contain the Core, Ingestion, or QueryFlow export files.")
	}

	readCloser, err := file.Open()
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not open a file contained within the zip")
	}
	defer readCloser.Close()

	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not create the temporary export file")
	}
	defer out.Close()

	if _, err := io.Copy(out, readCloser); err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not copy the file's contents")
	}

	return nil
}

// readInnerZipFiles reads the inner zip files that contain the entities to be imported.
// It writes the files into a temporary directory.
func readInnerZipFiles(tmpDir string, zipReader *zip.Reader) (map[string]string, error) {
	paths := map[string]string{}
	expectedPrefixes := []string{"core", "ingestion", "queryflow"}

	for _, file := range zipReader.File {
		// Validate zip slip vulnerability
		destPath := filepath.Join(tmpDir, file.Name)
		if !strings.HasPrefix(filepath.Clean(destPath)+string(os.PathSeparator),
			filepath.Clean(tmpDir)+string(os.PathSeparator)) {
			return nil, NewError(ErrorExitCode, "The sent file contains malicious entries.")
		}

		err := copyImportEntitiesToTempFile(file, destPath)
		if err != nil {
			return nil, err
		}

		base := filepath.Base(file.Name)
		for _, prefix := range expectedPrefixes {
			if strings.HasPrefix(base, prefix) {
				paths[prefix] = destPath
			}
		}
	}
	return paths, nil
}

// UnzipExportsToTemp parses the files read from the discovery export file.
func UnzipExportsToTemp(zipBytes []byte) (string, map[string]string, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return "", nil, NewErrorWithCause(ErrorExitCode, err, "Could not read the file with the entities")
	}

	if len(zipReader.File) > 3 {
		return "", nil, NewError(ErrorExitCode, "The sent file should only contain the Core, Ingestion, or QueryFlow export files.")
	}

	tmpDir, err := os.MkdirTemp("", "discovery-import-*")
	if err != nil {
		return "", nil, NewErrorWithCause(ErrorExitCode, err, "Could not create temporary directory to import entities")
	}

	paths, err := readInnerZipFiles(tmpDir, zipReader)
	if err != nil {
		return "", nil, err
	}

	return tmpDir, paths, nil
}

// CallImports calls the import endpoints of the given clients and adds the responses to the results JSON.
func callImports(clients []BackupRestoreClientEntry, zipPaths map[string]string, onConflict discoveryPackage.OnConflict) (string, error) {
	results := "{}"
	for _, client := range clients {
		if path, ok := zipPaths[client.Name]; ok {
			importResult, err := client.Client.Import(onConflict, path)
			if err == nil {
				results, err = sjson.SetRaw(results, client.Name, importResult.Raw)
				if err != nil {
					return "", NewErrorWithCause(ErrorExitCode, err, "Could not write import entities")
				}
			} else {
				results, err = sjson.Set(results, client.Name, err.Error())
				if err != nil {
					return "", NewErrorWithCause(ErrorExitCode, err, "Could not write import entities")
				}
			}
		}
	}

	return results, nil
}

// ImportEntitiesToClients reads a zip file that contains the zip files of exported entities. The files must have the name of the Discovery product in which the entities are going to be restored.
// The given zip file does not need to have the export files for Core, Ingestion, and QueryFlow. It can have some of them.
// The results of the imports are then printed out in a JSON to the user.
func (d discovery) ImportEntitiesToClients(clients []BackupRestoreClientEntry, path string, onConflict discoveryPackage.OnConflict, printer Printer) error {
	zipFile, err := os.ReadFile(path)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not open the file with the entities")
	}

	tmpDir, zipPaths, err := UnzipExportsToTemp(zipFile)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	results, err := callImports(clients, zipPaths, onConflict)
	if err != nil {
		return err
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.iostreams, gjson.Parse(results))
}
