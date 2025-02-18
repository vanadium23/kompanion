package storage_test

import (
	"context"
	"os"
	"testing"

	"gitea.chrnv.ru/vanadium23/kompanion/internal/storage"
)

func TestMemoryStorage(t *testing.T) {
	ctx := context.Background()
	storage := storage.NewMemoryStorage()

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

	err = storage.Write(ctx, tempFile.Name(), "test")
	if err != nil {
		t.Errorf("Error writing file: %v", err)
	}

	readFile, err := storage.Read(ctx, "test")
	defer readFile.Close()
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}
	readBody, err := os.ReadFile(readFile.Name())
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}
	if string(readBody) != string(body) {
		t.Errorf("Expected body %s, got %s", string(body), string(readBody))
	}
}
