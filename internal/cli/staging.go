package cli

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/tidwall/gjson"
)

// StagingBucketController defines the methods to interact with buckets.
type StagingBucketController interface {
	Create(bucket string, options gjson.Result) (gjson.Result, error)
	Get(bucket string) (gjson.Result, error)
	CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error)
	DeleteIndex(bucket, index string) (gjson.Result, error)
	Delete(bucket string) (gjson.Result, error)
}

// StagingContentController defines the methods to interact with a bucket's content.
type StagingContentController interface {
	Scroll(filters, projections gjson.Result, size *int) ([]gjson.Result, error)
}

// updateIndices updates the indices in a bucket with the new configuration.
// If any update fails, the function returns an error.
func updateIndices(client StagingBucketController, bucketName string, oldIndices []gjson.Result, newIndices gjson.Result) error {
	for _, index := range oldIndices {
		indexName := index.Get("name").String()

		oldIndexExists := newIndices.Get(fmt.Sprintf("#(name==%q)", indexName)).Exists()
		if !oldIndexExists {
			indexAck, err := client.DeleteIndex(bucketName, indexName)
			if err != nil || !(indexAck.Get("acknowledged").Bool()) {
				return NewErrorWithCause(ErrorExitCode, err, "Could not delete index with name %q of bucket %q.", indexName, bucketName)
			}
		}
	}
	for _, index := range newIndices.Array() {
		indexName := index.Get("name").String()

		indexAck, err := client.CreateIndex(bucketName, indexName, index.Get("fields").Array())
		if err != nil || !(indexAck.Get("acknowledged").Bool()) {
			return NewErrorWithCause(ErrorExitCode, err, "Could not update index with name %q of bucket %q.", indexName, bucketName)
		}
	}

	return nil
}

// callUpdateIndices is an auxiliary function to reduce the complexity of StoreBucket().
func callUpdateIndices(client StagingBucketController, bucketName string, options gjson.Result) error {
	bucketInfo, err := client.Get(bucketName)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not get bucket with name %q to update it.", bucketName)
	}
	oldIndices := bucketInfo.Get("indices").Array()
	newIndices := options.Get("indices")
	err = updateIndices(client, bucketName, oldIndices, newIndices)
	if err != nil {
		return err
	}
	return nil
}

// StoreBucket creates or updates a bucket if it exists. It receives an options parameter that contains the configuration of the bucket.
func (d discovery) StoreBucket(client StagingBucketController, bucketName string, options gjson.Result, printer Printer) error {
	const bucketError string = "Could not create bucket with name %q."
	result, err := client.Create(bucketName, options)
	if err != nil {
		discoveryErr, ok := err.(discoveryPackage.Error)
		if !ok {
			return NewErrorWithCause(ErrorExitCode, err, bucketError, bucketName)
		}

		if discoveryErr.Status != http.StatusConflict {
			return NewErrorWithCause(ErrorExitCode, err, bucketError, bucketName)
		}

		if options.Get("indices").Exists() {
			err = callUpdateIndices(client, bucketName, options)
			if err != nil {
				return err
			}
		} else {
			return NewErrorWithCause(ErrorExitCode, err, bucketError, bucketName)
		}
	}

	result, err = client.Get(bucketName)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not get the information of bucket with name %q.", bucketName)
	}
	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.IOStreams(), result)
}

// DeleteBucket deletes the bucket with the given name.
func (d discovery) DeleteBucket(client StagingBucketController, bucketName string, printer Printer) error {
	result, err := client.Delete(bucketName)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not delete the bucket with name %q.", bucketName)
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.IOStreams(), result)
}

func writeDocumentsToFile(documents []gjson.Result, bucket string) (string, error) {
	dir, err := os.MkdirTemp("", fmt.Sprintf("dump-%s-*", bucket))
	if err != nil {
		defer os.RemoveAll(dir)
		return "", err
	}

	for _, document := range documents {
		transaction := document.Get("transaction").String()

		err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s.json", transaction)), []byte(document.Raw), 0o644)
		if err != nil {
			defer os.RemoveAll(dir)
			return "", NormalizeWriteFileError(filepath.Join(dir, fmt.Sprintf("%s.json", transaction)), err)
		}
	}
	return dir, nil
}

func zipDocuments(file, dir string) error {
	zipFile, err := os.Create(file)
	if err != nil {
		return NormalizeWriteFileError(file, err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

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

		documentFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer documentFile.Close()

		_, err = io.Copy(fw, documentFile)
		if err != nil {
			return err
		}

		return nil
	})
}

// DumpBucket scrolls the contents of a bucket based on the given filters, projections and maximum page size.
func (d discovery) DumpBucket(client StagingContentController, bucketName, file string, filters, projections gjson.Result, size *int, printer Printer) error {
	records, err := client.Scroll(filters, projections, size)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not scroll the bucket with name %q.", bucketName)
	}

	dir, err := writeDocumentsToFile(records, bucketName)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not write documents to temporary folder.")
	}

	defer os.RemoveAll(dir)
	err = zipDocuments(file, dir)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not write dump to file.")
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.IOStreams(), gjson.Parse(`{"acknowledged": true}`))
}
