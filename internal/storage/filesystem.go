package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type FilesystemStorage struct {
	// contains filtered or unexported fields
	root string
}

func NewFilesystemStorage(root string) (*FilesystemStorage, error) {
	// Try to create the root directory
	if !strings.HasSuffix(root, "/") {
		root += "/"
	}
	dirPath := filepath.Dir(root)
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Check system writes on the provided root
	err = checkSystemWrites(root)
	if err != nil {
		return nil, err
	}

	return &FilesystemStorage{root: root}, nil
}

func (s *FilesystemStorage) Read(ctx context.Context, p string) (*os.File, error) {
	filepath := path.Join(s.root, p)
	_, err := os.Stat(filepath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return os.Open(filepath)
}

func (s *FilesystemStorage) Write(ctx context.Context, src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dst := path.Join(s.root, dest)
	dirPath := filepath.Dir(dst)
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

func (s *FilesystemStorage) Delete(ctx context.Context, p string) error {
	filepath := path.Join(s.root, p)
	err := os.Remove(filepath)
	if err != nil {
		return err
	}
	return nil
}

func checkSystemWrites(root string) error {
	// Create a temporary file in the root directory
	tempFile, err := os.CreateTemp(root, "write_test")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())

	// Try to write to the temporary file
	_, err = tempFile.WriteString("test")
	if err != nil {
		return err
	}

	return nil
}
