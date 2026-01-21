package storage

import (
	"context"
	"os"
)

type Storage interface {
	Write(ctx context.Context, source string, filepath string) error
	Read(ctx context.Context, filepath string) (*os.File, error)
	Delete(ctx context.Context, filepath string) error
}