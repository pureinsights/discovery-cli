package discovery

import (
	"fmt"

	"github.com/tidwall/gjson"
)

// Error represents an error response with an HTTP status and a JSON body.
type Error struct {
	Status int
	Body   gjson.Result
}

// Error implements the error interface.
func (e Error) Error() string {
	return fmt.Sprintf("status: %d, body: %s", e.Status, e.Body.String())
}
