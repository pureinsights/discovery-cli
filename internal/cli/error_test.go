package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		code     ExitCode
		message  string
		cause    error
		expected string
	}{
		{
			name:     "Error string with cause",
			code:     ErrorExitCode,
			message:  "Unable to connect to Discovery",
			cause:    errors.New(`Get "http://127.0.0.1:49190/down": dial tcp 127.0.0.1:49190: connectex: No connection could be made because the target machine actively refused it.`),
			expected: "Message: Unable to connect to Discovery\nCause: Get \"http://127.0.0.1:49190/down\": dial tcp 127.0.0.1:49190: connectex: No connection could be made because the target machine actively refused it.",
		},
		{
			name:     "Error string with cause",
			code:     ErrorExitCode,
			message:  "An unknown error occurred.",
			cause:    nil,
			expected: "Message: An unknown error occurred.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := Error{ExitCode: tc.code, Message: tc.message, Cause: tc.cause}
			assert.EqualError(t, e, tc.expected)
			assert.NotContains(t, e.Error(), tc.code)
		})
	}
}

func TestNewErrorWithCause(t *testing.T) {
	tests := []struct {
		name          string
		code          ExitCode
		cause         error
		msg           string
		args          []any
		expectedCode  ExitCode
		expectedMsg   string
		expectedCause error
	}{
		{
			name:          "Test with args",
			code:          ErrorExitCode,
			cause:         errors.New(`Get "http://127.0.0.1:49190/down": dial tcp 127.0.0.1:49190: connectex: No connection could be made because the target machine actively refused it.`),
			msg:           "Error %s",
			args:          []any{"connecting to Discovery"},
			expectedCode:  ErrorExitCode,
			expectedMsg:   "Error connecting to Discovery",
			expectedCause: errors.New(`Get "http://127.0.0.1:49190/down": dial tcp 127.0.0.1:49190: connectex: No connection could be made because the target machine actively refused it.`),
		},
		{
			name:          "Test with no args",
			code:          PanicErrorExitCode,
			cause:         errors.New(`Get "http://127.0.0.1:49190/down": dial tcp 127.0.0.1:49190: connectex: No connection could be made because the target machine actively refused it.`),
			msg:           "Error connecting to Discovery",
			args:          []any(nil),
			expectedCode:  PanicErrorExitCode,
			expectedMsg:   "Error connecting to Discovery",
			expectedCause: errors.New(`Get "http://127.0.0.1:49190/down": dial tcp 127.0.0.1:49190: connectex: No connection could be made because the target machine actively refused it.`),
		},
		{
			name:          "Test with empty message and cause",
			code:          SuccessExitCode,
			cause:         errors.New(`Get "http://127.0.0.1:49190/down": dial tcp 127.0.0.1:49190: connectex: No connection could be made because the target machine actively refused it.`),
			msg:           "",
			args:          nil,
			expectedCode:  SuccessExitCode,
			expectedMsg:   "",
			expectedCause: errors.New(`Get "http://127.0.0.1:49190/down": dial tcp 127.0.0.1:49190: connectex: No connection could be made because the target machine actively refused it.`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewErrorWithCause(tt.code, tt.cause, tt.msg, tt.args...)

			assert.Equal(t, tt.expectedCode, got.ExitCode)
			assert.Equal(t, tt.expectedMsg, got.Message)

			if tt.expectedCause != nil {
				assert.EqualError(t, got.Cause, tt.expectedCause.Error())
			} else {
				assert.Nil(t, got.Cause)
			}
		})
	}
}
