package discovery

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

// Enabler is a struct that adds the path to enable or disable Queryflow endpoints and Ingestion Seed Schedules.
type enabler struct {
	client
}

// Enable runs a request to enable an entity. The client's URL needs to already have the path that points to the correct type of entity.
func (e enabler) Enable(id uuid.UUID) (gjson.Result, error) {
	return execute(e.client, http.MethodPatch, "/"+id.String()+"/enable")
}

// Disable runs a request to disable an entity. The client's URL needs to already have the path that points to the correct type of entity.
func (e enabler) Disable(id uuid.UUID) (gjson.Result, error) {
	return execute(e.client, http.MethodPatch, "/"+id.String()+"/disable")
}
