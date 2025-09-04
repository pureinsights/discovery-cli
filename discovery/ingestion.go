package discovery

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

// SeedExecutionRecordsClient is the struct that can get the summary of records from a seed execution.
type seedExecutionRecordsClient struct {
	summarizer
}

// NewSeedExecutionRecordsClient is the constructor of seedExecutionRecordsClient.
func newSeedExecutionRecordsClient(c seedExecutionsClient, executionId uuid.UUID) seedExecutionRecordsClient {
	return seedExecutionRecordsClient{
		summarizer: summarizer{
			client: newSubClient(c.client, "/"+executionId.String()+"/record"),
		},
	}
}

// SeedExecutionJobsClient is the struct that can get the summary of jobs from a seed execution.
type seedExecutionJobsClient struct {
	summarizer
}

// NewSeedExecutionJobsClient is the constructor of seedExecutionJobsClient.
func newSeedExecutionJobsClient(c seedExecutionsClient, executionId uuid.UUID) seedExecutionJobsClient {
	return seedExecutionJobsClient{
		summarizer: summarizer{
			client: newSubClient(c.client, "/"+executionId.String()+"/job"),
		},
	}
}

// SeedRecordsClient is the struct that can get records and the summary of records from a seed.
type seedRecordsClient struct {
	getter
	summarizer
}

// NewSeedRecordsClient is the constructor of seedRecordsClient
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

// Get obtains a record based on the seed and record IDs.
// Since record IDs are not UUIDs, a new function was needed.
func (src seedRecordsClient) Get(id string) (gjson.Result, error) {
	return execute(src.getter.client, http.MethodGet, "/"+id)
}

// SeedExecutionClient can carry out every operation regarding seed executions.
// With its Getter embedded struct, it can obtain seed executions.
type seedExecutionsClient struct {
	getter
}

// NewSeedExecutionsClient is the constructor of seedExecutionClient.
func newSeedExecutionsClient(url, apiKey string, seedId uuid.UUID) seedExecutionsClient {
	return seedExecutionsClient{
		getter: getter{
			client: newClient(url+"/seed/"+seedId.String()+"/execution", apiKey),
		},
	}
}

// Halt stops a seed execution based on the seedId and executionId.
// It cannot halt an execution if it is already in a state that does not allow it.
func (c seedExecutionsClient) Halt(executionId uuid.UUID) (gjson.Result, error) {
	return execute(c.client, http.MethodPost, "/"+executionId.String()+"/halt")
}

// Audit gets the audited changes from a seed execution. It returns an array with the stages the execution has completed.
func (c seedExecutionsClient) Audit(executionId uuid.UUID) ([]gjson.Result, error) {
	auxClient := seedExecutionsClient{
		getter: getter{
			client: newSubClient(c.client, "/"+executionId.String()+"/audit"),
		},
	}
	return auxClient.GetAll()
}

// Seed gets the seed configuration of the seed execution.
func (c seedExecutionsClient) Seed(executionId uuid.UUID) (gjson.Result, error) {
	return execute(c.client, http.MethodGet, "/"+executionId.String()+"/config/seed")
}

// Pipeline gets the pipeline's configuration
func (c seedExecutionsClient) Pipeline(executionId uuid.UUID, pipelineId uuid.UUID) (gjson.Result, error) {
	return execute(c.client, http.MethodGet, "/"+executionId.String()+"/config/pipeline/"+pipelineId.String())
}

// Processor gets the configuration of a processor used in the pipeline of the seed.
func (c seedExecutionsClient) Processor(executionId uuid.UUID, processorId uuid.UUID) (gjson.Result, error) {
	return execute(c.client, http.MethodGet, "/"+executionId.String()+"/config/processor/"+processorId.String())
}

// Server gets the configuration of a server used by a processor of the seed.
func (c seedExecutionsClient) Server(executionId uuid.UUID, serverId uuid.UUID) (gjson.Result, error) {
	return execute(c.client, http.MethodGet, "/"+executionId.String()+"/config/server/"+serverId.String())
}

// Credential gets the configuration of a credential used by a server of the seed.
func (c seedExecutionsClient) Credential(executionId uuid.UUID, credentialId uuid.UUID) (gjson.Result, error) {
	return execute(c.client, http.MethodGet, "/"+executionId.String()+"/config/credential/"+credentialId.String())
}

// Records creates a seedExecutionsRecordClient.
func (c seedExecutionsClient) Records(executionId uuid.UUID) seedExecutionRecordsClient {
	return newSeedExecutionRecordsClient(c, executionId)
}

// Jobs creates a seedExecutionJobsClient.
func (c seedExecutionsClient) Jobs(executionId uuid.UUID) seedExecutionJobsClient {
	return newSeedExecutionJobsClient(c, executionId)
}
