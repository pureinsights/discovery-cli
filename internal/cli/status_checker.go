package cli

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// StatusChecker defines the StatusCheck() method.
type StatusChecker interface {
	StatusCheck() (gjson.Result, error)
}

// StatusCheck gets the Status of the sent product and prints out the result.
// If there was an error, the error message indicates the name of the product that is unavailable.
func (d discovery) StatusCheck(client StatusChecker, product string, printer Printer) error {
	result, err := client.StatusCheck()
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not check the status of Discovery %s.", product)
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.iostreams, result)
}

// StatusCheckClientEntry is used to easily store the different status checker structs of the Discovery products.
type StatusCheckClientEntry struct {
	Name   string
	Client StatusChecker
}

// StatusCheckOfClients checks the status of very Discovery product and returns their results.
func (d discovery) StatusCheckOfClients(clients []StatusCheckClientEntry, printer Printer) error {
	results := "{}"
	for _, entry := range clients {
		client := entry.Client
		status, err := client.StatusCheck()
		if err == nil {
			results, err = sjson.SetRaw(results, entry.Name, status.Raw)
			if err != nil {
				return NewErrorWithCause(ErrorExitCode, err, "Could not get the status of the Discovery products")
			}
		} else {
			results, err = sjson.Set(results, entry.Name, err.Error())
			if err != nil {
				return NewErrorWithCause(ErrorExitCode, err, "Could not get the status of the Discovery products")
			}
		}
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.iostreams, gjson.Parse(results))
}
