package cli

import (
	"errors"
	"io/fs"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			expected: "Unable to connect to Discovery\nGet \"http://127.0.0.1:49190/down\": dial tcp 127.0.0.1:49190: connectex: No connection could be made because the target machine actively refused it.\n",
		},
		{
			name:     "Error string with cause",
			code:     ErrorExitCode,
			message:  "An unknown error occurred.",
			cause:    nil,
			expected: "An unknown error occurred.\n",
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
		{
			name:          "FromError receives nil",
			input:         nil,
			expectedCode:  ErrorExitCode,
			expectedMsg:   "",
			expectedCause: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FromError(tc.input)

			if tc.input != nil {
				assert.Equal(t, tc.expectedCode, got.ExitCode)
				assert.Equal(t, tc.expectedMsg, got.Message)
				if tc.expectedCause == "" {
					assert.Nil(t, got.Cause)
				} else {
					assert.NotNil(t, got.Cause)
					assert.EqualError(t, got.Cause, tc.expectedCause)
				}
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

// TestNormalizeReadFileError tests the NormalizeReadFileError
func TestNormalizeReadFileError(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		err         error
		expectedErr error
	}{
		{
			name:        "file does not exist (direct)",
			path:        "entities.json",
			err:         fs.ErrNotExist,
			expectedErr: errors.New("file does not exist: entities.json"),
		},
		{
			name: "file does not exist wrapped in PathError",
			path: "entities.json",
			err: &fs.PathError{
				Op:   "open",
				Path: "entities.json",
				Err:  fs.ErrNotExist,
			},
			expectedErr: errors.New("file does not exist: entities.json"),
		},
		{
			name:        "permission denied",
			path:        "secrets.json",
			err:         fs.ErrPermission,
			expectedErr: errors.New("permission denied while reading file: secrets.json"),
		},
		{
			name:        "invalid file path",
			path:        "asdfjk;l///kasdfkasdf",
			err:         fs.ErrInvalid,
			expectedErr: errors.New("invalid file path: asdfjk;l///kasdfkasdf"),
		},
		{
			name:        "file closed unexpectedly",
			path:        "entities.json",
			err:         fs.ErrClosed,
			expectedErr: errors.New("file was closed unexpectedly: entities.json"),
		},
		{
			name:        "file already exists",
			path:        "config.toml",
			err:         fs.ErrExist,
			expectedErr: errors.New("file already exists: config.toml"),
		},
		{
			name:        "path is a directory",
			path:        "/dir",
			err:         syscall.EISDIR,
			expectedErr: errors.New("path is a directory, not a file: /dir"),
		},
		{
			name:        "too many open files",
			path:        "entities.json",
			err:         syscall.EMFILE,
			expectedErr: errors.New("too many open files while reading: entities.json"),
		},
		{
			name:        "out of memory",
			path:        "large-file.txt",
			err:         syscall.ENOMEM,
			expectedErr: errors.New("out of memory while reading file: large-file.txt"),
		},
		{
			name:        "invalid argument",
			path:        "invalid",
			err:         syscall.EINVAL,
			expectedErr: errors.New("invalid argument while reading file: invalid"),
		},
		{
			name:        "low-level IO error",
			path:        "io.txt",
			err:         syscall.EIO,
			expectedErr: errors.New("low-level I/O error while reading file: io.txt"),
		},
		{
			name:        "nil error",
			path:        "directory",
			err:         nil,
			expectedErr: nil,
		},
		{
			name:        "not a file error error",
			path:        "directory",
			err:         errors.New("this is not a file error"),
			expectedErr: errors.New("this is not a file error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NormalizeReadFileError(tc.path, tc.err)
			if tc.expectedErr != nil {
				require.Error(t, actual)
				assert.EqualError(t, actual, tc.expectedErr.Error())
			} else {
				assert.Nil(t, actual)
			}
		})
	}
}

// TestNormalizeWriteFileError tests the NormalizeWriteFileError() function
func TestNormalizeWriteFileError(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		err         error
		expectedErr error
	}{
		{
			name:        "file does not exist (direct)",
			path:        "doesnotexist/entities.json",
			err:         fs.ErrNotExist,
			expectedErr: errors.New("the given path does not exist: doesnotexist/entities.json"),
		},
		{
			name: "file does not exist wrapped in PathError",
			path: "entities.json",
			err: &fs.PathError{
				Op:   "open",
				Path: "entities.json",
				Err:  fs.ErrNotExist,
			},
			expectedErr: errors.New("the given path does not exist: entities.json"),
		},
		{
			name:        "permission denied",
			path:        "secrets.json",
			err:         fs.ErrPermission,
			expectedErr: errors.New("permission denied while writing file: secrets.json"),
		},
		{
			name:        "write to directory",
			path:        "/directory",
			err:         syscall.EISDIR,
			expectedErr: errors.New("cannot write to a directory: /directory"),
		},
		{
			name:        "file closed unexpectedly",
			path:        "entities.json",
			err:         syscall.ENOSPC,
			expectedErr: errors.New("no space left on device while writing file: entities.json"),
		},
		{
			name:        "file system is read only exists",
			path:        "config.toml",
			err:         syscall.EROFS,
			expectedErr: errors.New("filesystem is read-only: config.toml"),
		},
		{
			name:        "too many open files",
			path:        "entities.json",
			err:         syscall.EMFILE,
			expectedErr: errors.New("too many open files while writing file: entities.json"),
		},
		{
			name:        "low-level IO error",
			path:        "io.txt",
			err:         syscall.EIO,
			expectedErr: errors.New("low-level I/O error while writing file: io.txt"),
		},
		{
			name:        "nil error",
			path:        "directory",
			err:         nil,
			expectedErr: nil,
		},
		{
			name:        "not a file error error",
			path:        "directory",
			err:         errors.New("this is not a file error"),
			expectedErr: errors.New("this is not a file error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NormalizeWriteFileError(tc.path, tc.err)
			if tc.expectedErr != nil {
				require.Error(t, actual)
				assert.EqualError(t, actual, tc.expectedErr.Error())
			} else {
				assert.Nil(t, actual)
			}
		})
	}
}
