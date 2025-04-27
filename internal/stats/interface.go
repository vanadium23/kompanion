package stats

import (
	"context"
	"errors"
	"io"
	"time"
)

var ErrEmptyStats = errors.New("empty stats")

type GeneralStats struct {
	TotalReadPages    int
	TotalReadTime     int // in seconds
	AveragePagePerDay int
	AverageTimePerDay int // in seconds
	BookStats         []BookStatsWithTitle
}

type BookStatsWithTitle struct {
	Title string
	BookStats
}

type ReadingStats interface {
	GetBookStats(ctx context.Context, fileHash string) (*BookStats, error)
	GetGeneralStats(ctx context.Context, from, to time.Time) (*GeneralStats, error)
	GetDailyStats(ctx context.Context, from, to time.Time) ([]DailyStats, error)
	Write(ctx context.Context, r io.ReadCloser, deviceName string) error
}
