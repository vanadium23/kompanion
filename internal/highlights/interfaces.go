// Package highlights provides highlight synchronization logic.
package highlights

import (
	"context"

	"github.com/vanadium23/kompanion/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=highlights_test

// HighlightRepo defines repository interface for highlight storage.
type HighlightRepo interface {
	SyncHighlights(ctx context.Context, highlights []entity.Highlight) (int, error)
	ListHighlights(ctx context.Context, koreaderPartialMD5 string) ([]entity.Highlight, error)
}

// HighlightSync defines use case interface for highlight synchronization.
type HighlightSync interface {
	Sync(ctx context.Context, req entity.HighlightSyncRequest, deviceName string) (int, int, error)
}

// HighlightList defines use case interface for listing highlights.
type HighlightList interface {
	List(ctx context.Context, koreaderPartialMD5 string) ([]entity.Highlight, error)
}
