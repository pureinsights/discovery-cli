package cli

import (
	"net/http"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// StagingBucketController defines the methods to interact with buckets.
type StagingBucketController interface {
	Create(bucket string, options gjson.Result) (gjson.Result, error)
	Get(bucket string) (gjson.Result, error)
	Delete(bucket string) (gjson.Result, error)
	CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error)
}

func updateIndices(client StagingBucketController, bucketName string, indices []gjson.Result) (gjson.Result, error) {
	indexResults := "{}"
	for _, index := range indices {
		indexName := index.Get("name").String()

		indexAck, err := client.CreateIndex(bucketName, indexName, index.Get("fields").Array())
		if err != nil {
			return gjson.Result{}, NewErrorWithCause(ErrorExitCode, err, "Could not update index with name %q of bucket %q.", indexName, bucketName)
		}
		indexResults, err = sjson.SetRaw(indexResults, "indices."+indexName, indexAck.Raw)
		if err != nil {
			return gjson.Result{}, NewErrorWithCause(ErrorExitCode, err, "Could not update index with name %q of bucket %q.", indexName, bucketName)
		}
	}
	return gjson.Parse(indexResults), nil
}

// StoreBucket creates or updates a bucket if it exists. It receives an options parameter that contains the configuration of the bucket.
func (d discovery) StoreBucket(client StagingBucketController, bucketName string, options gjson.Result, printer Printer) error {
	result, err := client.Create(bucketName, options)
	if err != nil {
		discoveryErr, ok := err.(discoveryPackage.Error)
		if !ok {
			return NewErrorWithCause(ErrorExitCode, err, "Could not create bucket with name %q.", bucketName)
		}

		if discoveryErr.Status != http.StatusConflict {
			return NewErrorWithCause(ErrorExitCode, err, "Could not create bucket with name %q.", bucketName)
		}

		if options.Get("indices").Exists() {
			indices := options.Get("indices").Array()
			indexResults, err := updateIndices(client, bucketName, indices)
			if err != nil {
				return err
			}
			result = indexResults
		} else {
			return NewErrorWithCause(ErrorExitCode, err, "Could not create bucket with name %q.", bucketName)
		}
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.IOStreams(), result)
}
