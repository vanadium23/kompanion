// Package v1 implements routing paths. Each services in own file.
package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gitea.chrnv.ru/vanadium23/kompanion/internal/auth"
	"gitea.chrnv.ru/vanadium23/kompanion/internal/library"
	"gitea.chrnv.ru/vanadium23/kompanion/internal/sync"
	"gitea.chrnv.ru/vanadium23/kompanion/pkg/logger"
)

// NewRouter -.
func NewRouter(handler *gin.Engine, l logger.Interface, a auth.AuthInterface, p sync.Progress, shelf library.Shelf) {
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
}
