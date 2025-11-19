package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPath_CreatesTestdataDirWhenUpdate tests the Path() function when the Update flag is true
func TestPath_CreatesTestdataDirWhenUpdate(t *testing.T) {
	ChangeDirectoryHelper(t)

	old := *Update
	*Update = true
	t.Cleanup(func() { *Update = old })

	p := Path(t, "test")
	require.Equal(t, filepath.Join("testdata", "test.golden"), p)

	info, err := os.Stat(filepath.Dir(p))
	require.NoError(t, err)
	require.True(t, info.IsDir())
}

// TestPath_DoesNotCreateTestdataDirWhenNotUpdate tests the Path() function when update is false
func TestPath_DoesNotCreateTestdataDirWhenNotUpdate(t *testing.T) {
	ChangeDirectoryHelper(t)

	old := *Update
	*Update = false
	t.Cleanup(func() { *Update = old })

	p := Path(t, "test")
	require.Equal(t, filepath.Join("testdata", "test.golden"), p)

	_, err := os.Stat(filepath.Dir(p))
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}

// TestWrite_UpdateTrueWritesFile tests the Write() function fails when update is false.
func TestWrite_UpdateTrueWritesFile(t *testing.T) {
	ChangeDirectoryHelper(t)

	old := *Update
	*Update = true
	t.Cleanup(func() { *Update = old })

	Write(t, "test", []byte("test result\n"))

	b, err := os.ReadFile(filepath.Join("testdata", "test.golden"))
	require.NoError(t, err)
	require.Equal(t, []byte("test result\n"), b)
}

// TestRead_SucceedsWhenFileExists tests the Read() function when the file exists.
func TestRead_SucceedsWhenFileExists(t *testing.T) {
	ChangeDirectoryHelper(t)

	require.NoError(t, os.MkdirAll("testdata", 0o755))
	require.NoError(t, os.WriteFile(filepath.Join("testdata", "test.golden"), []byte("this is a test\n"), 0o644))

	old := *Update
	*Update = false
	t.Cleanup(func() { *Update = old })

	got := Read(t, "test")
	require.Equal(t, []byte("this is a test\n"), got)
}

// TestRead_ReturnsNilWhenUpdateTrue tests the Read function when the update flag is true
func TestRead_ReturnsNilWhenUpdateTrue(t *testing.T) {
	ChangeDirectoryHelper(t)

	old := *Update
	*Update = true
	t.Cleanup(func() { *Update = old })

	got := Read(t, "test")
	require.Nil(t, got)
}

// TestCompareBytes_UpdateWritesNewGolden tests the CompareBytes() function when the golden file needs to be updated.
func TestCompareBytes_UpdateWritesNewGolden(t *testing.T) {
	ChangeDirectoryHelper(t)

	old := *Update
	*Update = true
	t.Cleanup(func() { *Update = old })

	CompareBytes(t, "test", []byte{}, []byte("test data\n"))

	// verify file was created with content
	b, err := os.ReadFile(filepath.Join("testdata", "test.golden"))
	require.NoError(t, err)
	require.Equal(t, []byte("test data\n"), b)
}

// TestCompareBytes_NoUpdateMatchesPasses tests the CompareBytes() function when the golden file is the same as the result
func TestCompareBytes_NoUpdateMatchesPasses(t *testing.T) {
	ChangeDirectoryHelper(t)

	require.NoError(t, os.MkdirAll("testdata", 0o755))
	require.NoError(t, os.WriteFile(filepath.Join("testdata", "test.golden"), []byte("this is a test\n"), 0o644))

	old := *Update
	*Update = false
	t.Cleanup(func() { *Update = old })

	CompareBytes(t, "test", Read(t, "test"), []byte("this is a test\n"))
}
