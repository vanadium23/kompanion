package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vanadium23/kompanion/internal/entity"
	"github.com/vanadium23/kompanion/internal/highlight"
	"github.com/vanadium23/kompanion/pkg/logger"
)

type highlightRoutes struct {
	highlight highlight.Highlight
	l         logger.Interface
}

func newHighlightRoutes(handler *gin.RouterGroup, h highlight.Highlight, l logger.Interface) {
	r := &highlightRoutes{highlight: h, l: l}

	hl := handler.Group("/")
	{
		hl.POST("/highlights", r.syncHighlights)
		hl.GET("/highlights/:document", r.fetchHighlights)
	}
}

type highlightSyncRequest struct {
	Document   string            `json:"document" binding:"required"`
	Title      string            `json:"title"`
	Author     string            `json:"author"`
	Highlights []entity.Highlight `json:"highlights" binding:"required"`
}

func (r *highlightRoutes) syncHighlights(c *gin.Context) {
	var req highlightSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.l.Error(err)
		c.AsciiJSON(http.StatusBadRequest, gin.H{"message": "Bad request", "code": 4000})
		return
	}

	deviceName := c.GetString("device_name")
	synced, err := r.highlight.Sync(c, req.Document, req.Highlights, deviceName)
	if err != nil {
		r.l.Error(err)
		c.AsciiJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error", "code": 5000})
		return
	}

	c.AsciiJSON(http.StatusOK, gin.H{
		"synced": synced,
		"total":  len(req.Highlights),
	})
}

func (r *highlightRoutes) fetchHighlights(c *gin.Context) {
	document := c.Param("document")
	highlights, err := r.highlight.Fetch(c, document)
	if err != nil {
		r.l.Error(err)
		c.AsciiJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error", "code": 5000})
		return
	}

	c.AsciiJSON(http.StatusOK, highlights)
}
