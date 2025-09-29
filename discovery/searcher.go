package discovery

import (
	"net/http"

	"github.com/tidwall/gjson"
)

type searcher struct {
	client
}

func (s searcher) Search(filter gjson.Result) ([]gjson.Result, error) {
	return executeWithPagination(s.client, http.MethodPost, "/search", WithJSONBody(filter.Raw))
}
