package sync

import (
	"context"

	"github.com/vanadium23/kompanion/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=sync_test

type ProgressRepo interface {
	Store(ctx context.Context, t entity.Progress) error
	GetBookHistory(ctx context.Context, bookID string, limit int) ([]entity.Progress, error)
}

// Progress -.
type Progress interface {
	Sync(context.Context, entity.Progress) (entity.Progress, error)
	Fetch(ctx context.Context, bookID string) (entity.Progress, error)
}
