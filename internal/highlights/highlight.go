package highlights

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/vanadium23/kompanion/internal/entity"
	"github.com/vanadium23/kompanion/pkg/logger"
)

// HighlightSyncUseCase implements HighlightSync interface.
type HighlightSyncUseCase struct {
	repo HighlightRepo
	l    logger.Interface
}

// NewHighlightSyncUseCase creates a new HighlightSyncUseCase.
func NewHighlightSyncUseCase(r HighlightRepo, l logger.Interface) *HighlightSyncUseCase {
	return &HighlightSyncUseCase{
		repo: r,
		l:    l,
	}
}

// HighlightListUseCase implements HighlightList interface.
type HighlightListUseCase struct {
	repo HighlightRepo
}

// NewHighlightListUseCase creates a new HighlightListUseCase.
func NewHighlightListUseCase(r HighlightRepo) *HighlightListUseCase {
	return &HighlightListUseCase{
		repo: r,
	}
}

// Sync processes a batch of highlights from KOReader and stores them.
func (uc *HighlightSyncUseCase) Sync(ctx context.Context, req entity.HighlightSyncRequest, deviceName string) (int, int, error) {
	total := len(req.Entries)

	if total == 0 {
		return 0, 0, nil
	}

	highlights := make([]entity.Highlight, 0, total)

	for _, entry := range req.Entries {
		// Compute SHA-256 hash of the text for deduplication
		hash := sha256.Sum256([]byte(entry.Text))
		textHash := hex.EncodeToString(hash[:])

		highlight := entity.Highlight{
			KoreaderPartialMD5: req.Document,
			TextHash:           textHash,
			Text:               entry.Text,
			Note:               entry.Note,
			Page:               entry.Page,
			Chapter:            entry.Chapter,
			Time:               entry.Time,
			Drawer:             entry.Drawer,
			Color:              entry.Color,
			DeviceName:         deviceName,
		}

		highlights = append(highlights, highlight)
	}

	synced, err := uc.repo.SyncHighlights(ctx, highlights)
	if err != nil {
		return 0, total, fmt.Errorf("HighlightSyncUseCase - Sync - uc.repo.SyncHighlights: %w", err)
	}

	return synced, total, nil
}

// List returns all highlights for a given document.
func (uc *HighlightListUseCase) List(ctx context.Context, koreaderPartialMD5 string) ([]entity.Highlight, error) {
	return uc.repo.ListHighlights(ctx, koreaderPartialMD5)
}
