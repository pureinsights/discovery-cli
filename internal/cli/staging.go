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

		indexResults := "{}"
		indices := options.Get("indices").Array()
		for _, index := range indices {
			indexName := index.Get("name").String()
			indexConfig := []gjson.Result{}
			for _, field := range index.Get("fields").Array() {
				fieldJson, err := sjson.Set("{}", field.Get("key").String(), field.Get("value").String())
				if err != nil {
					return NewErrorWithCause(ErrorExitCode, err, "Could not update index with name %q of bucket %q.", indexName, bucketName)
				}
				indexConfig = append(indexConfig, gjson.Parse(fieldJson))
			}
			indexAck, err := client.CreateIndex(bucketName, indexName, indexConfig)
			if err != nil {
				return NewErrorWithCause(ErrorExitCode, err, "Could not update index with name %q of bucket %q.", indexName, bucketName)
			}
			indexResults, err = sjson.SetRaw(indexResults, indexName, indexAck.Raw)
			if err != nil {
				return NewErrorWithCause(ErrorExitCode, err, "Could not update index with name %q of bucket %q.", indexName, bucketName)
			}
		}

		result = gjson.Parse(indexResults)
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.IOStreams(), result)
}
