package cli

import (
	"errors"
	"fmt"
)

type ExitCode int

type Error struct {
	ExitCode ExitCode
	Message  string
	Cause    error
}

func (e Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("Message: %s\nCause: %s", e.Message, e.Cause.Error())
	} else {
		return fmt.Sprintf("Message: %s", e.Message)
	}
}

const (
	SuccessExitCode    ExitCode = 0
	ErrorExitCode      ExitCode = 1
	PanicErrorExitCode ExitCode = 2
)

func NewErrorWithCause(code ExitCode, cause error, message string, args ...any) Error {
	return Error{
		ExitCode: code,
		Message:  fmt.Sprintf(message, args...),
		Cause:    cause,
	}
}

func NewError(code ExitCode, message string, args ...any) Error {
	return Error{
		ExitCode: code,
		Message:  fmt.Sprintf(message, args...),
		Cause:    nil,
	}
}

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
