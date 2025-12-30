package discovery

import (
	"net/http"

	"github.com/tidwall/gjson"
)

// statusChecker is a struct that has methods to check the status or health of the Discovery products.
type statusChecker struct {
	client
}

// StatusCheck calls the health endpoint of a Discovery product. If it is online, the response should be a JSON with a "status" field whose value is "UP".
func (c statusChecker) StatusCheck() (gjson.Result, error) {
	return execute(c.client, http.MethodGet, "/health")
}
