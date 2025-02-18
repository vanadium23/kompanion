package library_test

import (
	"context"
	"testing"
	"time"

	"gitea.chrnv.ru/vanadium23/kompanion/internal/entity"
	"gitea.chrnv.ru/vanadium23/kompanion/internal/library"
	"gitea.chrnv.ru/vanadium23/kompanion/pkg/postgres"
	"github.com/pashagolub/pgxmock/v4"
)

func TestBookDatabaseRepoCreate(t *testing.T) {
	// book
	book := entity.Book{
		ID:         "1",
		Title:      "title",
		Author:     "author",
		Publisher:  "publisher",
		Year:       2021,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		ISBN:       "isbn",
		FilePath:   "file_path",
		DocumentID: "document_id",
		CoverPath:  "cover_path",
	}

	// создать mock
	mock, bdr := setupTestBookDatabaseRepo()
	defer mock.Close()

	mock.ExpectExec("INSERT INTO library_book").
		WithArgs(book.ID, book.Title, book.Author, book.Publisher, book.Year, book.CreatedAt, book.UpdatedAt, book.ISBN, book.FilePath, book.DocumentID, book.CoverPath).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// вызвать Create
	err := bdr.Store(context.Background(), book)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBookDatabaseRepoGetById(t *testing.T) {
	// book
	book := entity.Book{
		ID:         "1",
		Title:      "title",
		Author:     "author",
		Publisher:  "publisher",
		Year:       2021,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		ISBN:       "isbn",
		FilePath:   "file_path",
		DocumentID: "document_id",
		CoverPath:  "cover_path",
	}

	// создать mock
	mock, bdr := setupTestBookDatabaseRepo()
	defer mock.Close()

	rows := pgxmock.NewRows([]string{"id", "title", "author", "publisher", "year", "created_at", "updated_at", "isbn", "file_path", "file_hash", "cover_path"}).
		AddRow(book.ID, book.Title, book.Author, book.Publisher, book.Year, book.CreatedAt, book.UpdatedAt, book.ISBN, book.FilePath, book.DocumentID, book.CoverPath)

	mock.ExpectQuery("SELECT (.+) FROM library_book").
		WithArgs(book.ID).
		WillReturnRows(rows)

	// вызвать GetById
	result, err := bdr.GetById(context.Background(), book.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result.DocumentID != book.DocumentID {
		t.Errorf("expected DocumentID %v, got %v", book.DocumentID, result.DocumentID)
	}
}

func TestBookDatabaseRepoGetByFileHash(t *testing.T) {
	// book
	book := entity.Book{
		ID:         "1",
		Title:      "title",
		Author:     "author",
		Publisher:  "publisher",
		Year:       2021,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		ISBN:       "isbn",
		FilePath:   "file_path",
		DocumentID: "document_id",
		CoverPath:  "cover_path",
	}

	// создать mock
	mock, bdr := setupTestBookDatabaseRepo()
	defer mock.Close()

	rows := pgxmock.NewRows([]string{"id", "title", "author", "publisher", "year", "created_at", "updated_at", "isbn", "file_path", "file_hash", "cover_path"}).
		AddRow(book.ID, book.Title, book.Author, book.Publisher, book.Year, book.CreatedAt, book.UpdatedAt, book.ISBN, book.FilePath, book.DocumentID, book.CoverPath)

	mock.ExpectQuery("SELECT (.+) FROM library_book").
		WithArgs(book.DocumentID).
		WillReturnRows(rows)

	// вызвать GetByFileHash
	result, err := bdr.GetByFileHash(context.Background(), book.DocumentID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result.DocumentID != book.DocumentID {
		t.Errorf("expected DocumentID %v, got %v", book.DocumentID, result.DocumentID)
	}
}

func TestBookDatabaseRepoList(t *testing.T) {
	// book
	book := entity.Book{
		ID:         "1",
		Title:      "title",
		Author:     "author",
		Publisher:  "publisher",
		Year:       2021,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		ISBN:       "isbn",
		FilePath:   "file_path",
		DocumentID: "document_id",
		CoverPath:  "cover_path",
	}

	// создать mock
	mock, bdr := setupTestBookDatabaseRepo()
	defer mock.Close()

	rows := pgxmock.NewRows([]string{"id", "title", "author", "publisher", "year", "created_at", "updated_at", "isbn", "file_path", "file_hash", "cover_path"}).
		AddRow(book.ID, book.Title, book.Author, book.Publisher, book.Year, book.CreatedAt, book.UpdatedAt, book.ISBN, book.FilePath, book.DocumentID, book.CoverPath)

	mock.ExpectQuery("SELECT (.+) FROM library_book").
		WillReturnRows(rows)

	// вызвать List
	results, err := bdr.List(context.Background(), "created_at", "desc", 1, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %v", len(results))
	}

	if results[0].DocumentID != book.DocumentID {
		t.Errorf("expected DocumentID %v, got %v", book.DocumentID, results[0].DocumentID)
	}
}

func setupTestBookDatabaseRepo() (pgxmock.PgxPoolIface, *library.BookDatabaseRepo) {
	// создать mock
	mock, err := pgxmock.NewPool()
	if err != nil {
		panic(err)
	}

	// создать BookDatabaseRepo
	pg := postgres.Mock(mock)
	bdr := library.NewBookDatabaseRepo(pg)

	return mock, bdr
}
