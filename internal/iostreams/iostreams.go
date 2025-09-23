package iostreams

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// IOStreams is the struct contains the streams to read from the standard input and write to the standard output and error.
type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	Err    io.Writer
	reader *bufio.Reader
}

func (ios *IOStreams) lineReader() *bufio.Reader {
	if ios.reader == nil {
		ios.reader = bufio.NewReader(ios.In)
	}
	return ios.reader
}

func (ios *IOStreams) AskUser(question string) (string, error) {
	reader := ios.lineReader()

	if _, err := fmt.Fprint(ios.Out, question); err != nil {
		return "", err
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return strings.TrimRight(line, "\r\n"), nil
		}
		return "", err
	}

	return strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r"), nil
}
