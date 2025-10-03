package testutils

import (
	"fmt"
	"io"
)

// ErrWriter is used to force an error when writing to the output stream.
type ErrWriter struct{ Err error }

// Write completes the implementation of the io.Writer interface.
func (w ErrWriter) Write(p []byte) (int, error) { return 0, w.Err }

// ErrReader is used to force an error when reading from the input stream.
type ErrReader struct{ Err error }

// Read completes the implementation of the io.Reader interface.
func (r ErrReader) Read(p []byte) (int, error) { return 0, r.Err }

// FailOnNWriter is a struct that mocks the IOStreams.Out field.
// It is used to force errors when writing to an output stream.
type FailOnNWriter struct {
	Writer  io.Writer
	N       int
	counter int
}

// Write implements the io.Writer interface. This function fails on the N'th writing operation.
func (f *FailOnNWriter) Write(p []byte) (int, error) {
	f.counter++
	if f.counter == f.N {
		return 0, fmt.Errorf("write failed")
	}
	return f.Writer.Write(p)
}
