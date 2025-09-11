package iostreams

import "io"

// IOStreams is the struct contains the streams to read from the standard input and write to the standard output and error.
type IOStreams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}
