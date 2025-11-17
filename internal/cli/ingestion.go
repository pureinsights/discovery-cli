package cli

import (
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// ConvertJSONArrayToString transforms a []gjson.Result into a valid JSON array string
func ConvertJSONArrayToString(array []gjson.Result) string {
	arrayString := "[\n"
	for index, record := range array {
		arrayString = arrayString + record.Raw
		if index != (len(array) - 1) {
			arrayString = arrayString + ","
		}
		arrayString = arrayString + "\n"
	}
	return arrayString + "]"
}

// RecordGetter defines the methods to get seed records.
type RecordGetter interface {
	Get(id string) (gjson.Result, error)
	GetAll() ([]gjson.Result, error)
}

// AppendSeedRecord adds a "record" field to the seed, which contains the record obtained using the given id.
func AppendSeedRecord(seed gjson.Result, client RecordGetter, id string) (gjson.Result, error) {
	record, err := client.Get(id)
	if err != nil {
		return gjson.Result{}, err
	}

	seedWithRecord, err := sjson.SetRaw(seed.Raw, "record", record.Raw)
	return gjson.Parse(seedWithRecord), err
}

// AppendSeedRecord obtains a seed record, appends it to the seed, and prints out the seed.
func (d discovery) AppendSeedRecord(seed gjson.Result, client RecordGetter, id string, printer Printer) error {
	seedWithRecord, err := AppendSeedRecord(seed, client, id)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not get record with id %q", id)
	}

	if printer == nil {
		jsonPrinter := JsonObjectPrinter(false)
		err = jsonPrinter(*d.IOStreams(), seedWithRecord)
	} else {
		err = printer(*d.IOStreams(), seedWithRecord)
	}

	return err
}

// AppendSeedRecord adds a "record" field to the seed, which contains the record obtained using the given id.
func AppendSeedRecords(seed gjson.Result, client RecordGetter) (gjson.Result, error) {
	records, err := client.GetAll()
	if err != nil {
		return gjson.Result{}, err
	}

	recordsString := ConvertJSONArrayToString(records)
	seedWithRecord, err := sjson.SetRaw(seed.Raw, "records", recordsString)
	return gjson.Parse(seedWithRecord), err
}

// AppendSeedRecord obtains a seed record, appends it to the seed, and prints out the seed.
func (d discovery) AppendSeedRecords(seed gjson.Result, client RecordGetter, printer Printer) error {
	seedWithRecords, err := AppendSeedRecords(seed, client)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not get records")
	}

	if printer == nil {
		jsonPrinter := JsonObjectPrinter(false)
		err = jsonPrinter(*d.IOStreams(), seedWithRecords)
	} else {
		err = printer(*d.IOStreams(), seedWithRecords)
	}

	return err
}

// Summarizer defines the Summarize() method.
type Summarizer interface {
	Summarize() (gjson.Result, error)
}

// SeedExecutionGetter defines the methods to get seed executions and the audited changes.
type SeedExecutionGetter interface {
	Getter
	Audit(executionId uuid.UUID) ([]gjson.Result, error)
}

// AppendSeedExecutionDetails appends details, like the audited changes and summaries to a seed execution JSON.
func AppendSeedExecutionDetails(seedExecution gjson.Result, seedExecutionId uuid.UUID, client SeedExecutionGetter, summarizers map[string]Summarizer) (gjson.Result, error) {
	auditLogs, err := client.Audit(seedExecutionId)
	if err != nil {
		return gjson.Result{}, err
	}

	auditString := ConvertJSONArrayToString(auditLogs)
	raw, err := sjson.SetRaw(seedExecution.Raw, "audit", auditString)
	if err != nil {
		return gjson.Result{}, err
	}

	for field, summarizer := range summarizers {
		summary, err := summarizer.Summarize()
		if err != nil {
			return gjson.Result{}, err
		}

		sumString := "{}"
		if summary.Exists() {
			sumString = summary.Raw
		}
		raw, err = sjson.SetRaw(raw, field, sumString)
		if err != nil {
			return gjson.Result{}, err
		}
	}

	return gjson.Parse(raw), nil
}

// GetSeedExecution gets a seed execution, appends details if needed, and prints out the result.
func (d discovery) GetSeedExecution(client SeedExecutionGetter, seedExecutionId uuid.UUID, summarizers map[string]Summarizer, details bool, printer Printer) error {
	execution, err := client.Get(seedExecutionId)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not get seed execution with id %q", seedExecutionId.String())
	}

	if details {
		execution, err = AppendSeedExecutionDetails(execution, seedExecutionId, client, summarizers)
		if err != nil {
			return NewErrorWithCause(ErrorExitCode, err, "Could not get details for seed execution with id %q", seedExecutionId.String())
		}
	}

	if printer == nil {
		jsonPrinter := JsonObjectPrinter(true)
		err = jsonPrinter(*d.IOStreams(), execution)
	} else {
		err = printer(*d.IOStreams(), execution)
	}
	return err
}
