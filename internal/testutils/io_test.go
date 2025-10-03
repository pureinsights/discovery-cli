package testutils

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestErrWriter_Write tests the ErrWriter.Write() function.
func TestErrWriter_Write(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		expectedN int
	}{
		{
			name:      "with error",
			err:       errors.New("write failed"),
			expectedN: 0,
		},
		{
			name:      "no error",
			err:       nil,
			expectedN: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := ErrWriter{Err: tc.err}
			n, err := w.Write([]byte("test"))

			assert.Equal(t, tc.expectedN, n)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestErrReader_Read test the ErrReader.Read() function.
func TestErrReader_Read(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedN   int
		expectedErr error
	}{
		{
			name:      "with error",
			err:       errors.New("read failed"),
			expectedN: 0,
		},
		{
			name:      "no error",
			err:       nil,
			expectedN: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := ErrReader{Err: tc.err}
			buf := make([]byte, 8)

			n, err := r.Read(buf)

			assert.Equal(t, tc.expectedN, n)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestFailOnNWriter tests that the FailOnWriter.Write() function fails at the desired N.
func TestFailOnNWriter(t *testing.T) {
	tests := []struct {
		name      string
		N         int
		expectedN int
	}{
		{
			name:      "fails on first write when N=1",
			N:         1,
			expectedN: 1,
		},
		{
			name:      "fails on second write when N=2",
			N:         2,
			expectedN: 2,
		},
		{
			name:      "fails on third write when N=3",
			N:         3,
			expectedN: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := &FailOnNWriter{Writer: &bytes.Buffer{}, N: tc.N}

			for i := 1; i <= tc.expectedN; i++ {
				_, err := w.Write([]byte("x"))

				if i == tc.expectedN {
					require.Error(t, err, "expected error on write %d", i)
				} else {
					require.NoError(t, err, "did not expect error on write %d", i)
				}
			}
		})
	}
}
