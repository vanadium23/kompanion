package library

import (
	"context"
	"os"

	"github.com/vanadium23/kompanion/internal/entity"
)

type (
	// Shelf -.
	Shelf interface {
		StoreBook(ctx context.Context, tempFile *os.File, uploadedFilename string) (entity.Book, error)
		ListBooks(ctx context.Context,
			sortBy, sortOrder string,
			page, perPage int,
		) (PaginatedBookList, error)
		SearchBooks(ctx context.Context, query string, page, perPage int) (PaginatedBookList, error)
		ViewBook(ctx context.Context, bookID string) (entity.Book, error)
		DownloadBook(ctx context.Context, bookID string) (entity.Book, *os.File, error)
		UpdateBookMetadata(ctx context.Context, bookID string, metadata entity.Book) (entity.Book, error)
		ViewCover(ctx context.Context, bookID string) (*os.File, error)
		ListSeries(ctx context.Context, page, perPage int) (PaginatedSeriesList, error)
		ListBooksBySeries(ctx context.Context, series string, page, perPage int) (PaginatedBookList, error)
		ListAuthors(ctx context.Context, page, perPage int) (PaginatedAuthorList, error)
		ListBooksByAuthor(ctx context.Context, author string, page, perPage int) (PaginatedBookList, error)
	}

	// BookRepo -.
	BookRepo interface {
		Store(context.Context, entity.Book) error
		List(ctx context.Context,
			sortBy, sortOrder string,
			page, perPage int,
		) ([]entity.Book, error)
		Search(ctx context.Context, query string, page, perPage int) ([]entity.Book, error)
		Count(ctx context.Context) (int, error)
		CountSearch(ctx context.Context, query string) (int, error)
		GetById(context.Context, string) (entity.Book, error)
		GetByFileHash(context.Context, string) (entity.Book, error)
		Update(context.Context, entity.Book) error
		ListSeries(ctx context.Context, page, perPage int) ([]string, error)
		CountSeries(ctx context.Context) (int, error)
		ListBooksBySeries(ctx context.Context, series string, page, perPage int) ([]entity.Book, error)
		CountBooksBySeries(ctx context.Context, series string) (int, error)
		ListAuthors(ctx context.Context, page, perPage int) ([]string, error)
		CountAuthors(ctx context.Context) (int, error)
		ListBooksByAuthor(ctx context.Context, author string, page, perPage int) ([]entity.Book, error)
		CountBooksByAuthor(ctx context.Context, author string) (int, error)
	}
)
