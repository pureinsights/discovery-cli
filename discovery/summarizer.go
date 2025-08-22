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

	recordSummarySeed, err := execute(s.client, http.MethodGet, "/"+seedId.String()+"/execution/"+executionId.String()+"/job/summary")	if err != nil {
		return gjson.Result{}, err
	}

	jobSummaryExecution, err := execute(s.client, http.MethodGet, "/"+seedId.String()+"/execution/"+executionId.String()+"/job/summary")
	if err != nil {
		return gjson.Result{}, err
	}

	recordSummaryExecution, err := execute(s.client, http.MethodGet, "/"+seedId.String()+"/execution/"+executionId.String()+"/job/summary")
	if err != nil {
		return gjson.Result{}, err
	}


	return response, nil
}
