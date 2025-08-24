package discovery

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

type enabler struct {
	client
}

func (e enabler) Enable(id uuid.UUID) (gjson.Result, error) {
	enabled, err := execute(e.client, http.MethodPatch, "/"+id.String()+"/enable")
	if err != nil {
		return gjson.Result{}, err
	}

	return enabled, nil
}

func (e enabler) Disable(id uuid.UUID) (gjson.Result, error) {
	disabled, err := execute(e.client, http.MethodPatch, "/"+id.String()+"/disable")
	if err != nil {
		return gjson.Result{}, err
	}

	return disabled, nil
}
