package utils_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/vanadium23/kompanion/pkg/utils"
)

// getProjectRoot returns the absolute path to project root
func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	// pkg/utils/koreader_test.go -> pkg/utils -> pkg -> kompanion
	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}

func TestPartialMd5(t *testing.T) {
	expected := "5ee88058c4346a122c4ccf80e36b1dc8"
	testDataPath := filepath.Join(getProjectRoot(), "test", "test_data", "books", "CrimePunishment-EPUB2.epub")
	actual, err := utils.PartialMD5(testDataPath)
	if err != nil {
		t.Fatalf("Error calculating MD5: %v", err)
	}
	if expected != actual {
		t.Fatalf("Expected MD5 %s, got %s", expected, actual)
	}
}
