package discovery

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

// enabler is a struct that adds the path to enable or disable Discovery's entities.
type enabler struct {
	client
}

// Enable runs a request to enable the entity with the given ID. The client's URL needs to already have the path that points to the correct type of entity.
func (e enabler) Enable(id uuid.UUID) (gjson.Result, error) {
	return execute(e.client, http.MethodPatch, "/"+id.String()+"/enable")
}

// Disable runs a request to disable the entity with the given ID. The client's URL needs to already have the path that points to the correct type of entity.
func (e enabler) Disable(id uuid.UUID) (gjson.Result, error) {
	return execute(e.client, http.MethodPatch, "/"+id.String()+"/disable")
}
