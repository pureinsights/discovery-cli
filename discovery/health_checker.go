package discovery

import (
	"net/http"

	"github.com/tidwall/gjson"
)

// cloner is a struct that has methods to clone entities.
type healthChecker struct {
	client
}

// Clone makes a copy of the entity with the given ID.
func (c healthChecker) HealthCheck() (gjson.Result, error) {
	return execute(c.client, http.MethodGet, "/health")
}
