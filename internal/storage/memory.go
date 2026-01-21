package storage

import (
	"context"
	"errors"
	"os"
	"sync"
)

type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string][]byte
}

var ErrNotFound = errors.New("not found")

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		mu:   sync.RWMutex{},
		data: make(map[string][]byte),
	}
}

func (s *MemoryStorage) Read(ctx context.Context, filepath string) (*os.File, error) {
	s.mu.RLock()
	data, ok := s.data[filepath]
	s.mu.RUnlock()

	if !ok {
		return nil, ErrNotFound
	}
	// make by temp files
	tempFile, err := os.CreateTemp("", "")
	if err != nil {
		return nil, err
	}
	defer tempFile.Close()
	_, err = tempFile.Write(data)
	if err != nil {
		return nil, err
	}
	return tempFile, nil
}

func (s *MemoryStorage) Write(ctx context.Context, source string, filepath string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.data[filepath] = data
	s.mu.Unlock()
	return nil
}

func (s *MemoryStorage) Delete(ctx context.Context, filepath string) error {
	s.mu.Lock()
	s.data[filepath] = nil
	s.mu.Unlock()
	return nil
}