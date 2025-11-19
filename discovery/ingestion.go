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
	summarizer
}

// NewSeedRecordsClient is the constructor of seedRecordsClient
func newSeedRecordsClient(sc seedsClient, seedId uuid.UUID) seedRecordsClient {
	client := newSubClient(sc.crud.client, "/"+seedId.String()+"/record")
	return seedRecordsClient{
		summarizer: summarizer{
			client: client,
		},
	}
}

// Get obtains a record based on the seed and record IDs.
// Since record IDs are not UUIDs, a new function was needed.
func (src seedRecordsClient) Get(id string) (gjson.Result, error) {
	return execute(src.client, http.MethodGet, "/"+id)
}

// GetAll obtains every record in the seed.
func (src seedRecordsClient) GetAll() ([]gjson.Result, error) {
	return executeWithPagination(src.client, http.MethodGet, "")
}

// SeedExecutionClient can carry out every operation regarding seed executions.
// With its Getter embedded struct, it can obtain seed executions.
type seedExecutionsClient struct {
	getter
}

// NewSeedExecutionsClient is the constructor of a seedExecutionClient.
func newSeedExecutionsClient(sc seedsClient, seedId uuid.UUID) seedExecutionsClient {
	return seedExecutionsClient{
		getter: getter{
			client: newSubClient(sc.crud.client, "/"+seedId.String()+"/execution"),
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
	return executeWithPagination(c.client, http.MethodGet, "/"+executionId.String()+"/audit")
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

// IngestionProcessorsClient is the struct performs the CRUD and cloning of processors.
type ingestionProcessorsClient struct {
	crud
	cloner
	searcher
}

// NewIngestionProcessorsClient is the constructor of a ingestionProcessorsClient
func newIngestionProcessorsClient(url, apiKey string) ingestionProcessorsClient {
	client := newClient(url+"/processor", apiKey)
	return ingestionProcessorsClient{
		crud: crud{
			getter{
				client: client,
			},
		},
		cloner: cloner{
			client: client,
		},
		searcher: searcher{
			client: client,
		},
	}
}

// PipelinesClient is the struct that performs the CRUD and cloning of pipelines.
type pipelinesClient struct {
	crud
	cloner
	searcher
}

// NewPipelinesClient is the constructor of a pipelinesClient
func newPipelinesClient(url, apiKey string) pipelinesClient {
	client := newClient(url+"/pipeline", apiKey)
	return pipelinesClient{
		crud: crud{
			getter{
				client: client,
			},
		},
		cloner: cloner{
			client: client,
		},
		searcher: searcher{
			client: client,
		},
	}
}

// SeedsClient is the struct that performs the CRUD and cloning of seeds.
type seedsClient struct {
	crud
	cloner
	searcher
}

// NewSeedsClient is the constructor of seedsClient.
func newSeedsClient(url, apiKey string) seedsClient {
	client := newClient(url+"/seed", apiKey)
	return seedsClient{
		crud: crud{
			getter{
				client: client,
			},
		},
		cloner: cloner{
			client: client,
		},
		searcher: searcher{
			client: client,
		},
	}
}

// ScanType is used as an enum to easily represent the scanTypes for seeds.
type ScanType string

// The constants represent the respective scan type.
const (
	ScanFull        ScanType = "FULL"
	ScanIncremental ScanType = "INCREMENETAL"
)

// Start starts the execution of seed.
func (sc seedsClient) Start(id uuid.UUID, scan ScanType, executionProperties gjson.Result) (gjson.Result, error) {
	if !executionProperties.Exists() {
		return execute(sc.crud.client, http.MethodPost, "/"+id.String(), WithQueryParameters(map[string][]string{
			"scanType": {string(scan)},
		}))
	} else {
		return execute(sc.crud.client, http.MethodPost, "/"+id.String(), WithQueryParameters(map[string][]string{
			"scanType": {string(scan)},
		}), WithJSONBody(executionProperties.Raw))
	}
}

// Halt stops all the executions of a seed.
func (sc seedsClient) Halt(id uuid.UUID) ([]gjson.Result, error) {
	haltings, err := execute(sc.crud.client, http.MethodPost, "/"+id.String()+"/halt")
	if err != nil {
		return []gjson.Result(nil), err
	}

	return haltings.Array(), err
}

// Reset resets a seed.
// If the seed has no active executions, then the seed's metadata is reset and its records deleted.
func (sc seedsClient) Reset(id uuid.UUID) (gjson.Result, error) {
	return execute(sc.crud.client, http.MethodPost, "/"+id.String()+"/reset")
}

// Records creates a new seedRecordsClient.
func (sc seedsClient) Records(seedId uuid.UUID) seedRecordsClient {
	return newSeedRecordsClient(sc, seedId)
}

// Executions creates a new seedExecutionsClient.
func (sc seedsClient) Executions(seedId uuid.UUID) seedExecutionsClient {
	return newSeedExecutionsClient(sc, seedId)
}

// Ingestion is the struct that is used to interact with the Ingestion Component
type ingestion struct {
	Url, ApiKey string
}

// Procesors is used to create an ingestionProcessorsClient
func (i ingestion) Processors() ingestionProcessorsClient {
	return newIngestionProcessorsClient(i.Url, i.ApiKey)
}

// Pipelines is used to create a pipelinesClient
func (i ingestion) Pipelines() pipelinesClient {
	return newPipelinesClient(i.Url, i.ApiKey)
}

// Seeds is used to create a seedsClient
func (i ingestion) Seeds() seedsClient {
	return newSeedsClient(i.Url, i.ApiKey)
}

// BackupRestore creates a backUpRestore struct.
func (i ingestion) BackupRestore() backupRestore {
	return backupRestore{
		client: newClient(i.Url, i.ApiKey),
	}
}

// NewIngestion is the constructor of the ingestion struct.
// It adds a /v2 path to the URL in order to properly connect to Discovery.
func NewIngestion(url, apiKey string) ingestion {
	return ingestion{Url: url + "/v2", ApiKey: apiKey}
}
