package highlights

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vanadium23/kompanion/internal/entity"
	"github.com/vanadium23/kompanion/pkg/postgres"
)

// HighlightDatabaseRepo implements HighlightRepo interface for PostgreSQL.
type HighlightDatabaseRepo struct {
	*postgres.Postgres
}

// NewHighlightDatabaseRepo creates a new HighlightDatabaseRepo.
func NewHighlightDatabaseRepo(pg *postgres.Postgres) *HighlightDatabaseRepo {
	return &HighlightDatabaseRepo{pg}
}

// SyncHighlights performs batch upsert of highlights with timestamp-gated updates.
func (r *HighlightDatabaseRepo) SyncHighlights(ctx context.Context, highlights []entity.Highlight) (int, error) {
	if len(highlights) == 0 {
		return 0, nil
	}

	// Cast to pgxpool.Pool to access Begin
	pool, ok := r.Pool.(*pgxpool.Pool)
	if !ok {
		return 0, fmt.Errorf("HighlightDatabaseRepo - SyncHighlights - pool cast failed: cannot cast to *pgxpool.Pool")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("HighlightDatabaseRepo - SyncHighlights - pool.Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	sql := `INSERT INTO sync_highlight (
		koreader_partial_md5, text_hash, text, note, page, chapter,
		time, drawer, color, device_name
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	ON CONFLICT (koreader_partial_md5, text_hash) DO UPDATE SET
		note = EXCLUDED.note,
		chapter = EXCLUDED.chapter,
		page = EXCLUDED.page,
		time = EXCLUDED.time,
		drawer = EXCLUDED.drawer,
		color = EXCLUDED.color,
		device_name = EXCLUDED.device_name
	WHERE EXCLUDED.time > sync_highlight.time`

	for _, h := range highlights {
		args := []interface{}{
			h.KoreaderPartialMD5,
			h.TextHash,
			h.Text,
			h.Note,
			h.Page,
			h.Chapter,
			h.Time,
			h.Drawer,
			h.Color,
			h.DeviceName,
		}

		_, err := tx.Exec(ctx, sql, args...)
		if err != nil {
			return 0, fmt.Errorf("HighlightDatabaseRepo - SyncHighlights - tx.Exec: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("HighlightDatabaseRepo - SyncHighlights - tx.Commit: %w", err)
	}

	return len(highlights), nil
}

// ListHighlights retrieves all highlights for a specific book, sorted by page.
func (r *HighlightDatabaseRepo) ListHighlights(ctx context.Context, koreaderPartialMD5 string) ([]entity.Highlight, error) {
	query := `SELECT koreader_partial_md5, text_hash, text, note, page, chapter,
		time, drawer, color, device_name, created_at
		FROM sync_highlight
		WHERE koreader_partial_md5 = $1
		ORDER BY CASE
			WHEN page ~ '^[0-9]+$' THEN CAST(page AS INTEGER)
			ELSE 0
		END ASC`

	rows, err := r.Pool.Query(ctx, query, koreaderPartialMD5)
	if err != nil {
		return nil, fmt.Errorf("HighlightDatabaseRepo - ListHighlights - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	var highlights []entity.Highlight
	for rows.Next() {
		var h entity.Highlight
		err := rows.Scan(&h.KoreaderPartialMD5, &h.TextHash, &h.Text, &h.Note,
			&h.Page, &h.Chapter, &h.Time, &h.Drawer, &h.Color, &h.DeviceName, &h.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("HighlightDatabaseRepo - ListHighlights - rows.Scan: %w", err)
		}
		highlights = append(highlights, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("HighlightDatabaseRepo - ListHighlights - rows.Err: %w", err)
	}

	if highlights == nil {
		highlights = []entity.Highlight{}
	}

	return highlights, nil
}
