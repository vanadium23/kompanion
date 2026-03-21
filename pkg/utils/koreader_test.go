package utils_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/vanadium23/kompanion/pkg/utils"
)

func getTestDataPath() string {
	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "..", "..")
	return filepath.Join(projectRoot, "test", "test_data")
}

func TestPartialMd5(t *testing.T) {
	testDataPath := getTestDataPath()
	expected := "5ee88058c4346a122c4ccf80e36b1dc8"
	actual, err := utils.PartialMD5(filepath.Join(testDataPath, "books", "CrimePunishment-EPUB2.epub"))
	if err != nil {
		t.Fatalf("Error calculating MD5: %v", err)
	}
	if expected != actual {
		t.Fatalf("Expected MD5 %s, got %s", expected, actual)
	}
}
