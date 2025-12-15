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

// lineReader is used to set up the IOStream reader field.
// This field is used to be able to use the same reader when reading an input string that has the configuration for all of the Discovery Components.
// For example: "http://discovery.core.cn\n\nhttp://discovery.ingestion.cn\n\nhttp://discovery.queryflow.cn\n\n\n\n".
// The previous example changes the Core URL, Ingestion URL, and QueryFlow URL.
// If a new reader was created every call of the AskUser() function, the buffered lines would be lost and the string will not be read completely.
// This feature is mostly used to simplify the unit tests.
func (ios *IOStreams) lineReader() *bufio.Reader {
	if ios.reader == nil {
		ios.reader = bufio.NewReader(ios.In)
	}
	return ios.reader
}

// AskUser asks the given question to the user and reads for their response.
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
