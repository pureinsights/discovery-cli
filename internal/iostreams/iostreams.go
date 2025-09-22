package iostreams

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// IOStreams is the struct contains the streams to read from the standard input and write to the standard output and error.
type IOStreams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

func (ios IOStreams) AskUser(question, defaultValue string) (string, error) {
	reader := bufio.NewReader(ios.In)

	prompt := fmt.Sprintf("%s [%q]: ", question, defaultValue)
	if _, err := fmt.Fprint(ios.Out, prompt); err != nil {
		return "", err
	}

	line, err := reader.ReadString('\n')
	if err != nil {

		return "", err
	}

	return strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r"), nil
}
