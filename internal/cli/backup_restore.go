package cli

import (
	"os"

	"github.com/tidwall/gjson"

	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
)

type BackupRestore interface {
	Export() ([]byte, string, error)
	Import(discoveryPackage.OnConflict, string) (gjson.Result, error)
}

func (d discovery) ExportEntities(client BackupRestore) ([]byte, string, gjson.Result, error) {
	zipBytes, filename, err := client.Export()
	if err != nil {
		return []byte(nil), "", gjson.Parse(`{"acknowledged":false}`), NewErrorWithCause(ErrorExitCode, err, "Could not export entities")
	}

	return zipBytes, filename, gjson.Parse(`{"acknowledged":true}`), nil
}

func (d discovery) WriteExport(client BackupRestore, file string) (gjson.Result, error) {
	zipBytes, filename, _, err := d.ExportEntities(client)
	if err != nil {
		return gjson.Parse(`{"acknowledged":false}`), err
	}
	if file != "" {
		filename = file
	}

	err = os.WriteFile(filename, zipBytes, 0o644)
	if err == nil {
		return gjson.Parse(`{"acknowledged":true}`), err
	} else {
		return gjson.Parse(`{"acknowledged":false}`), err
	}
}
