package cli

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

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

// AppendSeedRecords adds a "records" field to the seed, which contains all of the records obtained from the seed
func AppendSeedRecords(seed gjson.Result, client RecordGetter) (gjson.Result, error) {
	records, err := client.GetAll()
	if err != nil {
		return gjson.Result{}, err
	}

	recordsString := "[\n"
	for index, record := range records {
		recordsString = recordsString + record.Raw
		if index != (len(records) - 1) {
			recordsString = recordsString + ","
		}
		recordsString = recordsString + "\n"
	}
	seedWithRecord, err := sjson.SetRaw(seed.Raw, "records", recordsString+"]")
	return gjson.Parse(seedWithRecord), err
}

// AppendSeedRecords obtains all of the seed's records, appends them to the seed, and prints out the seed.
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
