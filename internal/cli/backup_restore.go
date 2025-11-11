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

type BackupRestore interface {
	Export() ([]byte, string, error)
	Import(discoveryPackage.OnConflict, string) (gjson.Result, error)
}

func RenderExportStatus(err error) (gjson.Result, error) {
	if err != nil {
		acknowledged, _ := sjson.Set(`{"acknowledged": false}`, "error", err.Error())
		return gjson.Parse(acknowledged), NewErrorWithCause(ErrorExitCode, err, "Could not export entities")
	}
	return gjson.Parse(`{"acknowledged": true}`), nil
}

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

type BackupRestoreClientEntry struct {
	Name   string
	Client BackupRestore
}

func (d discovery) ExportEntitiesFromClients(clients []BackupRestoreClientEntry, path string, printer Printer) error {
	result := `{}`

	if path == "" {
		path = filepath.Join(".", "discovery.zip")
	}

	appendStatus := func(apiName string, err error) (string, error) {
		exportResult, _ := RenderExportStatus(err)
		return sjson.SetRaw(result, apiName, exportResult.Raw)
	}

	zipFile, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0o644,
	)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, entry := range clients {
		apiName := entry.Name
		client := entry.Client
		zipBytes, name, err := client.Export()
		if err != nil {
			result, err = appendStatus(apiName, err)
			if err != nil {
				return err
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
				return err
			}
			continue
		}

		_, err = fw.Write(zipBytes)
		result, err = appendStatus(apiName, err)
		if err != nil {
			return err
		}
	}

	if printer == nil {
		printer = JsonObjectPrinter(false)
	}

	return printer(*d.iostreams, gjson.Parse(result))
}

// func (d discovery) ExportEntities(client BackupRestore) ([]byte, string, gjson.Result, error) {
// 	zipBytes, filename, err := client.Export()
// 	if err != nil {
// 		return []byte(nil), "", gjson.Parse(`{"acknowledged":false}`), NewErrorWithCause(ErrorExitCode, err, "Could not export entities")
// 	}
//
// 	return zipBytes, filename, gjson.Parse(`{"acknowledged":true}`), nil
// }
//
// func (d discovery) WriteExport(client BackupRestore, file string) (gjson.Result, error) {
// 	zipBytes, filename, _, err := d.ExportEntities(client)
// 	if err != nil {
// 		return gjson.Parse(`{"acknowledged":false}`), err
// 	}
// 	if file != "" {
// 		filename = file
// 	}
//
// 	err = os.WriteFile(filename, zipBytes, 0o644)
// 	if err == nil {
// 		return gjson.Parse(`{"acknowledged":true}`), err
// 	} else {
// 		return gjson.Parse(`{"acknowledged":false}`), err
// 	}
// }
