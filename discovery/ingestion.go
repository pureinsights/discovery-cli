package discovery

import (
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

type seedExecutionRecordsClient struct {
	summarizer
}

func newSeedExecutionRecordsClient(url, apiKey string, seedId, executionId uuid.UUID) seedExecutionRecordsClient {
	return seedExecutionRecordsClient{
		summarizer: summarizer{
			client: newClient(url+"/seed/"+seedId.String()+"/execution/"+executionId.String()+"/record", apiKey),
		},
	}
}

type seedExecutionJobsClient struct {
	summarizer
}

func newSeedExecutionJobsClient(url, apiKey string, seedId, executionId uuid.UUID) seedExecutionJobsClient {
	return seedExecutionJobsClient{
		summarizer: summarizer{
			client: newClient(url+"/seed/"+seedId.String()+"/execution/"+executionId.String()+"/job", apiKey),
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

func (c seedExecutionsClient) Halt(executionId uuid.UUID) (gjson.Result, error)
func (c seedExecutionsClient) Audit(executionId uuid.UUID) ([]gjson.Result, error)
func (c seedExecutionsClient) Seed(executionId uuid.UUID) (gjson.Result, error)
func (c seedExecutionsClient) Pipeline(executionId uuid.UUID, pipelineId uuid.UUID) (gjson.Result, error)
func (c seedExecutionsClient) Processor(executionId uuid.UUID, processorId uuid.UUID) (gjson.Result, error)
func (c seedExecutionsClient) Server(executionId uuid.UUID, serverId uuid.UUID) (gjson.Result, error)
func (c seedExecutionsClient) Credential(executionId uuid.UUID, credentialId uuid.UUID) (gjson.Result, error)
func (c seedExecutionsClient) Records(executionId uuid.UUID) seedExecutionRecordsClient
func (c seedExecutionsClient) Jobs(executionId uuid.UUID) seedExecutionJobsClient
