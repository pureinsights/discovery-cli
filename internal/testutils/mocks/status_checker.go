package mocks

import (
	"errors"

	"github.com/stretchr/testify/mock"
	"github.com/tidwall/gjson"
)

// WorkingStatusChecker mocks the results of a StatusChecker that does a request to an online product.
type WorkingStatusChecker struct {
	mock.Mock
}

// StatusCheck returns the response of an online Discovery product.
func (g *WorkingStatusChecker) StatusCheck() (gjson.Result, error) {
	return gjson.Parse(`{
    "status": "UP"
}`), nil
}

// WorkingStatusChecker mocks the results of a StatusChecker that does a request to an offline product.
type FailingStatusChecker struct {
	mock.Mock
}

// StatusCheck returns the error of an offline Discovery product.
func (g *FailingStatusChecker) StatusCheck() (gjson.Result, error) {
	return gjson.Result{}, errors.New("Get \"http://localhost:12030/health\": dial tcp [::1]:12030: connectex: No connection could be made because the target machine actively refused it.")
}
