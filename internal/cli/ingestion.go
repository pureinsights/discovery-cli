package cli

import (
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
)

// IngestionSeedController defines the methods to start and halt a seed.
type IngestionSeedController interface {
	Searcher
	Start(id uuid.UUID, scan discoveryPackage.ScanType, executionProperties gjson.Result) (gjson.Result, error)
	Halt(id uuid.UUID) ([]gjson.Result, error)
}

// GetSeedId obtains the UUID from the result of a search
func GetSeedId(d Discovery, client Searcher, name string) (uuid.UUID, error) {
	seed, err := d.searchEntity(client, name)
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(seed.Get("id").String())
}

// StartSeed initiates the execution of a seed with the given scanType and execution properties.
func (d discovery) StartSeed(client IngestionSeedController, name string, scanType discoveryPackage.ScanType, properties gjson.Result, printer Printer) error {
	seedId, err := GetSeedId(d, client, name)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not get seed ID to start execution.")
	}

	startResult, err := client.Start(seedId, scanType, properties)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not start seed execution for seed with id %q", seedId.String())
	}

	if printer == nil {
		jsonPrinter := JsonObjectPrinter(false)
		err = jsonPrinter(*d.IOStreams(), startResult)
	} else {
		err = printer(*d.IOStreams(), startResult)
	}

	return err
}

// HaltSeed stops all the seed executions of a seed
func (d discovery) HaltSeed(client IngestionSeedController, name string, printer Printer) error {
	seedId, err := GetSeedId(d, client, name)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not get seed ID to halt execution.")
	}

	haltResults, err := client.Halt(seedId)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not halt seed execution for seed with id %q", seedId.String())
	}

	if printer == nil {
		jsonPrinter := JsonArrayPrinter(false)
		err = jsonPrinter(*d.IOStreams(), haltResults...)
	} else {
		err = printer(*d.IOStreams(), haltResults...)
	}

	return err
}

// IngestionSeedExecutionController defines all of the methods to manage seed executions from commands
type IngestionSeedExecutionController interface {
	Getter
	Halt(id uuid.UUID) (gjson.Result, error)
}

// HaltSeedExecution stops a single seed execution with its UUID
func (d discovery) HaltSeedExecution(client IngestionSeedExecutionController, execution uuid.UUID, printer Printer) error {
	haltResult, err := client.Halt(execution)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not halt the seed execution with id %q", execution.String())
	}

	if printer == nil {
		jsonPrinter := JsonObjectPrinter(false)
		err = jsonPrinter(*d.IOStreams(), haltResult)
	} else {
		err = printer(*d.IOStreams(), haltResult)
	}

	return err
}

type RecordGetter interface {
	Get(id string) (gjson.Result, error)
	GetAll() ([]gjson.Result, error)
}

func AppendSeedRecords(seed gjson.Result, client RecordGetter) (gjson.Result, error) {
	records, err := client.GetAll()
	if err != nil {
		return gjson.Result{}, err
	}

	raw, err := sjson.Set(seed.Raw, "records", records)
	if err != nil {
		return gjson.Result{}, err
	}

	return gjson.Parse(raw), err 
}


func AppendSeedRecord(seed gjson.Result, client RecordGetter, id string) (gjson.Result, error) {
	records, err := client.Get(id)
	if err != nil {
		return gjson.Result{}, err
	}

	raw, err := sjson.Set(seed.Raw, "record", records)
	if err != nil {
		return gjson.Result{}, err
	}

	return gjson.Parse(raw), err 
}

func (d discovery) AppendSeedRecords(seed gjson.Result, client RecordGetter, printer Printer) error {
	seed, err := AppendSeedRecords(seed, client)
	if err != nil {
		return err
	}

	if printer != nil {
		printer = JsonObjectPrinter(false)
	}

	return printer(*d.iostreams, seed)
}

func (d discovery) AppendSeedRecord(seed gjson.Result, client RecordGetter, id string, printer Printer) error {
	seed, err := AppendSeedRecord(seed, client, id)
	if err != nil {
		return err
	}

	if printer != nil {
		printer = JsonObjectPrinter(false)
	}

	return printer(*d.iostreams, seed)
}
