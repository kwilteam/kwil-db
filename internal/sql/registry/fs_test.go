package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFS(t *testing.T) {
	// Setup a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fstest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fs := defaultFilesystem{}

	// Test MkDirAll
	dirPath := filepath.Join(tempDir, "newdir")
	err = fs.MkdirAll(dirPath, 0755)
	assert.NoError(t, err)

	// Test ForEachFile
	testFileName := "testfile.txt"
	err = os.WriteFile(filepath.Join(dirPath, testFileName), []byte("hello world"), 0666)
	assert.NoError(t, err)

	fileFound := false
	err = fs.ForEachFile(dirPath, func(name string) error {
		if name == testFileName {
			fileFound = true
		}
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, fileFound, "ForEachFile should find the test file")

	// Test Remove
	err = fs.Remove(filepath.Join(dirPath, testFileName))
	assert.NoError(t, err)

	// Test Rename
	secondTestFileName := "renamedfile.txt"
	err = fs.Rename(filepath.Join(tempDir, "newdir"), filepath.Join(tempDir, secondTestFileName))
	assert.NoError(t, err)
}
