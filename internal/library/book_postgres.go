package library

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/vanadium23/kompanion/internal/entity"
	"github.com/vanadium23/kompanion/pkg/postgres"
)

// BookDatabaseRepo -.
type BookDatabaseRepo struct {
	*postgres.Postgres
}

// New -.
func NewBookDatabaseRepo(pg *postgres.Postgres) *BookDatabaseRepo {
	return &BookDatabaseRepo{pg}
}

// Store -. only insert in database
func (bdr *BookDatabaseRepo) Store(ctx context.Context, book entity.Book) error {
	query := `
		INSERT INTO library_book (id, title, author, publisher, year, created_at, updated_at, isbn, storage_file_path, koreader_partial_md5, storage_cover_path, series, series_index, summary)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	args := []interface{}{
		book.ID, book.Title, book.Author, book.Publisher, book.Year,
		book.CreatedAt, book.UpdatedAt, book.ISBN, book.FilePath,
		book.DocumentID, book.CoverPath, book.Series, book.SeriesIndex, book.Description,
	}

	_, err := bdr.Pool.Exec(ctx, query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return fmt.Errorf("BookDatabaseRepo - Store - r.Pool.Exec: %w", entity.ErrBookAlreadyExists)
		}
		return fmt.Errorf("BookDatabaseRepo - Store - r.Pool.Exec: %w", err)
	}

	return nil
}

// Update -. only update in database
func (bdr *BookDatabaseRepo) Update(ctx context.Context, book entity.Book) error {
	query := `
		UPDATE library_book
		SET title = $1,
			author = $2,
			publisher = $3,
			year = $4,
			updated_at = $5,
			isbn = $6,
			series = $7,
			series_index = $8,
			summary = $9
		WHERE id = $10
	`
	args := []interface{}{
		book.Title, book.Author, book.Publisher, book.Year,
		book.UpdatedAt, book.ISBN, book.Series, book.SeriesIndex, book.Description, book.ID,
	}

	rows, err := bdr.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("BookDatabaseRepo - Update - r.Pool.Exec: %w", err)
	}
	if rows.RowsAffected() == 0 {
		return fmt.Errorf("BookDatabaseRepo - Update - no rows affected")
	}
	return nil
}

// List -. only select from database
func (bdr *BookDatabaseRepo) List(ctx context.Context,
	sortBy, sortOrder string,
	page, perPage int,
) ([]entity.Book, error) {
	switch sortOrder {
	case "asc", "desc":
	default:
		sortOrder = "desc"
	}

	switch sortBy {
	case "title", "author", "publisher", "year", "created_at", "updated_at", "isbn":
	default:
		sortBy = "created_at"
	}

	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 25
	}

	// Use limit and offset for pagination, because we don't have a lot of books
	// (yes, it's not the best way to do pagination)
	query := fmt.Sprintf(`
		SELECT
			id, title, author, publisher, year, created_at, updated_at, isbn, storage_file_path, koreader_partial_md5, storage_cover_path, series, series_index, summary
		FROM library_book
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	`, sortBy, sortOrder, perPage, (page-1)*perPage)

	rows, err := bdr.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("BookDatabaseRepo - List - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	books := make([]entity.Book, 0)
	for rows.Next() {
		var book entity.Book
		var seriesIndex decimal.NullDecimal
		var summary sql.NullString
		err = rows.Scan(&book.ID, &book.Title, &book.Author, &book.Publisher, &book.Year, &book.CreatedAt, &book.UpdatedAt, &book.ISBN, &book.FilePath, &book.DocumentID, &book.CoverPath, &book.Series, &seriesIndex, &summary)
		if err != nil {
			return nil, fmt.Errorf("BookDatabaseRepo - List - rows.Scan: %w", err)
		}
		if seriesIndex.Valid {
			book.SeriesIndex = &seriesIndex
		}
		if summary.Valid {
			book.Description = summary.String
		}
		books = append(books, book)
	}

	return books, nil
}

// Get -. only select from database
func (bdr *BookDatabaseRepo) GetById(ctx context.Context, id string) (entity.Book, error) {
	query := `
		SELECT id, title, author, publisher, year, created_at, updated_at, isbn, storage_file_path, koreader_partial_md5, storage_cover_path, series, series_index, summary
		FROM library_book
		WHERE id = $1
	`
	args := []interface{}{id}

	row := bdr.Pool.QueryRow(ctx, query, args...)
	var book entity.Book
	var seriesIndex decimal.NullDecimal
	var summary sql.NullString
	err := row.Scan(&book.ID, &book.Title, &book.Author, &book.Publisher, &book.Year, &book.CreatedAt, &book.UpdatedAt, &book.ISBN, &book.FilePath, &book.DocumentID, &book.CoverPath, &book.Series, &seriesIndex, &summary)
	if err != nil {
		return entity.Book{}, fmt.Errorf("BookDatabaseRepo - Get - r.Pool.QueryRow: %w", err)
	}
	if seriesIndex.Valid {
		book.SeriesIndex = &seriesIndex
	}
	if summary.Valid {
		book.Description = summary.String
	}

	return book, nil
}

// GetByFileHash -. only select from database
func (bdr *BookDatabaseRepo) GetByFileHash(ctx context.Context, fileHash string) (entity.Book, error) {
	query := `
		SELECT id, title, author, publisher, year, created_at, updated_at, isbn, storage_file_path, koreader_partial_md5, storage_cover_path, series, series_index, summary
		FROM library_book
		WHERE koreader_partial_md5 = $1
	`
	args := []interface{}{fileHash}

	row := bdr.Pool.QueryRow(ctx, query, args...)
	var book entity.Book
	var seriesIndex decimal.NullDecimal
	var summary sql.NullString
	err := row.Scan(&book.ID, &book.Title, &book.Author, &book.Publisher, &book.Year, &book.CreatedAt, &book.UpdatedAt, &book.ISBN, &book.FilePath, &book.DocumentID, &book.CoverPath, &book.Series, &seriesIndex, &summary)
	if err != nil {
		return entity.Book{}, fmt.Errorf("BookDatabaseRepo - GetByFileHash - r.Pool.QueryRow: %w", err)
	}
	if seriesIndex.Valid {
		book.SeriesIndex = &seriesIndex
	}
	if summary.Valid {
		book.Description = summary.String
	}

	return book, nil
}

// Count -. only select from database
func (bdr *BookDatabaseRepo) Count(ctx context.Context) (int, error) {
	sql := `SELECT count(*) FROM library_book`

	row := bdr.Pool.QueryRow(ctx, sql)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("BookDatabaseRepo - Count - r.Pool.QueryRow: %w", err)
	}

	return count, nil
}

// Search -. search books with ILIKE and sorting
func (bdr *BookDatabaseRepo) Search(ctx context.Context, query entity.SearchQuery) ([]entity.Book, error) {
	// Validate and set defaults for sort order
	sortOrder := query.SortOrder
	switch sortOrder {
	case "asc", "desc":
	default:
		sortOrder = "asc"
	}

	// Validate and set defaults for sort field
	sortBy := query.SortBy
	switch sortBy {
	case "title", "author", "series", "created_at":
	default:
		sortBy = "title"
	}

	// Set defaults for pagination
	page := query.Page
	if page <= 0 {
		page = 1
	}
	limit := query.Limit
	if limit <= 0 || limit > 100 {
		limit = 25
	}

	// Build WHERE clause for search
	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	if query.Search != "" {
		// Split search into words and create AND conditions
		words := strings.Fields(query.Search)
		conditions := make([]string, 0, len(words))

		for _, word := range words {
			likePattern := "%" + strings.ToLower(word) + "%"
			conditions = append(conditions, fmt.Sprintf(
				"(LOWER(title) LIKE $%d OR LOWER(author) LIKE $%d OR LOWER(COALESCE(series, '')) LIKE $%d OR LOWER(COALESCE(summary, '')) LIKE $%d)",
				argIndex, argIndex, argIndex, argIndex,
			))
			args = append(args, likePattern)
			argIndex++
		}

		if len(conditions) > 0 {
			whereClause = "WHERE " + strings.Join(conditions, " AND ")
		}
	}

	// Build ORDER BY clause with NULL series sorted last
	orderByClause := ""
	if sortBy == "series" {
		// Sort by series, with NULL values last
		if sortOrder == "asc" {
			orderByClause = "ORDER BY series IS NULL, series ASC"
		} else {
			orderByClause = "ORDER BY series IS NULL, series DESC"
		}
	} else {
		orderByClause = fmt.Sprintf("ORDER BY %s %s", sortBy, sortOrder)
	}

	// Build complete query
	baseQuery := `
		SELECT
			id, title, author, publisher, year, created_at, updated_at, isbn, storage_file_path, koreader_partial_md5, storage_cover_path, series, series_index, summary
		FROM library_book
	`
	queryStr := fmt.Sprintf("%s %s %s LIMIT %d OFFSET %d",
		baseQuery, whereClause, orderByClause, limit, (page-1)*limit)

	rows, err := bdr.Pool.Query(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("BookDatabaseRepo - Search - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	books := make([]entity.Book, 0)
	for rows.Next() {
		var book entity.Book
		var seriesIndex decimal.NullDecimal
		var summary sql.NullString
		err = rows.Scan(&book.ID, &book.Title, &book.Author, &book.Publisher, &book.Year, &book.CreatedAt, &book.UpdatedAt, &book.ISBN, &book.FilePath, &book.DocumentID, &book.CoverPath, &book.Series, &seriesIndex, &summary)
		if err != nil {
			return nil, fmt.Errorf("BookDatabaseRepo - Search - rows.Scan: %w", err)
		}
		if seriesIndex.Valid {
			book.SeriesIndex = &seriesIndex
		}
		if summary.Valid {
			book.Description = summary.String
		}
		books = append(books, book)
	}

	return books, nil
}

// SearchCount -. count search results
func (bdr *BookDatabaseRepo) SearchCount(ctx context.Context, query entity.SearchQuery) (int, error) {
	// Build WHERE clause for search (same as Search)
	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	if query.Search != "" {
		// Split search into words and create AND conditions
		words := strings.Fields(query.Search)
		conditions := make([]string, 0, len(words))

		for _, word := range words {
			likePattern := "%" + strings.ToLower(word) + "%"
			conditions = append(conditions, fmt.Sprintf(
				"(LOWER(title) LIKE $%d OR LOWER(author) LIKE $%d OR LOWER(COALESCE(series, '')) LIKE $%d OR LOWER(COALESCE(summary, '')) LIKE $%d)",
				argIndex, argIndex, argIndex, argIndex,
			))
			args = append(args, likePattern)
			argIndex++
		}

		if len(conditions) > 0 {
			whereClause = "WHERE " + strings.Join(conditions, " AND ")
		}
	}

	queryStr := fmt.Sprintf("SELECT count(*) FROM library_book %s", whereClause)

	row := bdr.Pool.QueryRow(ctx, queryStr, args...)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("BookDatabaseRepo - SearchCount - r.Pool.QueryRow: %w", err)
	}

	return count, nil
}
