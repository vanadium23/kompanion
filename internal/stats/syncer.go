package stats

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"gitea.chrnv.ru/vanadium23/kompanion/pkg/postgres"

	_ "github.com/mattn/go-sqlite3"
)

type Book struct {
	ID             int
	Title          string
	Authors        string
	Notes          sql.NullInt64 // Use sql.NullInt64 to handle NULL values
	LastOpen       sql.NullInt64 // Use sql.NullInt64 for nullable integers
	Highlights     sql.NullInt64
	Pages          sql.NullInt64
	Series         sql.NullString // Use sql.NullString for nullable strings
	Language       sql.NullString
	MD5            string
	TotalReadTime  sql.NullInt64
	TotalReadPages sql.NullInt64
}

type PageStatData struct {
	MD5        string
	Page       int
	StartTime  int
	Duration   int
	TotalPages int
}

// Function to sync the databases
func SyncDatabases(pathToSQLite string, pg *postgres.Postgres, deviceName string) error {
	sqliteDB, err := sql.Open("sqlite3", pathToSQLite)
	if err != nil {
		fmt.Println(pathToSQLite)
		log.Fatalf("failed to connect to SQLite: %v", err)
		return err
	}
	defer sqliteDB.Close()
	defer os.Remove(pathToSQLite)

	pgDB := pg.Pool

	// Sync the book table
	fmt.Println("Syncing books...")
	err = syncBooks(sqliteDB, pgDB, deviceName)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("error syncing books: %v", err)
	}

	// Sync the page_stat_data table
	fmt.Println("Syncing pages...")
	err = syncPageStatData(sqliteDB, pgDB, deviceName)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("error syncing page_stat_data: %v", err)
	}
	fmt.Println("Fully synced", pathToSQLite)

	return nil
}

func syncBooks(sqliteDB *sql.DB, pgDB postgres.PostgresPool, deviceName string) error {
	// Query all books from the SQLite DB
	rows, err := sqliteDB.Query(`
		SELECT 
			id, title, authors, notes, last_open, highlights, pages, series, language, md5, total_read_time, total_read_pages 
		FROM book
		WHERE title IS NOT NULL AND md5 IS NOT NULL`,
	)
	if err != nil {
		return fmt.Errorf("failed to fetch books from SQLite: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.Authors, &book.Notes, &book.LastOpen, &book.Highlights, &book.Pages, &book.Series, &book.Language, &book.MD5, &book.TotalReadTime, &book.TotalReadPages)
		if err != nil {
			return fmt.Errorf("failed to scan book: %v", err)
		}

		// Perform upsert operation in PostgreSQL
		_, err = pgDB.Exec(context.Background(), `
            INSERT INTO stats_book (koreader_partial_md5, title, authors, notes, last_open, highlights, pages, series, language, total_read_time, total_read_pages, auth_device_name)
            VALUES ($1, $2, $3, $4, to_timestamp($5), $6, $7, $8, $9, $10, $11, $12)
            ON CONFLICT (koreader_partial_md5, auth_device_name) DO UPDATE
            SET title = EXCLUDED.title, 
                authors = EXCLUDED.authors, 
                notes = EXCLUDED.notes, 
                last_open = EXCLUDED.last_open, 
                highlights = EXCLUDED.highlights, 
                pages = EXCLUDED.pages, 
                series = EXCLUDED.series, 
                language = EXCLUDED.language, 
                total_read_time = EXCLUDED.total_read_time, 
                total_read_pages = EXCLUDED.total_read_pages;
        `,
			sanitizeString(book.MD5),
			sanitizeString(book.Title),
			sanitizeString(book.Authors),
			nullableToInterface(book.Notes),          // Handle nullable values
			nullableToInterface(book.LastOpen),       // Handle nullable values
			nullableToInterface(book.Highlights),     // Handle nullable values
			nullableToInterface(book.Pages),          // Handle nullable values
			nullableToInterface(book.Series),         // Handle nullable values
			nullableToInterface(book.Language),       // Handle nullable values
			nullableToInterface(book.TotalReadTime),  // Handle nullable values
			nullableToInterface(book.TotalReadPages), // Handle nullable values,
			deviceName,
		)
		if err != nil {
			return fmt.Errorf("failed to upsert book in PostgreSQL: %v, %v", err, book)
		}
	}
	return nil
}

func nullableToInterface(val interface{}) interface{} {
	switch v := val.(type) {
	case sql.NullInt64:
		if v.Valid {
			return v.Int64
		}
		return nil
	case sql.NullString:
		if v.Valid && v.String != "N/A" {
			return sanitizeString(v.String)
		}
		return sql.NullString{String: "", Valid: false}
	default:
		return val
	}
}

// Helper function to remove null bytes from strings
func sanitizeString(input string) string {
	// Remove any NULL bytes from the string
	return strings.ReplaceAll(input, "\x00", "")
}

func syncPageStatData(sqliteDB *sql.DB, pgDB postgres.PostgresPool, deviceName string) error {
	maxStartTime := 0
	// Query the maximum start time from the page_stat_data table
	pgDB.QueryRow(
		context.Background(),
		`SELECT 
			COALESCE(EXTRACT(EPOCH FROM MAX(start_time))::integer, 0) 
			FROM stats_page_stat_data WHERE auth_device_name = $1`, deviceName).Scan(&maxStartTime)
	fmt.Println("Max start time:", maxStartTime)
	// Query all page stat data from the SQLite DB
	rows, err := sqliteDB.Query(`
		SELECT book.md5, page, start_time, duration, total_pages 
		FROM page_stat_data
		JOIN book ON book.id = page_stat_data.id_book
		WHERE page_stat_data.start_time >= ?
			AND book.title IS NOT NULL AND book.md5 IS NOT NULL`, maxStartTime)
	if err != nil {
		return fmt.Errorf("failed to fetch page stat data from SQLite: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pageData PageStatData
		err := rows.Scan(&pageData.MD5, &pageData.Page, &pageData.StartTime, &pageData.Duration, &pageData.TotalPages)
		if err != nil {
			return fmt.Errorf("failed to scan page stat data: %v", err)
		}

		// Perform upsert operation in PostgreSQL
		_, err = pgDB.Exec(context.Background(), `
            INSERT INTO stats_page_stat_data (koreader_partial_md5, page, start_time, duration, total_pages, auth_device_name)
            VALUES ($1, $2, to_timestamp($3), $4, $5, $6)
            ON CONFLICT (koreader_partial_md5, page, start_time, auth_device_name) DO NOTHING;
        `,
			pageData.MD5, pageData.Page, pageData.StartTime, pageData.Duration, pageData.TotalPages, deviceName)

		if err != nil {
			return fmt.Errorf("failed to upsert page stat data in PostgreSQL: %v", err)
		}
		fmt.Println("Upserted page stat data for", pageData.MD5, "page", pageData.Page)
	}
	return nil
}
