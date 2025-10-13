package cli

import (
	"encoding/json"
	"fmt"

	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/tidwall/gjson"
)

type Printer func(iostreams.IOStreams, ...gjson.Result) error

// PrintJsonObject prints the given JSON object to the Out IOStream.
// If the pretty boolean is true, it prints the object with spacing and indentation
// If not, it prints it in a compact format.
func printJsonObject(ios iostreams.IOStreams, pretty bool, object gjson.Result) error {
	var v any
	if err := json.Unmarshal([]byte(object.Raw), &v); err != nil {
		return err
	}

	var b []byte
	var err error
	if pretty {
		b, err = json.MarshalIndent(v, "", "  ")
	} else {
		b, err = json.Marshal(v)
	}
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(ios.Out, string(b))
	return err
}

// PrintJsonArray prints the given JSON array to the Out IOStream.
// If the pretty boolean is true, it prints the objects in the array with spacing and indentation and adds brackets at each end.
// If not, it prints the objects in a compact format.
func printArrayObject(ios iostreams.IOStreams, pretty bool, array ...gjson.Result) error {
	if pretty {
		_, err := fmt.Fprint(ios.Out, "[\n")
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
		_, err := fmt.Fprint(ios.Out, "]\n")
		if err != nil {
			return err
		}
	}

	return nil
}

// JsonObjectPrinter receives the pretty boolean and returns the function that prints the JSON object with that boolean as a parameter.
func JsonObjectPrinter(pretty bool) Printer {
	return func(ios iostreams.IOStreams, objects ...gjson.Result) error {
		if len(objects) != 1 {
			return NewError(ErrorExitCode, "JsonObjectPrinter only works with a single JSON object")
		}
		err := printJsonObject(ios, pretty, objects[0])
		if err != nil {
			return NewErrorWithCause(ErrorExitCode, err, "Could not print JSON object")
		}
		return nil
	}
}

// JsonArrayPrinter receives the pretty boolean and returns the function that prints the JSON array with that boolean as a parameter.
func JsonArrayPrinter(pretty bool) Printer {
	return func(ios iostreams.IOStreams, objects ...gjson.Result) error {
		err := printArrayObject(ios, pretty, objects...)
		if err != nil {
			return NewErrorWithCause(ErrorExitCode, err, "Could not print JSON Array")
		}
		return nil
	}
}

// GetObjectPrinter chooses the most appropiate printer depending on the given printer name.
func GetObjectPrinter(name string) Printer {
	switch name {
	case "json":
		return JsonObjectPrinter(false)
	default:
		return nil
	}
}

// GetArrayPrinter chooses the most appropiate printer depending on the given printer name.
func GetArrayPrinter(name string) Printer {
	switch name {
	case "json":
		return JsonArrayPrinter(false)
	default:
		return nil
	}
}
