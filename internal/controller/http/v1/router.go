// Package v1 implements routing paths. Each services in own file.
package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/vanadium23/kompanion/internal/auth"
	"github.com/vanadium23/kompanion/internal/highlight"
	"github.com/vanadium23/kompanion/internal/library"
	"github.com/vanadium23/kompanion/internal/sync"
	"github.com/vanadium23/kompanion/pkg/logger"
)

// NewRouter -.
func NewRouter(handler *gin.Engine, l logger.Interface, a auth.AuthInterface, p sync.Progress, shelf library.Shelf, h highlight.Highlight) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	// K8s probe
	handler.GET("/healthcheck", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Prometheus metrics
	handler.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Routers
	newUserRoutes(handler.Group("/"), a, l)

	syncRoutes := handler.Group("/syncs")
	syncRoutes.Use(authDeviceMiddleware(a, l))
	newSyncRoutes(syncRoutes, p, l)

	// Highlight sync routes (uses same device auth as progress sync)
	highlightRoutes := handler.Group("/syncs")
	highlightRoutes.Use(authDeviceMiddleware(a, l))
	newHighlightRoutes(highlightRoutes, h, l)

	// Notes API (Nextcloud Notes compatible, Basic Auth)
	notesRoutes := handler.Group("/index.php/apps/notes/api/v1")
	notesRoutes.Use(notesBasicAuth(a))
	newNotesRoutes(notesRoutes, h, l)
}
