package cli

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
)

const (
	EqualsFilter = `{
	"equals": {
		"field": "%s",
		"value": "%s"
		}
	}`
)

// Getter defines the Get and GetAll methods.
type getter interface {
	Get(uuid.UUID) (gjson.Result, error)
	GetAll() ([]gjson.Result, error)
}

// GetEntity obtains the entity with the given ID using the given client and then prints out the result using the received printer or the JSON printer.
func (d discovery) GetEntity(client getter, id uuid.UUID, printer Printer) error {
	object, err := client.Get(id)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not get entity with id %q", id.String())
	}

	if printer == nil {
		jsonPrinter := JsonObjectPrinter(false)
		err = jsonPrinter(*d.IOStreams(), object)
	} else {
		err = printer(*d.IOStreams(), object)
	}

	return err
}

// GetEntities obtains all of the entities using the given client and then prints out the result using the received printer or the JSON array printer.
func (d discovery) GetEntities(client getter, printer Printer) error {
	objects, err := client.GetAll()
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not get all entities")
	}

	if printer == nil {
		arrayPrinter := JsonArrayPrinter(false)
		err = arrayPrinter(*d.IOStreams(), objects...)
	} else {
		err = printer(*d.IOStreams(), objects...)
	}

	return err
}

type searcher interface {
	getter
	Search(gjson.Result) ([]gjson.Result, error)
	SearchByName(name string) (gjson.Result, error)
}

func (d discovery) searchEntity(client searcher, id string) (gjson.Result, error) {
	result, err := client.SearchByName(id)
	if err != nil {
		discoveryErr, ok := err.(discoveryPackage.Error)
		if !ok {
			return gjson.Result{}, err
		}

		if discoveryErr.Status != http.StatusNotFound {
			return gjson.Result{}, discoveryErr
		}

		if parsedId, uuidErr := uuid.Parse(id); uuidErr == nil {
			result, err = client.Get(parsedId)
			if err != nil {
				return gjson.Result{}, err
			}

			return result, nil
		}

		return gjson.Result{}, discoveryErr
	}

	return result, nil
}

func (d discovery) SearchEntity(client searcher, id string, printer Printer) error {
	result, err := d.searchEntity(client, id)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not search for entity with id %q", id)
	}

	if printer == nil {
		jsonPrinter := JsonObjectPrinter(false)
		err = jsonPrinter(*d.IOStreams(), result)
	} else {
		err = printer(*d.IOStreams(), result)
	}

	return err
}

func (d discovery) SearchEntities(client searcher, filter gjson.Result, printer Printer) error {
	results, err := client.Search(filter)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not search for the entities")
	}

	if printer == nil {
		jsonPrinter := JsonArrayPrinter(false)
		err = jsonPrinter(*d.IOStreams(), results...)
	} else {
		err = printer(*d.IOStreams(), results...)
	}

	return err
}

func BuildEntitiesFilter(filters []string) (gjson.Result, error) {
	labelFilters := []string{}
	typeFilters := []string{}

	for _, filter := range filters {
		parts := strings.Split(filter, "=")
		filterType := parts[0]

		switch filterType {
		case "label":
			keyValue := strings.Split(parts[1], ":")
			key := keyValue[0]
			var value string
			if len(keyValue) > 1 {
				value = keyValue[1]
			}

			labelFilters = append(labelFilters, fmt.Sprintf(EqualsFilter, "labels.key", key))
			if value != "" {
				labelFilters = append(labelFilters, fmt.Sprintf(EqualsFilter, "labels.value", value))
			}
		case "type":
			typeFilters = append(typeFilters, fmt.Sprintf(EqualsFilter, "type", parts[1]))
		default:
			return gjson.Result{}, NewError(ErrorExitCode, "Filter %q does not exist", filterType)
		}
	}

	labelFilterString := "{}"
	var err error
	if len(labelFilters) > 1 {
		labelFilterString, err = sjson.SetRaw(labelFilterString, "and", "["+strings.Join(labelFilters, ",")+"]")
		if err != nil {
			return gjson.Result{}, NewErrorWithCause(ErrorExitCode, err, "Could not create label filters")
		}
	} else if len(labelFilters) == 1 {
		labelFilterString = labelFilters[0]
	}

	typeFilterString := "{}"
	if len(typeFilters) > 1 {
		typeFilterString, err = sjson.SetRaw(typeFilterString, "and", "["+strings.Join(typeFilters, ",")+"]")
		if err != nil {
			return gjson.Result{}, NewErrorWithCause(ErrorExitCode, err, "Could not create type filters")
		}
	} else if len(typeFilters) == 1 {
		typeFilterString = typeFilters[0]
	}

	filterString := "{}"
	switch {
	case len(labelFilters) > 0 && len(typeFilters) > 0:
		filterString, err = sjson.SetRaw(filterString, "and", fmt.Sprintf("[%s,%s]", labelFilterString, typeFilterString))
		if err != nil {
			return gjson.Result{}, NewErrorWithCause(ErrorExitCode, err, "Could not combine label and type filters")
		}
	case len(labelFilters) > 0:
		filterString = labelFilterString
	default:
		filterString = typeFilterString
	}

	return gjson.Parse(filterString), nil
}
