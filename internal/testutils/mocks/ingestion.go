package mocks

import (
	"net/http"

	"github.com/google/uuid"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/tidwall/gjson"
)

// WorkingSeedController simulates a working IngestionSeedController.
type WorkingSeedController struct {
	WorkingSearcher
}

// Start returns the result of a new seed execution.
func (c *WorkingSeedController) Start(uuid.UUID, discoveryPackage.ScanType, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","creationTimestamp":"2025-09-04T19:29:41.119013Z","lastUpdatedTimestamp":"2025-09-04T19:29:41.119013Z","triggerType":"MANUAL","status":"CREATED","scanType":"INCREMENTAL","properties":{"stagingBucket":"testBucket"}}`), nil
}

// Halt returns the results of halting a seed.
func (c *WorkingSeedController) Halt(uuid.UUID) ([]gjson.Result, error) {
	return gjson.Parse(`[{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","status":202}, {"id":"365d3ce3-4ea6-47a8-ada5-4ab4bedcbb3b","status":202}]`).Array(), nil
}

// FailingSeedControllerGetEntityIdFails simulates a failing IngestionSeedController when GetEntityId fails.
type FailingSeedControllerGetEntityIdFails struct {
	SearcherIDNotUUID
}

// Start implements the interface.
func (c *FailingSeedControllerGetEntityIdFails) Start(uuid.UUID, discoveryPackage.ScanType, gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","creationTimestamp":"2025-09-04T19:29:41.119013Z","lastUpdatedTimestamp":"2025-09-04T19:29:41.119013Z","triggerType":"MANUAL","status":"CREATED","scanType":"INCREMENTAL","properties":{"stagingBucket":"testBucket"}}`), nil
}

// Halt implements the interface.
func (c *FailingSeedControllerGetEntityIdFails) Halt(uuid.UUID) ([]gjson.Result, error) {
	return gjson.Parse(`[{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","status":202}, {"id":"365d3ce3-4ea6-47a8-ada5-4ab4bedcbb3b","status":202}]`).Array(), nil
}

// FailingSeedControllerStartFails simulates when starting a seed execution fails.
type FailingSeedControllerStartFails struct {
	WorkingSearcher
}

// Start mocks a failing seed execution response.
func (c *FailingSeedControllerStartFails) Start(uuid.UUID, discoveryPackage.ScanType, gjson.Result) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
			"status": 409,
			"code": 4001,
			"messages": [
				"The seed has 1 executions: 0c309dbb-0402-4710-8659-2c75f5d649b6"
			],
			"timestamp": "2025-09-04T20:17:00.116546400Z"
			}`)}
}

// Halt implements the IngestionSeedController interface.
func (c *FailingSeedControllerStartFails) Halt(uuid.UUID) ([]gjson.Result, error) {
	return []gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Seed not found: 986ce864-af76-4fcb-8b4f-f4e4c6ab0951"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// WorkingSeedController simulates a working IngestionSeedController.
type WorkingSeedExecutionController struct {
	WorkingGetter
}

// Halt returns the results of halting a seed.
func (c *WorkingSeedExecutionController) Halt(uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{"acknowledged":true}`), nil
}

// FailingSeedControllerStartFails simulates when starting a seed execution fails.
type FailingSeedExecutionControllerHaltFails struct {
	WorkingGetter
}

// Halt returns the results of halting a seed.
func (c *FailingSeedExecutionControllerHaltFails) Halt(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
			"status": 409,
			"code": 4001,
			"messages": [
				"Action HALT cannot be applied to seed execution cc89b714-d00a-4774-9c45-9497b5d9f8ef because of its current status: HALTING"
			],
			"timestamp": "2025-09-03T21:05:21.861757200Z"
			}`)}
}

// WorkingGetter mocks the RecordGetter interface to always answer a working result.
type WorkingRecordGetter struct{}

// Get returns a record as if the request worked successfully.
func (g *WorkingRecordGetter) Get(string) (gjson.Result, error) {
	return gjson.Parse(`{
  "id": {
    "plain": "4e7c8a47efd829ef7f710d64da661786",
    "hash": "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0="
  },
  "creationTimestamp": "2025-09-03T21:02:54Z",
  "lastUpdatedTimestamp": "2025-09-03T21:02:54Z",
  "status": "SUCCESS"
}`), nil
}

// GetAll returns a list of records.
func (g *WorkingRecordGetter) GetAll() ([]gjson.Result, error) {
	return gjson.Parse(`[
		{"id":{"plain":"4e7c8a47efd829ef7f710d64da661786","hash":"A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"},
		{"id":{"plain":"8148e6a7b952a3b2964f706ced8c6885","hash":"IJeF-losyj33EAuqjgGW2G7sT-eE7poejQ5HokerZio="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"},
		{"id":{"plain":"b1e3e4f42c0818b1580e306eb776d4a1","hash":"N2lubqCWTqEEaymQVntpdP5dqKDP-LYk81C_PCr6btQ="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"}
	]`).Array(), nil
}

// FailingRecordGetter mocks the RecordGetter struct to always return an HTTP error.
type FailingRecordGetter struct{}

// Get returns a 404 Not Found.
func (g *FailingRecordGetter) Get(string) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
  "status": 404,
  "code": 1003,
  "messages": [
    "Entity not found: SeedRecordId(seed=Seed(super=AbstractComponentConfigEntity(super=AbstractJsonConfigEntity(super=AbstractTypedConfigEntity(super=AbstractConfigEntity(super=AbstractUpdatableEntity(super=AbstractCoreEntity(id=2acd0a61-852c-4f38-af2b-9c84e152873e), creationTimestamp=2025-08-21T21:52:03Z, lastUpdatedTimestamp=2025-08-21T21:52:03Z), name=Search seed, description=null, active=true), type=staging), config={\"action\":\"scroll\",\"bucket\":\"blogs\"})), properties=null, labels=[], recordOptions=SeedRecordPolicy[timeoutPolicy=TimeoutPolicy[slice=PT1H], errorPolicy=FATAL, outboundPolicy=OutboundPolicy[idPolicy=IdPolicy[generator=null], batchPolicy=BatchPolicy[maxCount=25, flushAfter=PT1M]]], hooks=[], beforeHooksOptions=null, afterHooksOptions=null), recordId=[3, 113, -45, 12, 72, 2, 107, -82, 65, 21, -101, 26, 115, -44, -56, -100, 88, -84, -66, 90, 17, -108, -67, -52, -25, 72, -93, 9, 99, 66, 43, 31])"
  ],
  "timestamp": "2025-11-09T14:42:48.411373100Z"
}`),
	}
}

// GetAll returns 401 unauthorized.
func (g *FailingRecordGetter) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// WorkingSeedExecutionGetter mocks a working seed execution getter.
type WorkingSeedExecutionGetter struct{}

// Get returns a seed execution.
func (g *WorkingSeedExecutionGetter) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
  "id": "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3",
  "creationTimestamp": "2025-10-10T19:48:31Z",
  "lastUpdatedTimestamp": "2025-10-10T19:48:31Z",
  "triggerType": "MANUAL",
  "status": "RUNNING",
  "scanType": "FULL",
  "properties": {
    "stagingBucket": "testBucket"
  },
  "stages": ["BEFORE_HOOKS","INGEST"]
}`), nil
}

// GetAll implements the interface.
func (g *WorkingSeedExecutionGetter) GetAll() ([]gjson.Result, error) {
	return []gjson.Result{}, nil
}

// Audit returns real audited changes.
func (g *WorkingSeedExecutionGetter) Audit(uuid.UUID) ([]gjson.Result, error) {
	return gjson.Parse(`[
	{"timestamp":"2025-09-05T20:09:22.543Z","status":"CREATED","stages":[]},
	{"timestamp":"2025-09-05T20:09:26.621Z","status":"RUNNING","stages":[]},
	{"timestamp":"2025-09-05T20:09:37.592Z","status":"RUNNING","stages":["BEFORE_HOOKS"]},
	{"timestamp":"2025-09-05T20:13:26.602Z","status":"RUNNING","stages":["BEFORE_HOOKS","INGEST"]}
]`).Array(), nil
}

// FailingSeedExecutionGetterGetExecutionFails mocks when getting a seed execution fails.
type FailingSeedExecutionGetterGetExecutionFails struct{}

// Get returns seed execution not found.
func (g *FailingSeedExecutionGetterGetExecutionFails) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "status": 404,
  "code": 1003,
  "messages": [
    "Seed execution not found: f85a5e19-8ed9-4f8c-9e2e-e1d5484612f2"
  ],
  "timestamp": "2025-11-17T19:32:01.555127800Z"
}`)}
}

// GetAll implements the interface.
func (g *FailingSeedExecutionGetterGetExecutionFails) GetAll() ([]gjson.Result, error) {
	return []gjson.Result{}, nil
}

// Audit implements the interface.
func (g *FailingSeedExecutionGetterGetExecutionFails) Audit(uuid.UUID) ([]gjson.Result, error) {
	return []gjson.Result{}, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// FFailingSeedExecutionGetterAuditFails mocks when getting the audit fails.
type FailingSeedExecutionGetterAuditFails struct{}

// Get returns a seed execution.
func (g *FailingSeedExecutionGetterAuditFails) Get(uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
  "id": "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3",
  "creationTimestamp": "2025-10-10T19:48:31Z",
  "lastUpdatedTimestamp": "2025-10-10T19:48:31Z",
  "triggerType": "MANUAL",
  "status": "RUNNING",
  "scanType": "FULL",
  "properties": {
    "stagingBucket": "testBucket"
  },
  "stages": ["BEFORE_HOOKS","INGEST"]
}`), nil
}

// GetAll implements the interface.
func (g *FailingSeedExecutionGetterAuditFails) GetAll() ([]gjson.Result, error) {
	return []gjson.Result{}, nil
}

// Audit returns an error.
func (g *FailingSeedExecutionGetterAuditFails) Audit(uuid.UUID) ([]gjson.Result, error) {
	return []gjson.Result{}, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// WorkingRecordSummarizer mocks when getting the record summary works.
type WorkingRecordSummarizer struct{}

// Summarize returns a real result.
func (s *WorkingRecordSummarizer) Summarize() (gjson.Result, error) {
	return gjson.Parse(`{"PROCESSING":4,"DONE": 4}`), nil
}

// NoContentRecordSummarizer mocks when the summarize does not return anything.
type NoContentRecordSummarizer struct{}

// NoContentRecordSummarizer returns an empty JSON.
func (s *NoContentRecordSummarizer) Summarize() (gjson.Result, error) {
	return gjson.Parse(``), nil
}

// WorkingJobSummarizer mocks when getting the job summary works.
type WorkingJobSummarizer struct{}

// Summarizer returns real results.
func (s *WorkingJobSummarizer) Summarize() (gjson.Result, error) {
	return gjson.Parse(`{"DONE":5,"RUNNING":3}`), nil
}

// FailingJobSummarizer mocks when getting the job summary fails.
type FailingJobSummarizer struct{}

// Summarize returns an error.
func (s *FailingJobSummarizer) Summarize() (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "status": 404,
  "code": 1003,
  "messages": [
    "Seed execution not found: f85a5e19-8ed9-4f8c-9e2e-e1d5484612f2"
  ],
  "timestamp": "2025-11-17T19:32:01.555127800Z"
}`)}
}
