package highlight_test

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/vanadium23/kompanion/internal/entity"
	"github.com/vanadium23/kompanion/internal/highlight"
	"github.com/vanadium23/kompanion/pkg/postgres"
)

func TestHighlightRepo_Store(t *testing.T) {
	h := entity.Highlight{
		DocumentID:     "test-doc",
		Text:           "highlight text",
		Note:           "a note",
		Page:          "42",
		Chapter:        "Chapter 1",
		Drawer:         "highlight",
		Color:          "yellow",
		Timestamp:      time.Now().Unix(),
		Device:          "koreader-device",
		DeviceID:       "device-id",
		AuthDeviceName: "auth-device",
		HighlightHash:  "hash123",
	}

	mock, hdr := setupTestHighlightDatabaseRepo()
	defer mock.Close()

	mock.ExpectExec("INSERT INTO highlight_annotations").
		WithArgs(
			h.DocumentID, h.Text, h.Note, h.Page, h.Chapter, h.Drawer, h.Color,
			time.Unix(h.Timestamp, 0), h.Device, h.DeviceID, h.AuthDeviceName, h.HighlightHash,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err := hdr.Store(context.Background(), h)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestHighlightRepo_GetByDocumentID(t *testing.T) {
	mock, hdr := setupTestHighlightDatabaseRepo()
	defer mock.Close()

	documentID := "test-doc"
	now := time.Now()

	rows := pgxmock.NewRows([]string{
		"id", "koreader_partial_md5", "text", "note", "page", "chapter",
		"drawer", "color", "highlight_time", "koreader_device", "koreader_device_id",
		"auth_device_name", "created_at",
	}).
		AddRow(
			"uuid-1", documentID, "highlight text 1", "a note", "42", "Chapter 1", "highlight", "yellow", now, "koreader-device", "device-id", "auth-device", now,
		).
		AddRow(
			"uuid-2", documentID, "highlight text 2", nil, "43", nil, nil, nil, now, "koreader-device", "device-id", "auth-device", now,
		)

	mock.ExpectQuery("SELECT id, koreader_partial_md5, text, note, page, chapter").
		WithArgs(documentID).
		WillReturnRows(rows)

	highlights, err := hdr.GetByDocumentID(context.Background(), documentID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(highlights) != 2 {
		t.Fatalf("Expected 2 highlights, got %d", len(highlights))
	}

	if highlights[0].Text != "highlight text 1" {
		t.Errorf("Expected text %s, got %s", "highlight text 1", highlights[0].Text)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestHighlightRepo_GetByDocumentID_Empty(t *testing.T) {
	mock, hdr := setupTestHighlightDatabaseRepo()
	defer mock.Close()

	documentID := "nonexistent-doc"

	rows := pgxmock.NewRows([]string{
		"id", "koreader_partial_md5", "text", "note", "page", "chapter",
		"drawer", "color", "highlight_time", "koreader_device", "koreader_device_id",
		"auth_device_name", "created_at",
	})

	mock.ExpectQuery("SELECT id, koreader_partial_md5, text, note, page, chapter").
		WithArgs(documentID).
		WillReturnRows(rows)

	highlights, err := hdr.GetByDocumentID(context.Background(), documentID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(highlights) != 0 {
		t.Fatalf("Expected 0 highlights, got %d", len(highlights))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func setupTestHighlightDatabaseRepo() (pgxmock.PgxPoolIface, *highlight.HighlightDatabaseRepo) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		panic(err)
	}

	pg := postgres.Mock(mock)
	hdr := highlight.NewHighlightDatabaseRepo(pg)

	return mock, hdr
}
