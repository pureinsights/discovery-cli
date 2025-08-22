package discovery

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

type summarizer struct {
	client
}

func (s summarizer) Summarize(seedId, executionId uuid.UUID) (gjson.Result, error) {

	summary, err := execute(s.client, http.MethodGet, "/"+seedId.String()+"/summary")
	if err != nil {
		return gjson.Result{}, err
	}

	return summary, nil
}
