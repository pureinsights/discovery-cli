package iostreams

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test_lineReader_InitializesWhenNil tests the lineReader() function when there is no reader yet.
func Test_lineReader_InitializesWhenNil(t *testing.T) {
	ios := &IOStreams{In: strings.NewReader("testreader\n")}
	r := ios.lineReader()

	require.NotNil(t, r)
	require.Same(t, r, ios.reader)

	line, err := r.ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "testreader\n", line)
}

// Test_lineReader_CachesSameInstance tests the lineReader() function when there is already reader, so it should not create a new one.
func Test_lineReader_CachesSameInstance(t *testing.T) {
	ios := &IOStreams{In: strings.NewReader("line1\nline2\n")}
	r1 := ios.lineReader()
	r2 := ios.lineReader()

	require.Same(t, r1, r2, "lineReader must return the same cached instance")
}

// ErrWriter is used to force an error when writing to the output stream.
type errWriter struct{ err error }

// Write completes the implementation of the io.Writer interface.
func (w errWriter) Write(p []byte) (int, error) { return 0, w.err }

// ErrReader is used to force an error when reading from the input stream.
type errReader struct{ err error }

// Read completes the implementation of the io.Reader interface.
func (r errReader) Read(p []byte) (int, error) { return 0, r.err }

// TestAskUser tests the AskUser() function.
func TestAskUser(t *testing.T) {
	tests := []struct {
		name         string
		in           io.Reader
		out          io.Writer
		question     string
		defaultValue string
		expected     string
		err          error
	}{
		{
			name:     "Empty string with newline",
			in:       strings.NewReader("\n"),
			out:      &bytes.Buffer{},
			question: "Core URL [http://localhost:8080]",
			expected: "",
			err:      nil,
		},
		{
			name:     "Value with newline",
			in:       strings.NewReader("http://discovery.core.cn\n"),
			out:      &bytes.Buffer{},
			question: "Core URL [http://localhost:8080]",
			expected: "http://discovery.core.cn",
			err:      nil,
		},
		{
			name:     "EOF immediately (empty reader)",
			in:       strings.NewReader(""),
			out:      &bytes.Buffer{},
			question: "Core URL [http://localhost:8080]",
			err:      nil,
		},
		{
			name:     "Value without newline then EOF",
			in:       strings.NewReader("http://discovery.core.cn"),
			out:      &bytes.Buffer{},
			question: "Core URL [http://localhost:8080]",
			expected: "http://discovery.core.cn",
			err:      nil,
		},
		{
			name:     "Write prompt returns ios.out error",
			in:       strings.NewReader("willfail\n"),
			out:      errWriter{err: errors.New("write failed")},
			question: "Core URL [http://localhost:8080]",
			err:      errors.New("write failed"),
		},
		{
			name:     "Read error (not End Of File)",
			in:       errReader{err: errors.New("read failed")},
			out:      &bytes.Buffer{},
			question: "Core URL []",
			err:      errors.New("read failed"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ios := IOStreams{
				In:  tc.in,
				Out: tc.out,
				Err: &bytes.Buffer{},
			}

			got, err := ios.AskUser(tc.question)

			if tc.err != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, got)
			}
		})
	}
}
