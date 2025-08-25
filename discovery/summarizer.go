package discovery

import (
	"net/http"

	"github.com/tidwall/gjson"
)

// Summarizer is a struct that has the path to call the ingestion summary endpoints.
type summarizer struct {
	client
}

// Summarize adds /summary to the client's base URL and executes the GET method.
// This function works on clients that already have the seed id execution ids set in their URLs.
func (s summarizer) Summarize() (gjson.Result, error) {

	summary, err := execute(s.client, http.MethodGet, "/summary")
	if err != nil {
		return gjson.Result{}, err
	}

	return summary, nil
}
