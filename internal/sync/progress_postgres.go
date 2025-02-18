package sync

import (
	"context"
	"fmt"
	"time"

	"gitea.chrnv.ru/vanadium23/kompanion/internal/entity"
	"gitea.chrnv.ru/vanadium23/kompanion/pkg/postgres"
)

// ProgressDatabaseRepo -.
type ProgressDatabaseRepo struct {
	*postgres.Postgres
}

// New -.
func NewProgressDatabaseRepo(pg *postgres.Postgres) *ProgressDatabaseRepo {
	return &ProgressDatabaseRepo{pg}
}

// Store -.
func (r *ProgressDatabaseRepo) Store(ctx context.Context, t entity.Progress) error {
	sql := `INSERT INTO sync_progress
		(koreader_partial_md5, percentage, progress, koreader_device, koreader_device_id, created_at, auth_device_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	args := []interface{}{t.Document, t.Percentage, t.Progress, t.Device, t.DeviceID, time.Unix(t.Timestamp, 0), t.AuthDeviceName}

	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("TranslationRepo - Store - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *ProgressDatabaseRepo) GetBookHistory(ctx context.Context, bookID string, limit int) ([]entity.Progress, error) {
	sql := `SELECT koreader_partial_md5, percentage, progress, koreader_device, koreader_device_id, created_at, auth_device_name
		FROM sync_progress
		WHERE koreader_partial_md5 = $1
		ORDER BY created_at DESC
		LIMIT $2`
	args := []interface{}{bookID, limit}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ProgressDatabaseRepo - GetBookHistory - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	entities := make([]entity.Progress, 0, limit)

	for rows.Next() {
		e := entity.Progress{}
		timestamp := time.Time{}

		err = rows.Scan(&e.Document, &e.Percentage, &e.Progress, &e.Device, &e.DeviceID, &timestamp, &e.AuthDeviceName)
		e.Timestamp = timestamp.Unix()
		if err != nil {
			return nil, fmt.Errorf("ProgressDatabaseRepo - GetBookHistory - rows.Scan: %w", err)
		}

		entities = append(entities, e)
	}

	return entities, nil
}
