package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCreateTemporaryFile_Success tests when creating and writing to the temporary file was successful.
func TestCreateTemporaryFile_Success(t *testing.T) {
	dir := t.TempDir()
	filename := "testfile-*"
	content := "This is a test file"

	path, err := CreateTemporaryFile(dir, filename, content)
	require.NoError(t, err)

	require.FileExists(t, path)

	require.True(t, filepath.Dir(path) == dir)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, content, string(data))
}

// TestCreateTemporaryFile_EmptyContent tests when creating an empty file.
func TestCreateTemporaryFile_EmptyContent(t *testing.T) {
	dir := t.TempDir()
	filename := "empty-*"

	path, err := CreateTemporaryFile(dir, filename, "")
	require.NoError(t, err)
	require.FileExists(t, path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Empty(t, data)
}

// TestCreateTemporaryFile_InvalidDir tests when trying to create a file in an invalid directory.
func TestCreateTemporaryFile_InvalidDir(t *testing.T) {
	invalidDir := filepath.Join(os.TempDir(), "does-not-exist")
	filename := "fail-*"

	path, err := CreateTemporaryFile(invalidDir, filename, "irrelevant")
	require.Error(t, err)
	require.Empty(t, path)
}
