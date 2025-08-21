package discovery

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

// Cloner is a struct that has methods to clone entities.
type cloner struct {
	client
}

// Clone makes a copy of an entity.
// It returns the body of the new, duplicated entity or an error if the request failed.
func (c cloner) Clone(id uuid.UUID, params map[string][]string) (gjson.Result, error) {
	response, err := execute(c.client, http.MethodPost, "/"+id.String()+"/clone", WithQueryParameters(params))
	if err != nil {
		return gjson.Result{}, err
	} else {
		return response, nil
	}
}
