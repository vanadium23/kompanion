package sync

import (
	"context"
	"fmt"
	"time"

	"gitea.chrnv.ru/vanadium23/kompanion/internal/entity"
)

// ProgressSyncUseCase -.
type ProgressSyncUseCase struct {
	repo ProgressRepo
}

// NewProgressSync -.
func NewProgressSync(r ProgressRepo) *ProgressSyncUseCase {
	return &ProgressSyncUseCase{
		repo: r,
	}
}

func (uc *ProgressSyncUseCase) Sync(ctx context.Context, doc entity.Progress) (entity.Progress, error) {
	if doc.Timestamp == 0 {
		doc.Timestamp = time.Now().Unix()
	}
	err := uc.repo.Store(ctx, doc)
	if err != nil {
		return doc, fmt.Errorf("ProgressSyncUseCase - Sync - s.repo.Sync: %w", err)
	}

	return doc, nil
}

func (uc *ProgressSyncUseCase) Fetch(ctx context.Context, bookID string) (entity.Progress, error) {
	doc, err := uc.repo.GetBookHistory(ctx, bookID, 1)
	if err != nil {
		return entity.Progress{}, fmt.Errorf("ProgressSyncUseCase - Fetch - s.repo.GetBookHistory: %w", err)
	}

	if doc == nil {
		return entity.Progress{}, nil
	}

	last := doc[0]
	// rewrite koreader device with our authed device
	last.Device = last.AuthDeviceName

	return last, nil
}
