package cli

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
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
	fmt.Println("File written at " + path)
	return RenderExportStatus(err)
}

// ExportEntitiesFromClient exports the entities from a single Discovery Component and prints the acknowledgement message
func (d discovery) ExportEntitiesFromClient(client BackupRestore, path string, printer Printer) error {
	result, err := WriteExport(client, path)
	if err != nil {
		return err
	}

	if printer == nil {
		printer = JsonObjectPrinter(false)
	}

	return printer(*d.iostreams, result)
}

// BackupRestoreClientEntry is used to easily store the different backup and restore structs of the Discovery components.
type BackupRestoreClientEntry struct {
	Name   string
	Client BackupRestore
}

// WriteExportsIntoZip calls the export endpoints and writes the information into a file.
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
		printer = JsonObjectPrinter(false)
	}

	return printer(*d.iostreams, gjson.Parse(result))
}
