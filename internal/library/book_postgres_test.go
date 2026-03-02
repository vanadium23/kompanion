package library_test

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/vanadium23/kompanion/internal/entity"
	"github.com/vanadium23/kompanion/internal/library"
	"github.com/vanadium23/kompanion/pkg/postgres"
)

func TestBookDatabaseRepoCreate(t *testing.T) {
	// book
	seriesIndex := decimal.NewNullDecimal(decimal.RequireFromString("1.5"))
	book := entity.Book{
		ID:          "1",
		Title:       "title",
		Author:      "author",
		Publisher:   "publisher",
		Year:        2021,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ISBN:        "isbn",
		FilePath:    "file_path",
		DocumentID:  "document_id",
		CoverPath:   "cover_path",
		Series:      "Test Series",
		SeriesIndex: &seriesIndex,
	}

	// создать mock
	mock, bdr := setupTestBookDatabaseRepo()
	defer mock.Close()

	mock.ExpectExec("INSERT INTO library_book").
		WithArgs(book.ID, book.Title, book.Author, book.Publisher, book.Year, book.CreatedAt, book.UpdatedAt, book.ISBN, book.FilePath, book.DocumentID, book.CoverPath, book.Series, book.SeriesIndex).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// вызвать Create
	err := bdr.Store(context.Background(), book)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBookDatabaseRepoCreateWithoutSeries(t *testing.T) {
	// book without series
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
		WithArgs(book.ID, book.Title, book.Author, book.Publisher, book.Year, book.CreatedAt, book.UpdatedAt, book.ISBN, book.FilePath, book.DocumentID, book.CoverPath, book.Series, book.SeriesIndex).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// вызвать Create
	err := bdr.Store(context.Background(), book)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBookDatabaseRepoGetById(t *testing.T) {
	// book
	seriesIndex := decimal.NewNullDecimal(decimal.RequireFromString("2"))
	book := entity.Book{
		ID:          "1",
		Title:       "title",
		Author:      "author",
		Publisher:   "publisher",
		Year:        2021,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ISBN:        "isbn",
		FilePath:    "file_path",
		DocumentID:  "document_id",
		CoverPath:   "cover_path",
		Series:      "Test Series",
		SeriesIndex: &seriesIndex,
	}

	// создать mock
	mock, bdr := setupTestBookDatabaseRepo()
	defer mock.Close()

	// Pass float64 for series_index - pgxmock will use scanner to convert
	rows := pgxmock.NewRows([]string{"id", "title", "author", "publisher", "year", "created_at", "updated_at", "isbn", "file_path", "file_hash", "cover_path", "series", "series_index"}).
		AddRow(book.ID, book.Title, book.Author, book.Publisher, book.Year, book.CreatedAt, book.UpdatedAt, book.ISBN, book.FilePath, book.DocumentID, book.CoverPath, book.Series, 2.0)

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
	if result.Series != book.Series {
		t.Errorf("expected Series %v, got %v", book.Series, result.Series)
	}
	if result.SeriesIndex == nil || !result.SeriesIndex.Valid || !book.SeriesIndex.Valid ||
		!result.SeriesIndex.Decimal.Equal(book.SeriesIndex.Decimal) {
		t.Errorf("expected SeriesIndex %v, got %v", book.SeriesIndex, result.SeriesIndex)
	}
}

func TestBookDatabaseRepoGetByIdWithoutSeries(t *testing.T) {
	// book without series
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

	rows := pgxmock.NewRows([]string{"id", "title", "author", "publisher", "year", "created_at", "updated_at", "isbn", "file_path", "file_hash", "cover_path", "series", "series_index"}).
		AddRow(book.ID, book.Title, book.Author, book.Publisher, book.Year, book.CreatedAt, book.UpdatedAt, book.ISBN, book.FilePath, book.DocumentID, book.CoverPath, book.Series, nil)

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
	if result.SeriesIndex != nil {
		t.Errorf("expected SeriesIndex nil, got %v", *result.SeriesIndex)
	}
}

func TestBookDatabaseRepoGetByFileHash(t *testing.T) {
	// book
	seriesIndex := decimal.NewNullDecimal(decimal.RequireFromString("1"))
	book := entity.Book{
		ID:          "1",
		Title:       "title",
		Author:      "author",
		Publisher:   "publisher",
		Year:        2021,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ISBN:        "isbn",
		FilePath:    "file_path",
		DocumentID:  "document_id",
		CoverPath:   "cover_path",
		Series:      "Test Series",
		SeriesIndex: &seriesIndex,
	}

	// создать mock
	mock, bdr := setupTestBookDatabaseRepo()
	defer mock.Close()

	// Pass float64 for series_index - pgxmock will use scanner to convert
	rows := pgxmock.NewRows([]string{"id", "title", "author", "publisher", "year", "created_at", "updated_at", "isbn", "file_path", "file_hash", "cover_path", "series", "series_index"}).
		AddRow(book.ID, book.Title, book.Author, book.Publisher, book.Year, book.CreatedAt, book.UpdatedAt, book.ISBN, book.FilePath, book.DocumentID, book.CoverPath, book.Series, 1.0)

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
	if result.Series != book.Series {
		t.Errorf("expected Series %v, got %v", book.Series, result.Series)
	}
}

func TestBookDatabaseRepoList(t *testing.T) {
	// book
	seriesIndex := decimal.NewNullDecimal(decimal.RequireFromString("3.5"))
	book := entity.Book{
		ID:          "1",
		Title:       "title",
		Author:      "author",
		Publisher:   "publisher",
		Year:        2021,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ISBN:        "isbn",
		FilePath:    "file_path",
		DocumentID:  "document_id",
		CoverPath:   "cover_path",
		Series:      "Test Series",
		SeriesIndex: &seriesIndex,
	}

	// создать mock
	mock, bdr := setupTestBookDatabaseRepo()
	defer mock.Close()

	// Pass float64 for series_index - pgxmock will use scanner to convert
	rows := pgxmock.NewRows([]string{"id", "title", "author", "publisher", "year", "created_at", "updated_at", "isbn", "file_path", "file_hash", "cover_path", "series", "series_index"}).
		AddRow(book.ID, book.Title, book.Author, book.Publisher, book.Year, book.CreatedAt, book.UpdatedAt, book.ISBN, book.FilePath, book.DocumentID, book.CoverPath, book.Series, 3.5)

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
	if results[0].Series != book.Series {
		t.Errorf("expected Series %v, got %v", book.Series, results[0].Series)
	}
	if results[0].SeriesIndex == nil || !results[0].SeriesIndex.Valid || !book.SeriesIndex.Valid ||
		!results[0].SeriesIndex.Decimal.Equal(book.SeriesIndex.Decimal) {
		t.Errorf("expected SeriesIndex %v, got %v", book.SeriesIndex, results[0].SeriesIndex)
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
