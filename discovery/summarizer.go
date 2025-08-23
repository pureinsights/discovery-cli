package discovery

import (
	"net/http"

	"github.com/tidwall/gjson"
)

type summarizer struct {
	client
}

func (s summarizer) Summarize() (gjson.Result, error) {

	summary, err := execute(s.client, http.MethodGet, "/summary")
	if err != nil {
		return gjson.Result{}, err
	}

	return summary, nil
}
