package discovery

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

type seedExecutionRecordsClient struct {
	summarizer
}

func newSeedExecutionRecordsClient(c seedExecutionsClient, executionId uuid.UUID) seedExecutionRecordsClient {
	return seedExecutionRecordsClient{
		summarizer: summarizer{
			client: newSubClient(c.client, "/"+executionId.String()+"/record"),
		},
	}
}

type seedExecutionJobsClient struct {
	summarizer
}

func newSeedExecutionJobsClient(c seedExecutionsClient, executionId uuid.UUID) seedExecutionJobsClient {
	return seedExecutionJobsClient{
		summarizer: summarizer{
			client: newSubClient(c.client, "/"+executionId.String()+"/job"),
		},
	}
}

type seedRecordsClient struct {
	getter
	summarizer
}

func newSeedRecordsClient(url, apiKey string, seedId uuid.UUID) seedRecordsClient {
	client := newClient(url+"/seed/"+seedId.String()+"/record", apiKey)
	return seedRecordsClient{
		summarizer: summarizer{
			client: client,
		},
		getter: getter{
			client: client,
		},
	}
}

type seedExecutionsClient struct {
	getter
}

func newSeedExecutionsClient(url, apiKey string, seedId uuid.UUID) seedExecutionsClient {
	return seedExecutionsClient{
		getter: getter{
			client: newClient(url+"/seed/"+seedId.String()+"/execution", apiKey),
		},
	}
}

func (c seedExecutionsClient) Halt(executionId uuid.UUID) (gjson.Result, error) {
	return execute(c.client, http.MethodPost, "/"+executionId.String()+"/halt")
}

func (c seedExecutionsClient) Audit(executionId uuid.UUID) ([]gjson.Result, error) {
	response, err := execute(c.client, http.MethodPost, "/"+executionId.String()+"/audit")
	if err != nil {
		return []gjson.Result(nil), err
	}

	elementNumber := response.Get("numberOfElements").Int()
	pageNumber := response.Get("pageNumber").Int()
	totalPages := response.Get("totalPages").Int()
	totalSize := response.Get("totalSize").Int()
	elements := response.Get("content").Array()
	pageNumber++
	for pageNumber < totalPages && elementNumber < totalSize {
		response, err = execute(c.client, http.MethodGet, "", WithQueryParameters(map[string][]string{"page": {strconv.FormatInt(pageNumber, 10)}}))
		if err != nil {
			return []gjson.Result(nil), err
		}

		pageElements := response.Get("content").Array()
		elements = append(elements, pageElements...)

		pageNumber++
		pageElementNumber := response.Get("numberOfElements").Int()
		elementNumber += pageElementNumber
	}
	return elements, nil
}

func (c seedExecutionsClient) Seed(executionId uuid.UUID) (gjson.Result, error) {
	return execute(c.client, http.MethodPost, "/"+executionId.String()+"/config/seed")
}

func (c seedExecutionsClient) Pipeline(executionId uuid.UUID, pipelineId uuid.UUID) (gjson.Result, error) {
	return execute(c.client, http.MethodPost, "/"+executionId.String()+"/config/pipeline/"+pipelineId.String())
}

func (c seedExecutionsClient) Processor(executionId uuid.UUID, processorId uuid.UUID) (gjson.Result, error) {
	return execute(c.client, http.MethodPost, "/"+executionId.String()+"/config/processor/"+processorId.String())
}

func (c seedExecutionsClient) Server(executionId uuid.UUID, serverId uuid.UUID) (gjson.Result, error) {
	return execute(c.client, http.MethodPost, "/"+executionId.String()+"/config/server/"+serverId.String())
}

func (c seedExecutionsClient) Credential(executionId uuid.UUID, credentialId uuid.UUID) (gjson.Result, error) {
	return execute(c.client, http.MethodPost, "/"+executionId.String()+"/config/credential/"+credentialId.String())
}

func (c seedExecutionsClient) Records(executionId uuid.UUID) seedExecutionRecordsClient {
	return newSeedExecutionRecordsClient(c, executionId)
}

func (c seedExecutionsClient) Jobs(executionId uuid.UUID) seedExecutionJobsClient {
	return newSeedExecutionJobsClient(c, executionId)
}
