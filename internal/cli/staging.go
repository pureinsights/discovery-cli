package cli

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/tidwall/gjson"
)

// StagingContentController defines the methods to interact with a bucket's content.
type StagingContentController interface {
	Scroll(filters, projections gjson.Result, size *int) ([]gjson.Result, error)
	Count(filters gjson.Result) (gjson.Result, error)
}

// writeRecordsToFile writes the records obtained from the scroll to a JSON file in a temporary directory.
func writeRecordsToFile(records []gjson.Result, bucket string) (string, error) {
	dir, err := os.MkdirTemp("", fmt.Sprintf("dump-%s-*", bucket))
	if err != nil {
		defer os.RemoveAll(dir)
		return "", NormalizeWriteFileError(os.TempDir(), err)
	}

	for _, record := range records {
		transaction := record.Get("transaction").String()

		err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s.json", transaction)), []byte(record.Raw), 0o644)
		if err != nil {
			defer os.RemoveAll(dir)
			return dir, NormalizeWriteFileError(filepath.Join(dir, fmt.Sprintf("%s.json", transaction)), err)
		}
	}
	return dir, nil
}

// writeToZip is a function that separates writing the record files to the zip in order to reduce zipRecords' cognitive complexity.
func writeToZip(dir, path string, d fs.DirEntry, zipWriter *zip.Writer) error {
	info, err := d.Info()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Method = zip.Deflate

	relPath, err := filepath.Rel(dir, path)
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(relPath)

	fw, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	recordFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer recordFile.Close()

	_, err = io.Copy(fw, recordFile)
	if err != nil {
		return NormalizeWriteFileError(path, err)
	}

	return nil
}

// zipRecords zips the temporary directory into a file containing all of the records.
func zipRecords(file, dir string) error {
	zipFile, err := os.Create(file)
	if err != nil {
		return NormalizeWriteFileError(file, err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			return nil
		}

		walkErr = writeToZip(dir, path, d, zipWriter)
		if walkErr != nil {
			return walkErr
		}

		return nil
	})

	if err != nil {
		return NormalizeWriteFileError(dir, err)
	}

	return nil
}

// DumpConfig is a struct that contains fields necessary to dump a bucket.
type DumpConfig struct {
	File        string
	Filters     gjson.Result
	Projections gjson.Result
	Size        *int
}

// DumpBucket scrolls the contents of a bucket based on the given filters, projections and maximum page size.
func (d discovery) DumpBucket(client StagingContentController, bucketName, file string, filters, projections gjson.Result, size *int, printer Printer) error {
	records, err := client.Scroll(filters, projections, size)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not scroll the bucket with name %q.", bucketName)
	}

	dir, err := writeRecordsToFile(records, bucketName)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not write records to temporary folder.")
	}

	defer os.RemoveAll(dir)
	err = zipRecords(file, dir)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not write dump to file.")
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.IOStreams(), gjson.Parse(`{"acknowledged": true}`))
}

// CountBucket counts the number of records in a bucket based on the given filters.
func (d discovery) CountBucket(client StagingContentController, bucketName string, filters gjson.Result, printer Printer) error {
	result, err := client.Count(filters)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not count the bucket with name %q.", bucketName)
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.IOStreams(), result)
}
