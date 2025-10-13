package cli

import (
	"errors"
	"fmt"
)

// ExitCode is a type definition that is used to define the possible execution exit codes.
type ExitCode int

// Error is a struct that wraps any error that can occur with the CLI and Discovery in order to improve their traceability.
type Error struct {
	ExitCode ExitCode
	Message  string
	Cause    error
}

// Error() defines how the error struct is formatted into a string.
// This is done to comply with the error interface.
func (e Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s\n%s", e.Message, e.Cause.Error())
	} else {
		return e.Message
	}
}

// The following constants represent possible exit codes of the CLI.
const (
	SuccessExitCode    ExitCode = 0 // This code is used when the CLI finished the command successfully.
	ErrorExitCode      ExitCode = 1 // This code is used when the CLI failed due to a normal error.
	PanicErrorExitCode ExitCode = 2 // This code is used when the CLI failed because it panicked somewhere in the code.
)

// NewErrorWithCause creates an Error with a cause. It receives the exit code, cause, message, and any arguments that can be added to the message in a formatted string.
func NewErrorWithCause(code ExitCode, cause error, message string, args ...any) Error {
	return Error{
		ExitCode: code,
		Message:  fmt.Sprintf(message, args...),
		Cause:    cause,
	}
}

// NewErrorWithCause creates an Error with a nil cause. It receives the exit code, message, and any arguments that can be added to the message in a formatted string.
func NewError(code ExitCode, message string, args ...any) Error {
	return Error{
		ExitCode: code,
		Message:  fmt.Sprintf(message, args...),
		Cause:    nil,
	}
}

// FromError transforms an error into the Error struct.
func FromError(err error) Error {
	var e Error
	if errors.As(err, &e) {
		return e
	}
	return Error{
		ExitCode: ErrorExitCode,
		Message:  "",
		Cause:    err,
	}
}
