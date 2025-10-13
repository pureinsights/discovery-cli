package cli

import (
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

type getter interface {
	Get(uuid.UUID) (gjson.Result, error)
	GetAll() ([]gjson.Result, error)
}

func (d discovery) GetEntity(client getter, id uuid.UUID, printer Printer) error {
	object, err := client.Get(id)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not get entity with id %q", id.String())
	}

}
