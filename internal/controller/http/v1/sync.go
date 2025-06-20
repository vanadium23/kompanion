package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vanadium23/kompanion/internal/entity"
	"github.com/vanadium23/kompanion/internal/sync"
	"github.com/vanadium23/kompanion/pkg/logger"
)

type syncRoutes struct {
	progress sync.Progress
	l        logger.Interface
}

func newSyncRoutes(handler *gin.RouterGroup, p sync.Progress, l logger.Interface) {
	r := &syncRoutes{p, l}

	h := handler.Group("/")
	{
		h.PUT("/progress", r.updateProgress)
		h.GET("/progress/:document", r.fetchProgress)
	}
}

func (r *syncRoutes) updateProgress(c *gin.Context) {
	var doc entity.Progress
	if err := c.ShouldBindJSON(&doc); err != nil {
		r.l.Error(err)
		c.AsciiJSON(http.StatusBadRequest, gin.H{"message": "Bad request", "code": 4000})
		return
	}

	doc.AuthDeviceName = c.GetString("device_name")
	savedDoc, err := r.progress.Sync(c, doc)
	if err != nil {
		r.l.Error(err)
		c.AsciiJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error", "code": 5000})
		return
	}

	c.AsciiJSON(http.StatusOK, gin.H{"timestamp": savedDoc.Timestamp, "document": savedDoc})
}

func (r *syncRoutes) fetchProgress(c *gin.Context) {
	doc, err := r.progress.Fetch(c, c.Param("document"))
	if err != nil {
		r.l.Error(err)
		c.AsciiJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error", "code": 5000})
		return
	}

	c.AsciiJSON(http.StatusOK, doc)
}
