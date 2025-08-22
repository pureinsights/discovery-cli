package testutils

import (
	"os"
)

// CreateTemporaryFile creates a temporary file in the given directory.
func CreateTemporaryFile(directory, name, content string) (string, error) {
	tmpfile, err := os.CreateTemp(directory, name)
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
