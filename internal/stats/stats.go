package stats

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"gitea.chrnv.ru/vanadium23/kompanion/internal/storage"
	"gitea.chrnv.ru/vanadium23/kompanion/pkg/postgres"
)

const KOReaderFile = "statistics.sqlite3"

// KOReaderStats implements ReadingStats interface
type KOReaderStats struct {
	rw sync.RWMutex
	st storage.Storage
	pg *postgres.Postgres
}

func NewKOReaderStats(st storage.Storage, pg *postgres.Postgres) *KOReaderStats {
	return &KOReaderStats{st: st, rw: sync.RWMutex{}, pg: pg}
}

func (s *KOReaderStats) Read(ctx context.Context) (*os.File, error) {
	s.rw.RLock()
	stats, err := s.st.Read(ctx, KOReaderFile)
	s.rw.RUnlock()
	if errors.Is(err, storage.ErrNotFound) {
		return nil, ErrEmptyStats
	}
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (s *KOReaderStats) Write(ctx context.Context, r io.ReadCloser, deviceName string) error {
	// make by temp files
	tempFile, err := os.CreateTemp("", fmt.Sprintf("%s-", deviceName))
	if err != nil {
		return err
	}
	filepath := tempFile.Name()
	defer tempFile.Close()

	_, err = io.Copy(tempFile, r)
	if err != nil {
		return err
	}

	s.rw.Lock()
	err = s.st.Write(ctx, tempFile.Name(), KOReaderFile)
	s.rw.Unlock()
	if err != nil {
		return err
	}
	go SyncDatabases(filepath, s.pg, deviceName)
	return nil
}

type BookStats struct {
	TotalReadPages     int
	TotalReadTime      int // in seconds
	AverageTimePerPage int // in seconds
	TotalReadDays      int
}

func (s *KOReaderStats) GetBookStats(ctx context.Context, fileHash string) (*BookStats, error) {
	query := `
		WITH daily_reads AS (
			SELECT DISTINCT DATE(start_time) as read_date
			FROM stats_page_stat_data
			WHERE koreader_partial_md5 = $1
		)
		SELECT 
			COUNT(DISTINCT page) as total_read_pages,
			SUM(duration) as total_read_time,
			COUNT(DISTINCT DATE(start_time)) as total_read_days
		FROM stats_page_stat_data
		WHERE koreader_partial_md5 = $1
	`

	var stats BookStats
	err := s.pg.Pool.QueryRow(ctx, query, fileHash).Scan(
		&stats.TotalReadPages,
		&stats.TotalReadTime,
		&stats.TotalReadDays,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get book stats: %w", err)
	}

	if stats.TotalReadPages > 0 {
		stats.AverageTimePerPage = stats.TotalReadTime / stats.TotalReadPages
	}

	return &stats, nil
}

func (s *KOReaderStats) GetGeneralStats(ctx context.Context, from, to time.Time) (*GeneralStats, error) {
	var stats GeneralStats

	// Get per book statistics
	bookQuery := `
		SELECT 
			b.title,
			COUNT(DISTINCT kpsd.page) as total_read_pages,
			SUM(kpsd.duration) as total_read_time,
			COUNT(DISTINCT DATE(kpsd.start_time)) as total_read_days
		FROM stats_page_stat_data kpsd
		JOIN stats_book b ON b.koreader_partial_md5 = kpsd.koreader_partial_md5 AND b.auth_device_name = kpsd.auth_device_name
		WHERE kpsd.start_time BETWEEN $1 AND $2
		GROUP BY b.title, b.koreader_partial_md5
		HAVING COUNT(DISTINCT kpsd.page) > 0
	`

	rows, err := s.pg.Pool.Query(ctx, bookQuery, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get book stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var bookStat BookStatsWithTitle
		err := rows.Scan(
			&bookStat.Title,
			&bookStat.TotalReadPages,
			&bookStat.TotalReadTime,
			&bookStat.TotalReadDays,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan book stats: %w", err)
		}

		if bookStat.TotalReadPages > 0 {
			bookStat.AverageTimePerPage = int(bookStat.TotalReadTime / bookStat.TotalReadPages)
		}

		// Add to totals
		stats.TotalReadPages += bookStat.TotalReadPages
		stats.TotalReadTime += bookStat.TotalReadTime
		stats.BookStats = append(stats.BookStats, bookStat)
	}

	// Calculate days between dates for averages
	days := to.Sub(from).Hours() / 24
	if days > 0 {
		stats.AveragePagePerDay = int(float64(stats.TotalReadPages)/days + 0.5)
		stats.AverageTimePerDay = int(float64(stats.TotalReadTime)/days + 0.5)
	}

	return &stats, nil
}

type DailyStats struct {
	Date               time.Time
	PageCount          int
	AvgDurationPerPage float64
}

func (s *KOReaderStats) GetDailyStats(ctx context.Context, from, to time.Time) ([]DailyStats, error) {
	query := `
		WITH RECURSIVE dates AS (
			SELECT date_trunc('day', $1::timestamp)::date as date
			UNION ALL
			SELECT (date + interval '1 day')::date
			FROM dates
			WHERE date < date_trunc('day', $2::timestamp)::date
		)
		SELECT 
			d.date,
			COALESCE(COUNT(DISTINCT kpsd.page), 0) as pages_read,
			CASE 
				WHEN COUNT(DISTINCT kpsd.page) > 0 THEN COALESCE(AVG(kpsd.duration), 0)
				ELSE 0 
			END as avg_duration_per_page
		FROM dates d
		LEFT JOIN stats_page_stat_data kpsd 
			ON date_trunc('day', kpsd.start_time)::date = d.date
			AND kpsd.start_time BETWEEN $1 AND $2
		GROUP BY d.date
		ORDER BY d.date;
	`

	rows, err := s.pg.Pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}
	defer rows.Close()

	var stats []DailyStats
	for rows.Next() {
		var stat DailyStats
		err := rows.Scan(&stat.Date, &stat.PageCount, &stat.AvgDurationPerPage)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily stats: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}
