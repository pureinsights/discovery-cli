package discovery

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

type cloner struct {
	client
}

func (c cloner) Clone(id uuid.UUID, params map[string][]string) (gjson.Result, error) {
	response, err := execute(c.client, http.MethodPost, "/"+id.String()+"/clone", WithQueryParameters(params))
	if err != nil {
		return gjson.Result{}, err
	} else {
		return response, nil
	}
}
