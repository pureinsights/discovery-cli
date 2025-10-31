package cli

import (
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
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
		jsonPrinter := JsonObjectPrinter(false)
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
