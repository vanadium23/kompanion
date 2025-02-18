package stats_test

import (
	"io"
	"os"
	"testing"

	"gitea.chrnv.ru/vanadium23/kompanion/internal/stats"
	"gitea.chrnv.ru/vanadium23/kompanion/pkg/postgres"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestSyncer(t *testing.T) {
	// syncDatabase deletes uploaded koreader,
	// that's why we need to copy the test database
	koreaderTestSQlite := "../../test/test_data/koreader/koreader_statistics_example.sqlite3"

	fp, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	src, err := os.Open(koreaderTestSQlite)
	if err != nil {
		t.Fatalf("failed to open source file: %v", err)
	}
	io.Copy(fp, src)
	src.Close()

	pgmock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer pgmock.Close()

	pg := postgres.Mock(pgmock)

	deviceName := "test_device"

	// Expect books upsert
	pgmock.ExpectExec(`INSERT INTO koreader_books`).
		WithArgs(
			pgxmock.AnyArg(), // md5
			pgxmock.AnyArg(), // title
			pgxmock.AnyArg(), // authors
			pgxmock.AnyArg(), // notes
			pgxmock.AnyArg(), // last_open
			pgxmock.AnyArg(), // highlights
			pgxmock.AnyArg(), // pages
			pgxmock.AnyArg(), // series
			pgxmock.AnyArg(), // language
			pgxmock.AnyArg(), // total_read_time
			pgxmock.AnyArg(), // total_read_pages
			deviceName,       // device
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Expect page stats upsert
	for i := 0; i < 4; i++ {
		pgmock.ExpectExec(`INSERT INTO koreader_page_stat_data`).
			WithArgs(
				pgxmock.AnyArg(), // md5
				pgxmock.AnyArg(), // page
				pgxmock.AnyArg(), // start_time
				pgxmock.AnyArg(), // duration
				pgxmock.AnyArg(), // total_pages
				deviceName,       // device
			).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
	}
	// Sync the databases
	err = stats.SyncDatabases(fp.Name(), pg, deviceName)
	assert.NoError(t, err)

	// Verify that all expectations were met
	if err := pgmock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
