package cli

import (
	"errors"
	"net/http"
	"net/url"

	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// StatusChecker defines the StatusCheck() method.
type StatusChecker interface {
	StatusCheck() (gjson.Result, error)
}

// statusFromError converts a failed status check into a readable status that keeps the original error message.
// An empty or invalid URL or a 4xx response is reported as an invalid configuration, and a 5xx response or any other failure is reported as down.
func statusFromError(err error) string {
	status := "DOWN"
	var discoveryErr discoveryPackage.Error
	var urlErr *url.Error
	switch {
	case errors.As(err, &discoveryErr):
		if discoveryErr.Status < http.StatusInternalServerError {
			status = "INVALID CONFIGURATION"
		}
	case errors.As(err, &urlErr):
		if u, parseErr := url.Parse(urlErr.URL); parseErr != nil || (u.Scheme != "http" && u.Scheme != "https") {
			status = "INVALID CONFIGURATION"
		}
	}

	result, _ := sjson.Set(`{}`, "status", status)
	result, _ = sjson.Set(result, "error", err.Error())
	return result
}

// StatusCheck gets the Status of the sent product and prints out the result.
// If there was an error, the error message indicates the name of the product that is unavailable.
func (d discovery) StatusCheck(client StatusChecker, product string, printer Printer) error {
	result, checkErr := client.StatusCheck()
	if checkErr != nil {
		result = gjson.Parse(statusFromError(checkErr))
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	if err := printer(*d.iostreams, result); err != nil {
		return err
	}

	if checkErr != nil {
		return NewError(ErrorExitCode, "Could not check the status of Discovery %s.", product)
	}
	return nil
}

// StatusCheckClientEntry is used to easily store the different status checker structs of the Discovery products.
type StatusCheckClientEntry struct {
	Name   string
	Client StatusChecker
}

// StatusCheckOfClients checks the status of very Discovery product and returns their results.
func (d *discovery) StatusCheckOfClients(clients []StatusCheckClientEntry, printer Printer) error {
	results := "{}"
	for _, entry := range clients {
		client := entry.Client
		status, err := client.StatusCheck()
		raw := status.Raw
		if err != nil {
			d.markFailed()
			raw = statusFromError(err)
		}

		results, err = sjson.SetRaw(results, entry.Name, raw)
		if err != nil {
			return NewErrorWithCause(ErrorExitCode, err, "Could not get the status of the Discovery products")
		}
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.iostreams, gjson.Parse(results))
}
