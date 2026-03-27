package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vanadium23/kompanion/internal/entity"
	"github.com/vanadium23/kompanion/internal/highlights"
	"github.com/vanadium23/kompanion/pkg/logger"
)

type highlightRoutes struct {
	highlight highlights.HighlightSync
	l         logger.Interface
}

func newHighlightRoutes(handler *gin.RouterGroup, h highlights.HighlightSync, l logger.Interface) {
	r := &highlightRoutes{h, l}

	handler.POST("/highlights", r.syncHighlights)
}

func (r *highlightRoutes) syncHighlights(c *gin.Context) {
	var req entity.HighlightSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.l.Error(err)
		c.AsciiJSON(http.StatusBadRequest, gin.H{"message": "Bad request", "code": 4000})
		return
	}

	deviceName := c.GetString("device_name")
	synced, total, err := r.highlight.Sync(c, req, deviceName)
	if err != nil {
		r.l.Error(err)
		c.AsciiJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error", "code": 5000})
		return
	}

	c.AsciiJSON(http.StatusOK, gin.H{"synced": synced, "total": total})
}
