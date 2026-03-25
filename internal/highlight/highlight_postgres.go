package highlight

import (
	"context"
	"fmt"
    "time"

    "github.com/vanadium23/kompanion/internal/entity"
    "github.com/vanadium23/kompanion/pkg/postgres"
)

// HighlightDatabaseRepo implements HighlightRepo using PostgreSQL.
type HighlightDatabaseRepo struct {
    *postgres.Postgres
}

// NewHighlightDatabaseRepo creates a new highlight repository.
func NewHighlightDatabaseRepo(pg *postgres.Postgres) *HighlightDatabaseRepo {
    return &HighlightDatabaseRepo{pg}
}

// Store inserts a highlight with ON CONFLICT DO NOTHING for deduplication.
func (r *HighlightDatabaseRepo) Store(ctx context.Context, h entity.Highlight) error {
    sql := `INSERT INTO highlight_annotations
        (koreader_partial_md5, text, note, page, chapter, drawer, color,
         highlight_time, koreader_device, koreader_device_id, auth_device_name, highlight_hash)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        ON CONFLICT (koreader_partial_md5, highlight_hash) DO NOTHING`

    args := []interface{}{
        h.DocumentID, h.Text, h.Note, h.Page, h.Chapter, h.Drawer, h.Color,
        time.Unix(h.Timestamp, 0), h.Device, h.DeviceID, h.AuthDeviceName, h.HighlightHash,
    }

    _, err := r.Pool.Exec(ctx, sql, args...)
    if err != nil {
        return fmt.Errorf("HighlightDatabaseRepo - Store - r.Pool.Exec: %w", err)
    }
    return nil
}

// GetByDocumentID retrieves all highlights for a document, ordered by time.
func (r *HighlightDatabaseRepo) GetByDocumentID(ctx context.Context, documentID string) ([]entity.Highlight, error) {
    sql := `SELECT id, koreader_partial_md5, text, note, page, chapter, drawer, color,
            highlight_time, koreader_device, koreader_device_id, auth_device_name, created_at
            FROM highlight_annotations
            WHERE koreader_partial_md5 = $1
            ORDER BY highlight_time ASC`

    rows, err := r.Pool.Query(ctx, sql, documentID)
    if err != nil {
        return nil, fmt.Errorf("HighlightDatabaseRepo - GetByDocumentID - r.Pool.Query: %w", err)
    }
    defer rows.Close()

    var highlights []entity.Highlight
    for rows.Next() {
        var h entity.Highlight
        var highlightTime time.Time
        err = rows.Scan(&h.ID, &h.DocumentID, &h.Text, &h.Note, &h.Page, &h.Chapter,
            &h.Drawer, &h.Color, &highlightTime, &h.Device, &h.DeviceID, &h.AuthDeviceName, &h.CreatedAt)
        if err != nil {
            return nil, fmt.Errorf("HighlightDatabaseRepo - GetByDocumentID - rows.Scan: %w", err)
        }
        h.Timestamp = highlightTime.Unix()
        highlights = append(highlights, h)
    }
    return highlights, nil
}

// GetDocumentsByDevice retrieves unique documents with highlights for a device.
func (r *HighlightDatabaseRepo) GetDocumentsByDevice(ctx context.Context, deviceName string) ([]DocumentInfo, error) {
    sql := `SELECT DISTINCT h.koreader_partial_md5, b.title, b.author
            FROM highlight_annotations h
            LEFT JOIN books b ON h.koreader_partial_md5 = b.partial_md5
            WHERE h.auth_device_name = $1
            ORDER BY b.title`

    rows, err := r.Pool.Query(ctx, sql, deviceName)
    if err != nil {
        return nil, fmt.Errorf("HighlightDatabaseRepo - GetDocumentsByDevice - r.Pool.Query: %w", err)
    }
    defer rows.Close()

    var docs []DocumentInfo
    for rows.Next() {
        var doc DocumentInfo
        err = rows.Scan(&doc.PartialMD5, &doc.Title, &doc.Author)
        if err != nil {
            return nil, fmt.Errorf("HighlightDatabaseRepo - GetDocumentsByDevice - rows.Scan: %w", err)
        }
        docs = append(docs, doc)
    }
    return docs, nil
}
