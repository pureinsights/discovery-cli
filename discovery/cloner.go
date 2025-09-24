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

// Clone makes a copy of the entity with the given ID.
// It returns the body of the new, duplicated entity or an error if the request failed.
// It receives the UUID of the entity's identifier and the query parameters that can be used as arguments for the different clone endpoints in Discovery's entities.
// For example, those parameters can be the name of the cloned entity, the URI of a cloned QueryFlow endpoint, or the depth of the copy.
func (c cloner) Clone(id uuid.UUID, params map[string][]string) (gjson.Result, error) {
	return execute(c.client, http.MethodPost, "/"+id.String()+"/clone", WithQueryParameters(params))
}
