package web

import (
	"bytes"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vanadium23/kompanion/internal/stats"
	"github.com/vanadium23/kompanion/pkg/logger"
	"github.com/wcharczuk/go-chart/v2"
	charts "github.com/wcharczuk/go-chart/v2"
)

func generateDailyStatsChart(stats []stats.DailyStats) ([]byte, error) {
	xValues := make([]float64, len(stats))
	yPagesValues := make([]float64, len(stats))
	yDurationValues := make([]float64, len(stats))

	for i, stat := range stats {
		xValues[i] = float64(stat.Date.Unix())
		yPagesValues[i] = float64(stat.PageCount)
		yDurationValues[i] = float64(int(stat.AvgDurationPerPage))
	}

	// Find max values for both Y axes
	maxPages := 0.0
	maxDuration := 0.0
	for i := range stats {
		if yPagesValues[i] > maxPages {
			maxPages = yPagesValues[i]
		}
		if yDurationValues[i] > maxDuration {
			maxDuration = yDurationValues[i]
		}
	}
	// Add 10% padding to max values
	maxPages = maxPages * 1.1
	maxDuration = maxDuration * 1.1

	// Create the chart
	graph := charts.Chart{
		Title: "",
		Background: charts.Style{
			Padding: charts.Box{
				Top:    20,
				Left:   50,
				Right:  50,
				Bottom: 40,
			},
		},
		XAxis: charts.XAxis{
			ValueFormatter: func(v interface{}) string {
				if ts, ok := v.(float64); ok {
					return time.Unix(int64(ts), 0).Format("2006-01-02")
				}
				return ""
			},
		},
		YAxis: charts.YAxis{
			Name: "Pages Read",
			ValueFormatter: func(v interface{}) string {
				if value, ok := v.(float64); ok {
					return fmt.Sprintf("%d", int(value))
				}
				return ""
			},
			NameStyle: charts.Style{
				FontColor: chart.GetDefaultColor(0),
			},
			Range: &charts.ContinuousRange{
				Min: 0.0,
				Max: maxPages,
			},
		},
		Series: []charts.Series{
			charts.ContinuousSeries{
				Name: "Pages Read",
				Style: charts.Style{
					FillColor:   charts.GetDefaultColor(0).WithAlpha(180),
					StrokeColor: charts.GetDefaultColor(0),
					StrokeWidth: 1,
				},
				XValues: xValues,
				YValues: yPagesValues,
			},
			charts.ContinuousSeries{
				Name: "Average Time per Page",
				Style: charts.Style{
					StrokeColor: charts.GetDefaultColor(1),
					StrokeWidth: 2,
				},
				XValues: xValues,
				YValues: yDurationValues,
				YAxis:   charts.YAxisSecondary,
			},
		},
		YAxisSecondary: charts.YAxis{
			Name: "Seconds per Page",
			ValueFormatter: func(v interface{}) string {
				if value, ok := v.(float64); ok {
					return fmt.Sprintf("%.1fs", value)
				}
				return ""
			},
			Range: &charts.ContinuousRange{
				Min: 0.0,
				Max: maxDuration,
			},
			NameStyle: charts.Style{
				FontColor: chart.GetDefaultColor(1),
			},
		},
		Width:  800,
		Height: 400,
	}

	buffer := &bytes.Buffer{}
	err := graph.Render(charts.PNG, buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to render chart: %w", err)
	}

	return buffer.Bytes(), nil
}

func newStatsRoutes(handler *gin.RouterGroup, stats stats.ReadingStats, l logger.Interface) {
	handler.GET("/", func(c *gin.Context) {
		// Get date range from query params, default: a month ago to current day
		now := time.Now()
		from := now.AddDate(0, -1, 0)
		from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.Local)
		to := from.AddDate(0, 1, 1).Add(-time.Second)

		// Parse from and to dates if provided
		if fromStr := c.Query("from"); fromStr != "" {
			if parsedFrom, err := time.Parse("2006-01-02", fromStr); err == nil {
				from = parsedFrom
			}
		}
		if toStr := c.Query("to"); toStr != "" {
			if parsedTo, err := time.Parse("2006-01-02", toStr); err == nil {
				to = parsedTo.Add(24*time.Hour - time.Second)
			}
		}

		generalStats, err := stats.GetGeneralStats(c.Request.Context(), from, to)
		if err != nil {
			l.Error(err, "failed to get general stats")
			c.HTML(500, "error", passStandartContext(c, gin.H{
				"error": err,
			}))
			return
		}

		c.HTML(200, "stats", passStandartContext(c, gin.H{
			"from":  from.Format("2006-01-02"),
			"to":    to.Format("2006-01-02"),
			"stats": generalStats,
		}))
	})

	handler.GET("/chart", func(c *gin.Context) {
		// Get date range from query params, default: a month ago to current day
		now := time.Now()
		from := now.AddDate(0, -1, 0)
		from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.Local)
		to := from.AddDate(0, 1, 1).Add(-time.Second)

		// Parse from and to dates if provided
		if fromStr := c.Query("from"); fromStr != "" {
			if parsedFrom, err := time.Parse("2006-01-02", fromStr); err == nil {
				from = parsedFrom
			}
		}
		if toStr := c.Query("to"); toStr != "" {
			if parsedTo, err := time.Parse("2006-01-02", toStr); err == nil {
				to = parsedTo.Add(24*time.Hour - time.Second)
			}
		}

		dailyStats, err := stats.GetDailyStats(c.Request.Context(), from, to)
		if err != nil {
			l.Error(err, "failed to get daily stats")
			c.Status(500)
			return
		}

		chartBytes, err := generateDailyStatsChart(dailyStats)
		if err != nil {
			l.Error(err, "failed to generate chart")
			c.Status(500)
			return
		}

		c.Header("Content-Type", "image/png")
		c.Data(200, "image/png", chartBytes)
	})
}
