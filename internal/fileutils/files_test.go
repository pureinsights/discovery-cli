package fileutils

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Defines the update flag in the package
var Update = flag.Bool("update", false, "rewrite golden files")

// TestCreateTemporaryFile_Success tests when creating and writing to the temporary file was successful.
// It tests what happens when either of the parameters is empty.
func TestCreateTemporaryFile_Success(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "File with name and real content",
			filename: "testfile.txt",
			content:  "This is a test file",
		},
		{
			name:     "File with name and empty content",
			filename: "empty.txt",
			content:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path, err := CreateTemporaryFile(dir, tc.filename, tc.content)
			require.NoError(t, err)

			require.FileExists(t, path)

			require.True(t, filepath.Dir(path) == dir)

			data, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, tc.content, string(data))
		})
	}
}

// TestCreateTemporaryFile_InvalidDir tests when trying to create a file in an invalid directory.
func TestCreateTemporaryFile_InvalidDir(t *testing.T) {
	invalidDir := filepath.Join(os.TempDir(), "does-not-exist")
	filename := "fail.txt"

	path, err := CreateTemporaryFile(invalidDir, filename, "test")
	require.Error(t, err)
	require.Empty(t, path)
}

func TestCreateTemporaryFile_EmptyFile(t *testing.T) {
	invalidDir := os.TempDir()
	filename := ""

	path, err := CreateTemporaryFile(invalidDir, filename, "test")
	require.Error(t, err)
	require.NoFileExists(t, path)
}
