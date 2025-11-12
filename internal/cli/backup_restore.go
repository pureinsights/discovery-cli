package cli

import (
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
