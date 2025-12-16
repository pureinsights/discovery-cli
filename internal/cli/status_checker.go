package cli

import (
	"github.com/tidwall/gjson"
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
		return NewErrorWithCause(ErrorExitCode, err, "Discovery %s is not online.", product)
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.iostreams, result)
}
