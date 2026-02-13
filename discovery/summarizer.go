package discovery

import (
	"net/http"

	"github.com/tidwall/gjson"
)

// summarizer is a struct that has the path to call the ingestion summary endpoints.
type summarizer struct {
	client
}

// Summarize adds /summary to the client's base URL and executes the GET method to get a summary of the entity.
func (s summarizer) Summarize() (gjson.Result, error) {
	return execute(s.client, http.MethodGet, "/summary")
}
