package storage_test

import (
	"context"
	"os"
	"testing"

	"github.com/vanadium23/kompanion/internal/storage"
)

func TestFilesystemStorage(t *testing.T) {
	ctx := context.Background()
	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("Error creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	st, err := storage.NewFilesystemStorage(tmpdir)
	if err != nil {
		t.Fatalf("Error creating filesystem storage: %v", err)
	}

	body := []byte("Hello, World!")
	tempFile, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	_, err = tempFile.Write(body)
	if err != nil {
		t.Fatalf("Error writing to temp file: %v", err)
	}

	defer os.Remove(tempFile.Name())

	err = st.Write(ctx, tempFile.Name(), "test")
	if err != nil {
		t.Errorf("Error writing file: %v", err)
	}

	readFile, err := st.Read(ctx, "test")
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}

	defer readFile.Close()
	readBody, err := os.ReadFile(readFile.Name())
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}
	if string(readBody) != string(body) {
		t.Errorf("Expected body %s, got %s", string(body), string(readBody))
	}

	err = st.Delete(ctx, "test")
	if err != nil {
		t.Errorf("Error deleting file: %v", err)
	}

	defer readFile.Close()
	_, err =os.ReadFile(readFile.Name())
	if err == nil {
		t.Errorf("Error deleting file failed")
	}
}
