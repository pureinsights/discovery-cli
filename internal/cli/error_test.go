package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestError_Error tests the Error.Error() function that outputs a string.
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
			expected: "Unable to connect to Discovery\nGet \"http://127.0.0.1:49190/down\": dial tcp 127.0.0.1:49190: connectex: No connection could be made because the target machine actively refused it.",
		},
		{
			name:     "Error string with cause",
			code:     ErrorExitCode,
			message:  "An unknown error occurred.",
			cause:    nil,
			expected: "An unknown error occurred.",
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

// TestNewErrorWithCause tests the NewErrorWithCause constructor function
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NewErrorWithCause(tc.code, tc.cause, tc.msg, tc.args...)

			assert.Equal(t, tc.expectedCode, got.ExitCode)
			assert.Equal(t, tc.expectedMsg, got.Message)

			if tc.expectedCause != nil {
				assert.EqualError(t, got.Cause, tc.expectedCause.Error())
			} else {
				assert.Nil(t, got.Cause)
			}
		})
	}
}

// TestNewError test the NewError() constructor function.
func TestNewError(t *testing.T) {
	tests := []struct {
		name         string
		code         ExitCode
		msg          string
		args         []any
		expectedCode ExitCode
		expectedMsg  string
	}{
		{
			name:         "Test with args",
			code:         ErrorExitCode,
			msg:          "Error %s",
			args:         []any{"connecting to Discovery"},
			expectedCode: ErrorExitCode,
			expectedMsg:  "Error connecting to Discovery",
		},
		{
			name:         "Test with no args",
			code:         PanicErrorExitCode,
			msg:          "Error connecting to Discovery",
			args:         []any(nil),
			expectedCode: PanicErrorExitCode,
			expectedMsg:  "Error connecting to Discovery",
		},
		{
			name:         "Test with empty message",
			code:         SuccessExitCode,
			msg:          "",
			args:         nil,
			expectedCode: SuccessExitCode,
			expectedMsg:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NewError(tc.code, tc.msg, tc.args...)

			assert.Equal(t, tc.expectedCode, got.ExitCode)
			assert.Equal(t, tc.expectedMsg, got.Message)
			assert.Nil(t, got.Cause)
		})
	}
}

// TestFromError tests the FromError() function that converts any error to the Error struct.
func TestFromError(t *testing.T) {
	tests := []struct {
		name          string
		input         error
		expectedCode  ExitCode
		expectedMsg   string
		expectedCause string
	}{
		{
			name: "Input is already Error",
			input: Error{
				ExitCode: ErrorExitCode,
				Message:  "An error occurred",
				Cause:    errors.New("Unable to connect to Discovery"),
			},
			expectedCode:  ErrorExitCode,
			expectedMsg:   "An error occurred",
			expectedCause: "Unable to connect to Discovery",
		},
		{
			name:          "Input is a generic error",
			input:         errors.New("JSON unmarshal failed"),
			expectedCode:  ErrorExitCode,
			expectedMsg:   "",
			expectedCause: "JSON unmarshal failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FromError(tc.input)

			assert.Equal(t, tc.expectedCode, got.ExitCode)
			assert.Equal(t, tc.expectedMsg, got.Message)

			if tc.expectedCause == "" {
				assert.Nil(t, got.Cause)
			} else {
				assert.NotNil(t, got.Cause)
				assert.EqualError(t, got.Cause, tc.expectedCause)
			}
		})
	}
}
