package storage_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitea.chrnv.ru/vanadium23/kompanion/internal/storage"
	"gitea.chrnv.ru/vanadium23/kompanion/pkg/postgres"
)

func TestPostgresStorage(t *testing.T) {
	t.Run("write and read file", func(t *testing.T) {
		// Setup mock
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		pg := postgres.Mock(mock)
		store := storage.NewPostgresStorage(pg)

		// Create a temporary file with some content
		content := []byte("test content")
		tmpfile, err := os.CreateTemp("", "example")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.Write(content)
		require.NoError(t, err)
		err = tmpfile.Close()
		require.NoError(t, err)

		// Expect Write query
		mock.ExpectExec("INSERT INTO storage_blob").
			WithArgs("test.txt", pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		// Test Write
		err = store.Write(context.Background(), tmpfile.Name(), "test.txt")
		require.NoError(t, err)

		// Expect Read query
		mock.ExpectQuery("SELECT file_data FROM storage_blob").
			WithArgs("test.txt").
			WillReturnRows(mock.NewRows([]string{"file_data"}).AddRow(content))

		// Test Read
		readFile, err := store.Read(context.Background(), "test.txt")
		require.NoError(t, err)
		defer os.Remove(readFile.Name())

		readContent, err := os.ReadFile(readFile.Name())
		require.NoError(t, err)
		assert.Equal(t, content, readContent)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("read non-existent file", func(t *testing.T) {
		// Setup mock
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		pg := postgres.Mock(mock)
		store := storage.NewPostgresStorage(pg)

		// Expect Read query to return no rows
		mock.ExpectQuery("SELECT file_data FROM storage_blob").
			WithArgs("non-existent.txt").
			WillReturnError(pgx.ErrNoRows)

		_, err = store.Read(context.Background(), "non-existent.txt")
		assert.Error(t, err)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("write non-existent file", func(t *testing.T) {
		// Setup mock
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		pg := postgres.Mock(mock)
		store := storage.NewPostgresStorage(pg)

		err = store.Write(context.Background(), "non-existent.txt", "test.txt")
		assert.Error(t, err)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
}
