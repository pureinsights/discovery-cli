package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"syscall"
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
		return fmt.Sprintf("%s\n%s\n", e.Message, e.Cause.Error())
	} else {
		return e.Message + "\n"
	}
}

// The following constants represent possible exit codes of the CLI.
const (
	// This code is used when the CLI finished the command successfully.
	SuccessExitCode ExitCode = 0
	// This code is used when the CLI failed due to a normal error.
	ErrorExitCode ExitCode = 1
	// This code is used when the CLI failed because it panicked somewhere in the code.
	PanicErrorExitCode ExitCode = 2
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
func FromError(err error) *Error {
	if err != nil {
		var e Error
		if errors.As(err, &e) {
			return &e
		}
		return &Error{
			ExitCode: ErrorExitCode,
			Message:  "",
			Cause:    err,
		}
	}

	return nil
}

// NormalizeReadFileError makes the read file errors OS agnostic.
func NormalizeReadFileError(path string, err error) error {
	if err == nil {
		return nil
	}

	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		err = pathErr.Err
	}

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("file does not exist: %s", path)

	case errors.Is(err, fs.ErrPermission):
		return fmt.Errorf("permission denied while reading file: %s", path)

	case errors.Is(err, fs.ErrInvalid):
		return fmt.Errorf("invalid file path: %s", path)

	case errors.Is(err, fs.ErrClosed):
		return fmt.Errorf("file was closed unexpectedly: %s", path)

	case errors.Is(err, fs.ErrExist):
		return fmt.Errorf("file already exists: %s", path)

	case errors.Is(err, syscall.EISDIR):
		return fmt.Errorf("path is a directory, not a file: %s", path)

	case errors.Is(err, syscall.EMFILE):
		return fmt.Errorf("too many open files while reading: %s", path)

	case errors.Is(err, syscall.ENOMEM):
		return fmt.Errorf("out of memory while reading file: %s", path)

	case errors.Is(err, syscall.EINVAL):
		return fmt.Errorf("invalid argument while reading file: %s", path)

	case errors.Is(err, syscall.EIO):
		return fmt.Errorf("low-level I/O error while reading file: %s", path)

	default:
		return err
	}
}

// NormalizeWriteFileError makes the write file errors OS agnostic.
func NormalizeWriteFileError(path string, err error) error {
	if err == nil {
		return nil
	}

	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		err = pathErr.Err
	}

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("the given path does not exist: %s", path)

	case errors.Is(err, fs.ErrPermission):
		return fmt.Errorf("permission denied while writing file: %s", path)

	case errors.Is(err, syscall.EISDIR):
		return fmt.Errorf("cannot write to a directory: %s", path)

	case errors.Is(err, syscall.ENOSPC):
		return fmt.Errorf("no space left on device while writing file: %s", path)

	case errors.Is(err, syscall.EROFS):
		return fmt.Errorf("filesystem is read-only: %s", path)

	case errors.Is(err, syscall.EMFILE):
		return fmt.Errorf("too many open files while writing file: %s", path)

	case errors.Is(err, syscall.EIO):
		return fmt.Errorf("low-level I/O error while writing file: %s", path)

	default:
		return err
	}
}
