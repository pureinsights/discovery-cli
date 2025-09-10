package cli

import "fmt"

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

func NewErrorWithCause(code int, cause error, message string, args ...any) Error

func NewError(code int, message string, args ...any) Error
