package testutils

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Defines the update flag in the package.
var Update = flag.Bool("update", false, "rewrite golden files")

// Path creates the testdata directory when update is true.
func Path(t *testing.T, name string) string {
	t.Helper()
	p := filepath.Join("testdata", name+".golden")
	if *Update {
		os.MkdirAll(filepath.Dir(p), 0o755)
	}
	return p
}

// Write writes bytes to the golden file only when -update is set.
func Write(t *testing.T, name string, got []byte) {
	t.Helper()
	require.True(t, *Update)
	require.NoError(t, os.WriteFile(Path(t, name), got, 0o644))
}

// Read reads the golden file contents.
func Read(t *testing.T, name string) []byte {
	t.Helper()
	if !(*Update) {
		b, err := os.ReadFile(Path(t, name))
		require.NoError(t, err)
		return b
	}

	return nil
}

// CompareBytesOption is used to add further modifications to the expected and gotten bytes.
type CompareBytesOption func(*[]byte) error

// WithNormalizePaths changes the "\" Windows path separator to the standard "/"
func WithNormalizePaths() CompareBytesOption {
	return func(receivedBytes *[]byte) error {
		*receivedBytes = bytes.ReplaceAll(*receivedBytes, []byte("\\"), []byte("/"))
		return nil
	}
}

// CompareBytes reads the golden file and verifies that its contents and the current response are the same.
func CompareBytes(t *testing.T, name string, expected, got []byte, options ...CompareBytesOption) {
	t.Helper()
	if *Update {
		Write(t, name, got)
	} else {
		normalizedExpected := bytes.ReplaceAll(expected, []byte("\r\n"), []byte("\n"))
		normalizedGot := bytes.ReplaceAll(got, []byte("\r\n"), []byte("\n"))
		for _, opt := range options {
			err := opt(&normalizedExpected)
			require.NoError(t, err)
			err = opt(&normalizedGot)
			require.NoError(t, err)
		}
		require.Equal(t, string(normalizedExpected), string(normalizedGot))
	}
}

// ChangeDirectoryHelper changes the working directory to t.TempDir()
func ChangeDirectoryHelper(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(wd) })
	return tmp
}
