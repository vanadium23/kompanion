package highlight

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/vanadium23/kompanion/internal/entity"
)

// HighlightSyncUseCase implements highlight synchronization logic.
type HighlightSyncUseCase struct {
	repo HighlightRepo
}

// NewHighlightSync creates a new highlight sync use case.
func NewHighlightSync(r HighlightRepo) *HighlightSyncUseCase {
	return &HighlightSyncUseCase{repo: r}
}

// Sync stores highlights and returns count of successfully synced items.
func (uc *HighlightSyncUseCase) Sync(ctx context.Context, documentID string, highlights []entity.Highlight, deviceName string) (int, error) {
	synced := 0
	for i := range highlights {
		highlights[i].DocumentID = documentID
		highlights[i].AuthDeviceName = deviceName
		highlights[i].CreatedAt = time.Now()
		highlights[i].HighlightHash = generateHash(highlights[i].Text, highlights[i].Page, highlights[i].Timestamp)

		if err := uc.repo.Store(ctx, highlights[i]); err != nil {
			// Log and continue - unique constraint violation is expected for duplicates (SYNC-01)
			continue
		}
		synced++
	}
	return synced, nil
}

// Fetch retrieves all highlights for a document.
func (uc *HighlightSyncUseCase) Fetch(ctx context.Context, documentID string) ([]entity.Highlight, error) {
	return uc.repo.GetByDocumentID(ctx, documentID)
}

// generateHash creates MD5 hash from text:page:timestamp for deduplication.
func generateHash(text, page string, timestamp int64) string {
	data := fmt.Sprintf("%s:%s:%d", text, page, timestamp)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}
