package fileutils

import (
	"os"
	"path/filepath"
)

// CreateTemporaryFile creates a temporary file in the given directory. It writes the content parameter into the file and closes it.
// It returns the file's name, which is the complete path that can be used to remove the file, or an error if any occurred.
// Removing the file is the responsibility of the function's callers.
func CreateTemporaryFile(directory, name, content string) (string, error) {
	tmpfile, err := os.Create(filepath.Join(directory, name))
	if err != nil {
		return "", err
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		return "", err
	}

	if err := tmpfile.Close(); err != nil {
		return "", err
	}

	return tmpfile.Name(), nil
}
