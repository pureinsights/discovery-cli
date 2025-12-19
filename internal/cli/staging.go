package cli

import (
	"net/http"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/tidwall/gjson"
)

// StagingBucketController defines the methods to interact with buckets.
type StagingBucketController interface {
	Create(bucket string, options gjson.Result) (gjson.Result, error)
	Get(bucket string) (gjson.Result, error)
	Delete(bucket string) (gjson.Result, error)
}

// StoreBucket creates or updates a bucket if it exists. It receives an options parameter that contains the configuration of the bucket.
func (d discovery) StoreBucket(client StagingBucketController, name string, options gjson.Result, printer Printer) error {
	result, err := client.Create(name, options)
	if err != nil {
		discoveryErr, ok := err.(discoveryPackage.Error)
		if !ok {
			return NewErrorWithCause(ErrorExitCode, err, "Could not create bucket with name %q.", name)
		}

		if discoveryErr.Status != http.StatusConflict {
			return NewErrorWithCause(ErrorExitCode, err, "Could not create bucket with name %q.", name)
		}

		deleteResults, err := client.Delete(name)
		if err != nil || !(deleteResults.Get("acknowledged").Bool()) {
			return NewErrorWithCause(ErrorExitCode, err, "Could not delete and update bucket with name %q.", name)
		}

		result, err = client.Create(name, options)
		if err != nil {
			return NewErrorWithCause(ErrorExitCode, err, "Could not update bucket with name %q", name)
		}
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.IOStreams(), result)
}
