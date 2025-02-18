package sync_test

import (
	"context"
	"testing"
	"time"

	"gitea.chrnv.ru/vanadium23/kompanion/internal/entity"
	"gitea.chrnv.ru/vanadium23/kompanion/internal/sync"
	"gitea.chrnv.ru/vanadium23/kompanion/pkg/postgres"
	"github.com/pashagolub/pgxmock/v4"
)

func TestProgressRepo_Store(t *testing.T) {
	pr := entity.Progress{
		Document:       "test",
		Percentage:     1.5,
		Progress:       "some",
		Device:         "some",
		DeviceID:       "test",
		Timestamp:      time.Now().Unix(),
		AuthDeviceName: "nothing",
	}

	mock, pdr := setupTestProgressDatabaseRepo()
	defer mock.Close()

	mock.ExpectExec("INSERT INTO sync_progress").
		WithArgs(pr.Document, pr.Percentage, pr.Progress, pr.Device, pr.DeviceID, time.Unix(pr.Timestamp, 0), pr.AuthDeviceName).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err := pdr.Store(context.Background(), pr)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestProgressRepo_GetBookHistory(t *testing.T) {
	mock, pdr := setupTestProgressDatabaseRepo()
	defer mock.Close()

	bookID := "test-book"
	limit := 10
	now := time.Now()
	expectedProgress := []entity.Progress{
		{
			Document:       bookID,
			Percentage:     1.5,
			Progress:       "some",
			Device:         "some",
			DeviceID:       "test",
			Timestamp:      now.Unix(),
			AuthDeviceName: "nothing",
		},
	}

	rows := pgxmock.NewRows([]string{"koreader_partial_md5", "percentage", "progress", "koreader_device", "koreader_device_id", "created_at", "auth_device_name"}).
		AddRow(expectedProgress[0].Document, expectedProgress[0].Percentage, expectedProgress[0].Progress,
			expectedProgress[0].Device, expectedProgress[0].DeviceID, now, expectedProgress[0].AuthDeviceName)

	mock.ExpectQuery("SELECT koreader_partial_md5, percentage, progress, koreader_device, koreader_device_id, created_at, auth_device_name").
		WithArgs(bookID, limit).
		WillReturnRows(rows)

	progress, err := pdr.GetBookHistory(context.Background(), bookID, limit)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(progress) != 1 {
		t.Fatalf("Expected 1 progress record, got %d", len(progress))
	}

	if progress[0].Document != expectedProgress[0].Document {
		t.Errorf("Expected document %s, got %s", expectedProgress[0].Document, progress[0].Document)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func setupTestProgressDatabaseRepo() (pgxmock.PgxPoolIface, *sync.ProgressDatabaseRepo) {
	// создать mock
	mock, err := pgxmock.NewPool()
	if err != nil {
		panic(err)
	}

	// создать BookDatabaseRepo
	pg := postgres.Mock(mock)
	bdr := sync.NewProgressDatabaseRepo(pg)

	return mock, bdr
}
