package cli

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
)

const (
	// EqualsFilter contains the JSON string for the Equals DSL filter.
	EqualsFilter = `{
	"equals": {
		"field": "%s",
		"value": "%s"
		}
	}`
)

// Getter defines the Get and GetAll methods.
type Getter interface {
	Get(uuid.UUID) (gjson.Result, error)
	GetAll() ([]gjson.Result, error)
}

// GetEntity obtains the entity with the given ID using the given client and then prints out the result using the received printer or the JSON printer.
func (d discovery) GetEntity(client Getter, id uuid.UUID, printer Printer) error {
	object, err := client.Get(id)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not get entity with id %q", id.String())
	}

	if printer == nil {
		jsonPrinter := JsonObjectPrinter(true)
		err = jsonPrinter(*d.IOStreams(), object)
	} else {
		err = printer(*d.IOStreams(), object)
	}

	return err
}

// GetEntities obtains all the entities using the given client and then prints out the result using the received printer or the JSON array printer.
func (d discovery) GetEntities(client Getter, printer Printer) error {
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

// Searcher is the interface that implements searching methods.
type Searcher interface {
	Getter
	Search(gjson.Result) ([]gjson.Result, error)
	SearchByName(name string) (gjson.Result, error)
}

// searchEntity tries to search an entity by name, and if it fails, it tries to get the entity by its id.
func (d discovery) searchEntity(client Searcher, id string) (gjson.Result, error) {
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
			return client.Get(parsedId)
		}

		return gjson.Result{}, discoveryErr
	}

	return result, nil
}

// SearchEntity is an exported auxiliary function that calls the discovery.searchEntity() function.
func SearchEntity(d Discovery, client Searcher, id string) (gjson.Result, error) {
	return d.searchEntity(client, id)
}

// SearchEntity searches for the entity and prints it into the Out IOStream.
func (d discovery) SearchEntity(client Searcher, id string, printer Printer) error {
	result, err := d.searchEntity(client, id)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not search for entity with id %q", id)
	}

	if printer == nil {
		jsonPrinter := JsonObjectPrinter(true)
		err = jsonPrinter(*d.IOStreams(), result)
	} else {
		err = printer(*d.IOStreams(), result)
	}

	return err
}

// SearchEntities searches for entities and prints the results into the Out IOStream.
func (d discovery) SearchEntities(client Searcher, filter gjson.Result, printer Printer) error {
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

// parseFilter converts a filter in the format type=key:value to the JSON DSL Filter in Discovery.
func parseFilter(filter string) (string, []string, error) {
	filterType, keyValue, found := strings.Cut(filter, "=")
	if !found {
		return "", []string(nil), NewError(ErrorExitCode, "Filter %q does not follow the format {type}={key}[:{value}]", filter)
	}
	filters := []string{}
	switch filterType {
	case "label":
		key, value, found := strings.Cut(keyValue, ":")
		if key == "" {
			return "", []string(nil), NewError(ErrorExitCode, "The label's key in the filter %q cannot be empty", filter)
		}
		filters = append(filters, fmt.Sprintf(EqualsFilter, "labels.key", key))
		if found {
			if value == "" {
				return "", []string(nil), NewError(ErrorExitCode, "The label's value in the filter %q cannot be empty if ':' is included", filter)
			}
			filters = append(filters, fmt.Sprintf(EqualsFilter, "labels.value", value))
		}
	case "type":
		if keyValue == "" {
			return "", []string(nil), NewError(ErrorExitCode, "The value in the type filter %q cannot be empty", filter)
		}

		filters = append(filters, fmt.Sprintf(EqualsFilter, "type", keyValue))
	default:
		return "", []string(nil), NewError(ErrorExitCode, "Filter type %q does not exist", filterType)
	}

	return filterType, filters, nil
}

// getAndFilterString returns the filter string for the given filters.
// If there are multiple filters, they are joined with an "and" filter.
// If there is only one filter, it is returned.
// If there are no filters, an empty filter is returned.
func getAndFilterString(filters []string) (string, error) {
	if len(filters) > 1 {
		return sjson.SetRaw("{}", "and", "["+strings.Join(filters, ",")+"]")
	} else if len(filters) == 1 {
		return filters[0], nil
	}

	return "{}", nil
}

// BuildEntitiesFilter builds a filter based on the arguments sent to the get command.
// The filters are combined through the "and" operator.
func BuildEntitiesFilter(filters []string) (gjson.Result, error) {
	labelFilters := []string{}
	typeFilters := []string{}

	var err error
	for _, filter := range filters {
		filterType, parsedFilters, err := parseFilter(filter)
		if err != nil {
			return gjson.Result{}, err
		}
		switch filterType {
		case "label":
			labelFilters = append(labelFilters, parsedFilters...)
		case "type":
			typeFilters = append(typeFilters, parsedFilters...)
		}
	}

	labelFilterString, err := getAndFilterString(labelFilters)
	if err != nil {
		return gjson.Result{}, NewErrorWithCause(ErrorExitCode, err, "Could not create label filters")
	}

	typeFilterString, err := getAndFilterString(typeFilters)
	if err != nil {
		return gjson.Result{}, NewErrorWithCause(ErrorExitCode, err, "Could not create type filters")
	}

	filterString := "{}"
	switch {
	case len(labelFilters) > 0 && len(typeFilters) > 0:
		filterString, err = getAndFilterString([]string{labelFilterString, typeFilterString})
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

// Creator defines the methods to create and update entities.
type Creator interface {
	Create(config gjson.Result) (gjson.Result, error)
	Update(id uuid.UUID, config gjson.Result) (gjson.Result, error)
}

// UpsertEntity creates or updates an entity in Discovery with the given configuration.
func (d discovery) UpsertEntity(client Creator, config gjson.Result) (gjson.Result, error) {
	if config.Get("id").Exists() {
		if parsedId, uuidErr := uuid.Parse(config.Get("id").String()); uuidErr == nil {
			return client.Update(parsedId, config)
		} else {
			return gjson.Result{}, uuidErr
		}
	} else {
		return client.Create(config)
	}
}

// UpsertEntities creates or updates one or multiple entities based on the array of configurations it receives.
func (d discovery) UpsertEntities(client Creator, configurations gjson.Result, abortOnError bool, printer Printer) error {
	configArray := configurations.Array()
	upsertedEntites := []gjson.Result{}

	var upsertErr error
	for _, config := range configArray {
		upsert, err := d.UpsertEntity(client, config)
		if err != nil {
			if abortOnError {
				upsertErr = NewErrorWithCause(ErrorExitCode, err, "Could not store entities")
				break
			}

			var discoveryErr discoveryPackage.Error
			if errors.As(err, &discoveryErr) {
				upsertedEntites = append(upsertedEntites, discoveryErr.Body)
			} else {
				errJson := gjson.Parse(fmt.Sprintf("{\"error\":%q}", err.Error()))
				upsertedEntites = append(upsertedEntites, errJson)
			}
		} else {
			upsertedEntites = append(upsertedEntites, upsert)
		}
	}

	var err error
	if printer == nil {
		jsonPrinter := JsonArrayPrinter(false)
		err = jsonPrinter(*d.IOStreams(), upsertedEntites...)
	} else {
		err = printer(*d.IOStreams(), upsertedEntites...)
	}
	return errors.Join(err, upsertErr)
}

// Deleter is the interface that implements the delete method.
type Deleter interface {
	Delete(uuid.UUID) (gjson.Result, error)
}

// DeleteEntity deletes an entity with the received ID and prints the result using the given printer.
func (d discovery) DeleteEntity(client Deleter, id uuid.UUID, printer Printer) error {
	object, err := client.Delete(id)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not delete entity with id %q", id.String())
	}

	if printer == nil {
		jsonPrinter := JsonObjectPrinter(true)
		err = jsonPrinter(*d.IOStreams(), object)
	} else {
		err = printer(*d.IOStreams(), object)
	}

	return err
}

// SearchDeleter is the interface that implements the delete method that works with names.
type SearchDeleter interface {
	Deleter
	Searcher
}

// SearchDeleteEntity searches for the entity with the given name and then deletes it.
// It then prints out the results using the printer parameter.
func (d discovery) SearchDeleteEntity(client SearchDeleter, name string, printer Printer) error {
	result, err := d.searchEntity(client, name)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not search for entity with name %q", name)
	}

	deleteId, err := uuid.Parse(result.Get("id").String())
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not delete entity with name %q", name)
	}
	return d.DeleteEntity(client, deleteId, printer)
}
