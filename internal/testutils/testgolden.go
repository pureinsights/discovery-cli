package testutils

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var Update = flag.Bool("update", false, "rewrite golden files")

func Path(t *testing.T, name string) string {
	t.Helper()
	p := filepath.Join("testdata", name+".golden")
	if *Update {
		os.MkdirAll(filepath.Dir(p), 0o755)
	}
	return p
}

// Write writes bytes to the golden file (only when -update is set).
func Write(t *testing.T, name string, got []byte) {
	t.Helper()
	require.True(t, *Update)
	fmt.Println("Write")
	require.NoError(t, os.WriteFile(Path(t, name), got, 0o644))
}

// Read reads the golden file contents.
func Read(t *testing.T, name string) []byte {
	t.Helper()
	fmt.Println("Read")
	b, err := os.ReadFile(Path(t, name))
	require.NoError(t, err)
	return b
}

func CompareBytes(t *testing.T, name string, got []byte) {
	t.Helper()
	if *Update {
		Write(t, name, got)
	} else {
		expected := Read(t, name)
		require.True(t, bytes.Equal(got, expected))
	}
}
