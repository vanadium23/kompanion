package metadata

import (
	"os"
	"path/filepath"
	"testing"
)


func TestReadAllPDFBooks(t *testing.T) {
	booksDir := "../../test/test_data/books"

	err := filepath.Walk(booksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".pdf" {
			t.Logf("\nProcessing file: %s\n", path)

			file, err := os.Open(path)
			if err != nil {
				t.Logf("  Error: failed to open file - %v\n", err)
				return nil
			}
			defer file.Close()

			metadata, err := extractPdfMetadata(file)
			if err != nil {
				t.Logf("  Error: failed to parse metadata - %v\n", err)
				return nil
			}

			t.Logf("  Title: %s\n", metadata.Title)
			t.Logf("  Author: %s\n", metadata.Author)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to traverse directory: %v", err)
	}
}