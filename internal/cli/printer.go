package cli

import (
	"encoding/json"
	"fmt"

	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/tidwall/gjson"
)

type Printer func(iostreams.IOStreams, ...gjson.Result) error

func printJsonObject(ios iostreams.IOStreams, pretty bool, object gjson.Result) error {
	var (
		formattedJson []byte
		err           error
	)
	if pretty {
		formattedJson, err = json.MarshalIndent([]byte(object.Raw), "", "  ")
	} else {
		formattedJson, err = json.Marshal([]byte(object.Raw))
	}
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(ios.Out, string(formattedJson)+"\n")
	return err
}

func printArrayObject(ios iostreams.IOStreams, pretty bool, array ...gjson.Result) error {
	if pretty {
		_, err := fmt.Fprint(ios.Out, "[")
		if err != nil {
			return err
		}
	}

	for _, object := range array {
		err := printJsonObject(ios, pretty, object)
		if err != nil {
			return err
		}
	}

	if pretty {
		_, err := fmt.Fprint(ios.Out, "]")
		if err != nil {
			return err
		}
	}

	return nil
}

func JsonObjectPrinter(pretty bool) Printer {
	return func(ios iostreams.IOStreams, objects ...gjson.Result) error {
		if len(objects) != 1 {
			return NewError(ErrorExitCode, "JsonObjectPrinter only works with a single JSON object")
		}
		err := printJsonObject(ios, pretty, objects[0])
		if err != nil {
			NewErrorWithCause(ErrorExitCode, err, "Could not print JSON object")
		}
		return nil
	}
}

func JsonArrayPrinter(pretty bool) Printer {
	return func(ios iostreams.IOStreams, objects ...gjson.Result) error {
		err := printArrayObject(ios, pretty, objects...)
		if err != nil {
			NewErrorWithCause(ErrorExitCode, err, "Could not print JSON Array")
		}
		return nil
	}
}

func GetObjectPrinter(name string) Printer {
	switch name {
	case "json":
		return JsonObjectPrinter(false)
	default:
		return nil
	}
}

func GetArrayPrinter(name string) Printer {
	switch name {
	case "json":
		return JsonArrayPrinter(false)
	default:
		return nil
	}
}
