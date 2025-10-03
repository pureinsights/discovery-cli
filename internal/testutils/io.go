package testutils

// ErrWriter is used to force an error when writing to the output stream.
type ErrWriter struct{ Err error }

// Write completes the implementation of the io.Writer interface.
func (w ErrWriter) Write(p []byte) (int, error) { return 0, w.Err }

// ErrReader is used to force an error when reading from the input stream.
type ErrReader struct{ Err error }

// Read completes the implementation of the io.Reader interface.
func (r ErrReader) Read(p []byte) (int, error) { return 0, r.Err }
