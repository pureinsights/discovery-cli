package discovery

import (
	"net/http"

	"github.com/tidwall/gjson"
)

type searcher struct {
	client
}

func (c cloner) Search(filter gjson.Result, params map[string][]string) ([]gjson.Result, error) {
	response, err := execute(c.client, http.MethodPost, "/search", WithQueryParameters(params))
	if err == nil {
		elementNumber := response.Get("numberOfElements").Int()
		pageNumber := response.Get("pageNumber").Int()
		totalPages := response.Get("totalPages").Int()
		totalSize := response.Get("totalSize").Int()
		elements := response.Get("content").Array()
		for pageNumber < totalPages || elementNumber < totalSize {
			response, err = execute(c.client, http.MethodPost, "/search", WithQueryParameters(params))
			if err == nil {
				pageElementNumber := response.Get("numberOfElements").Int()
				pagePageNumber := response.Get("pageNumber").Int()
				pageElements := response.Get("content").Array()
			} else {

				return []gjson.Result{}, err
			}
		}
		return response, nil
	} else {

		return []gjson.Result{}, err
	}
}

func Autocomplete(q string) ([]gjson.Result, error) {

}
