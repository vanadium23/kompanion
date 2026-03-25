package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vanadium23/kompanion/internal/auth"
	"github.com/vanadium23/kompanion/internal/highlight"
	"github.com/vanadium23/kompanion/internal/notes"
	"github.com/vanadium23/kompanion/pkg/logger"
)

// NoteResponse represents a note in Nextcloud Notes API format.
type NoteResponse struct {
	ID       int    `json:"id"`
	Etag     string `json:"etag"`
	ReadOnly bool   `json:"readonly"`
	Content  string `json:"content"`
	Title    string `json:"title"`
	Category int    `json:"category"`
	Favorite bool   `json:"favorite"`
	Modified int64  `json:"modified"`
}

type notesRoutes struct {
	highlight highlight.Highlight
	l         logger.Interface
}

// newNotesRoutes creates routes for Nextcloud Notes API compatibility.
func newNotesRoutes(handler *gin.RouterGroup, h highlight.Highlight, l logger.Interface) {
	r := &notesRoutes{highlight: h, l: l}

	notesGroup := handler.Group("/notes")
	{
		notesGroup.GET("", r.listNotes)
		notesGroup.POST("", r.createNote)
		notesGroup.PUT("/:id", r.updateNote)
	}
}

// listNotes returns all notes (one per book with highlights) for the authenticated device.
func (r *notesRoutes) listNotes(c *gin.Context) {
	deviceName := c.GetString("device_name")

	// Get all documents with highlights for this device
	docs, err := r.highlight.GetDocumentsByDevice(c.Request.Context(), deviceName)
	if err != nil {
		r.l.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error", "code": 5000})
		return
	}

	// Build note for each document
	var notesList []NoteResponse
	for _, doc := range docs {
		highlights, err := r.highlight.Fetch(c.Request.Context(), doc.PartialMD5)
		if err != nil {
			r.l.Error(err)
			continue
		}

		content := notes.FormatHighlights(doc.Title, doc.Author, highlights)
		note := NoteResponse{
			ID:       notes.HashToInt(doc.PartialMD5),
			Title:    notes.FormatTitle(doc.Author, doc.Title),
			Content:  content,
			Category: 0,
			Modified: 0,
		}
		notesList = append(notesList, note)
	}

	c.JSON(http.StatusOK, notesList)
}

// createNoteRequest represents the request body for creating/updating a note.
type createNoteRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category int    `json:"category"`
	Favorite bool   `json:"favorite"`
	Modified int64  `json:"modified"`
}

// createNote acknowledges a note creation (actual data flows via /syncs/highlights).
func (r *notesRoutes) createNote(c *gin.Context) {
	var req createNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.l.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request", "code": 4000})
		return
	}

	// Acknowledge the note creation - highlights are stored via /syncs/highlights
	note := NoteResponse{
		ID:       notes.HashToInt(req.Title),
		Title:    req.Title,
		Content:  req.Content,
		Category: req.Category,
		Favorite: req.Favorite,
		Modified: req.Modified,
	}

	c.JSON(http.StatusOK, note)
}

// updateNote acknowledges a note update (actual data flows via /syncs/highlights).
func (r *notesRoutes) updateNote(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, err := strconv.Atoi(noteIDStr)
	if err != nil {
		r.l.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid note ID", "code": 4000})
		return
	}

	var req createNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.l.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request", "code": 4000})
		return
	}

	// Acknowledge the note update - highlights are stored via /syncs/highlights
	note := NoteResponse{
		ID:       noteID,
		Title:    req.Title,
		Content:  req.Content,
		Category: req.Category,
		Favorite: req.Favorite,
		Modified: req.Modified,
	}

	c.JSON(http.StatusOK, note)
}

// notesBasicAuth provides Basic Auth middleware for Notes API using device credentials.
func notesBasicAuth(a auth.AuthInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", `Basic realm="KOmpanion Notes"`)
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized", "code": 2001})
			c.Abort()
			return
		}
		if !a.CheckDevicePassword(c.Request.Context(), username, password, true) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized", "code": 2001})
			c.Abort()
			return
		}
		c.Set("device_name", username)
		c.Next()
	}
}
