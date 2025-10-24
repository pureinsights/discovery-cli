package testutils

import (
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
