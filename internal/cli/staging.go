package cli

import (
	"fmt"
	"net/http"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// StagingBucketController defines the methods to interact with buckets.
type StagingBucketController interface {
	Create(bucket string, options gjson.Result) (gjson.Result, error)
	Get(bucket string) (gjson.Result, error)
	CreateIndex(bucket, index string, config []gjson.Result) (gjson.Result, error)
	DeleteIndex(bucket, index string) (gjson.Result, error)
}

// updateIndices updates the indices in a bucket with the new configuration.
// It returns a JSON with an "indices" field that has the acknowledgements of the index updates.
func updateIndices(client StagingBucketController, bucketName string, oldIndices []gjson.Result, newIndices gjson.Result) (gjson.Result, error) {
	indexResults := "{}"
	for _, index := range newIndices.Array() {
		indexName := index.Get("name").String()

		indexAck, err := client.CreateIndex(bucketName, indexName, index.Get("fields").Array())
		if err != nil {
			indexResults, err = sjson.Set(indexResults, "indices."+indexName, err.Error())
		} else {
			indexResults, err = sjson.SetRaw(indexResults, "indices."+indexName, indexAck.Raw)
		}
		if err != nil {
			return gjson.Result{}, NewErrorWithCause(ErrorExitCode, err, "Could not update index with name %q of bucket %q.", indexName, bucketName)
		}
	}
	for _, index := range oldIndices {
		indexName := index.Get("name").String()

		oldIndexExists := newIndices.Get(fmt.Sprintf("#(name==%q)", indexName)).Exists()
		if !oldIndexExists {
			deleteIndex, err := client.DeleteIndex(bucketName, indexName)
			if err != nil {
				indexResults, err = sjson.Set(indexResults, "indices."+indexName, err.Error())
			} else {
				indexResults, err = sjson.SetRaw(indexResults, "indices."+indexName, deleteIndex.Raw)
			}
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
			bucketInfo, err := client.Get(bucketName)
			if err != nil {
				return NewErrorWithCause(ErrorExitCode, err, "Could not get bucket with name %q to update it.", bucketName)
			}
			oldIndices := bucketInfo.Get("indices").Array()
			newIndices := options.Get("indices")
			indexResults, err := updateIndices(client, bucketName, oldIndices, newIndices)
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
