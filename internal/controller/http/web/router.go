package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"time"

	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"
	"github.com/vanadium23/kompanion/internal/auth"
	"github.com/vanadium23/kompanion/internal/library"
	"github.com/vanadium23/kompanion/internal/stats"
	"github.com/vanadium23/kompanion/internal/sync"
	"github.com/vanadium23/kompanion/pkg/logger"
)

func formatDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

func NewRouter(
	handler *gin.Engine,
	l logger.Interface,
	a auth.AuthInterface,
	p sync.Progress,
	shelf library.Shelf,
	stats stats.ReadingStats,
	version string,
) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())
	handler.Use(func(c *gin.Context) {
		c.Set("startTime", time.Now())
	})
	// static files
	handler.Static("/static", "web/static")

	config := goview.DefaultConfig
	config.Root = "web/templates"
	config.DisableCache = gin.IsDebugging()
	config.Funcs = template.FuncMap{
		"formatDuration": formatDuration,
		"json": func(v interface{}) template.JS {
			b, err := json.Marshal(v)
			if err != nil {
				return template.JS("[]")
			}
			return template.JS(b)
		},
		"subtract": func(a, b int) int {
			return a - b
		},
		// https://github.com/go-gitea/gitea/blob/f35850f48ed0bd40ec288e2547ac687a7bf1746c/modules/templates/helper.go#L76
		"LoadTimes": func(startTime time.Time) string {
			return fmt.Sprint(time.Since(startTime).Nanoseconds()/1e6) + "ms"
		},
		"Version": func() string {
			return template.HTMLEscapeString(version)
		},
	}
	handler.HTMLRender = ginview.New(config)

	// Home
	handler.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/books")
	})

	// Login
	authGroup := handler.Group("/auth")
	newAuthRoutes(authGroup, a, l)

	// Product pages
	bookGroup := handler.Group("/books")
	bookGroup.Use(authMiddleware(a))
	newBooksRoutes(bookGroup, shelf, stats, l)

	// Stats pages
	statsGroup := handler.Group("/stats")
	statsGroup.Use(authMiddleware(a))
	newStatsRoutes(statsGroup, stats, l)

	// Device management
	deviceGroup := handler.Group("/devices")
	deviceGroup.Use(authMiddleware(a))
	newDeviceRoutes(deviceGroup, a, l)
}

func passStandartContext(c *gin.Context, data gin.H) gin.H {
	data["isAuthenticated"] = c.GetBool("isAuthenticated")
	data["startTime"] = c.GetTime("startTime")
	return data
}
