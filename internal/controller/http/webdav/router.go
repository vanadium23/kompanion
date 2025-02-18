package webdav

import (
	"net/http"

	"gitea.chrnv.ru/vanadium23/kompanion/internal/auth"
	"gitea.chrnv.ru/vanadium23/kompanion/internal/stats"
	"gitea.chrnv.ru/vanadium23/kompanion/pkg/logger"
	"github.com/gin-gonic/gin"
)

func NewRouter(
	handler *gin.Engine,
	a auth.AuthInterface,
	l logger.Interface,
	rs stats.ReadingStats,
	dirPath string,
) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	h := handler.Group("/webdav")
	h.Use(basicAuth(a))
	h.Handle("PROPFIND", "/", func(c *gin.Context) {
		// Static response for PROPFIND
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<D:multistatus
			xmlns:D="DAV:">
		</D:multistatus>`
		c.Header("Content-Type", "application/xml")
		c.String(http.StatusMultiStatus, response)
	})
	h.PUT("/statistics.sqlite3", func(c *gin.Context) {
		device := c.GetString("device_name")
		err := rs.Write(c.Request.Context(), c.Request.Body, device)
		if err != nil {
			l.Info("error writing statistics", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "error writing statistics"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "statistics updated"})
	})
}

func basicAuth(auth auth.AuthInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", `Basic realm="KOmpanion WebDav"`)
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized", "code": 2001})
			c.Abort()
			return
		}
		if !auth.CheckDevicePassword(c.Request.Context(), username, password, true) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized", "code": 2001})
			c.Abort()
			return
		}
		c.Set("device_name", username)
		c.Next()
	}
}
