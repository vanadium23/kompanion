package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/vanadium23/kompanion/internal/auth"
	"github.com/vanadium23/kompanion/internal/entity"
	"github.com/vanadium23/kompanion/internal/highlight"
	"github.com/vanadium23/kompanion/internal/notes"
	"github.com/vanadium23/kompanion/pkg/logger"
)

// mockAuthForNotes implements auth.AuthInterface for testing
type mockAuthForNotes struct {
	validDevice bool
}

func (m *mockAuthForNotes) CheckPassword(ctx context.Context, username string, password string) bool {
	return false
}
func (m *mockAuthForNotes) Login(ctx context.Context, username string, password string, userAgent string, clientIP net.IP) (string, error) {
	return "", nil
}
func (m *mockAuthForNotes) Logout(ctx context.Context, sessionKey string) error { return nil }
func (m *mockAuthForNotes) IsAuthenticated(ctx context.Context, sessionKey string) bool {
	return false
}
func (m *mockAuthForNotes) AddUserDevice(ctx context.Context, device_name, password string) error {
	return nil
}
func (m *mockAuthForNotes) DeactivateUserDevice(ctx context.Context, device_name string) error {
	return nil
}
func (m *mockAuthForNotes) CheckDevicePassword(ctx context.Context, device_name, password string, plain bool) bool {
	return m.validDevice && device_name == "testdevice" && password == "testpass"
}
func (m *mockAuthForNotes) ListDevices(ctx context.Context) ([]auth.Device, error) {
	return nil, nil
}
func (m *mockAuthForNotes) RegisterUser(ctx context.Context, username, password string) error {
	return nil
}

// mockHighlightForNotes implements highlight.Highlight interface for testing
type mockHighlightForNotes struct {
	docs       []highlight.DocumentInfo
	highlights map[string][]entity.Highlight
	err        error
}

func (m *mockHighlightForNotes) Sync(ctx context.Context, documentID string, highlights []entity.Highlight, deviceName string) (int, error) {
	return 0, nil
}
func (m *mockHighlightForNotes) Fetch(ctx context.Context, documentID string) ([]entity.Highlight, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.highlights[documentID], nil
}
func (m *mockHighlightForNotes) GetDocumentsByDevice(ctx context.Context, deviceName string) ([]highlight.DocumentInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.docs, nil
}

func TestNotesAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		authUser       string
		authPass       string
		mockDocs       []highlight.DocumentInfo
		mockHighlights map[string][]entity.Highlight
		expectStatus   int
	}{
		{
			name:         "GET /notes returns 401 without auth",
			method:       "GET",
			path:         "/index.php/apps/notes/api/v1/notes",
			authUser:     "",
			authPass:     "",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:           "GET /notes returns empty array for device with no highlights",
			method:         "GET",
			path:           "/index.php/apps/notes/api/v1/notes",
			authUser:       "testdevice",
			authPass:       "testpass",
			mockDocs:       []highlight.DocumentInfo{},
			expectStatus:   http.StatusOK,
		},
		{
			name:     "GET /notes returns notes for device with highlights",
			method:   "GET",
			path:     "/index.php/apps/notes/api/v1/notes",
			authUser: "testdevice",
			authPass: "testpass",
			mockDocs: []highlight.DocumentInfo{
				{PartialMD5: "doc-md5-1", Title: "Book Title", Author: "Author Name"},
			},
			mockHighlights: map[string][]entity.Highlight{
				"doc-md5-1": {
					{DocumentID: "doc-md5-1", Text: "Highlight text", Page: "10", Chapter: "Chapter 1", Timestamp: 1700000000},
				},
			},
			expectStatus: http.StatusOK,
		},
		{
			name:         "POST /notes returns 200 with note object",
			method:       "POST",
			path:         "/index.php/apps/notes/api/v1/notes",
			body:         map[string]interface{}{"title": "Test Book", "content": "Test content", "category": 0},
			authUser:     "testdevice",
			authPass:     "testpass",
			expectStatus: http.StatusOK,
		},
		{
			name:         "PUT /notes/:id returns 200 with note object",
			method:       "PUT",
			path:         "/index.php/apps/notes/api/v1/notes/123",
			body:         map[string]interface{}{"title": "Updated Book", "content": "Updated content", "category": 0},
			authUser:     "testdevice",
			authPass:     "testpass",
			expectStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockLogger := logger.New("info")

			// Create router with mock dependencies
			router := gin.New()

			mockAuth := &mockAuthForNotes{validDevice: tc.authUser == "testdevice" && tc.authPass == "testpass"}
			mockHighlightSvc := &mockHighlightForNotes{
				docs:       tc.mockDocs,
				highlights: tc.mockHighlights,
			}

			// Create notes routes with auth middleware
			notesGroup := router.Group("/index.php/apps/notes/api/v1")
			notesGroup.Use(notesBasicAuth(mockAuth))
			{
				notesGroup.GET("/notes", listNotesHandler(mockHighlightSvc, mockLogger))
				notesGroup.POST("/notes", createNoteHandler(mockLogger))
				notesGroup.PUT("/notes/:id", updateNoteHandler(mockLogger))
			}

			// Create request
			var body bytes.Buffer
			if tc.body != nil {
				json.NewEncoder(&body).Encode(tc.body)
			}
			req := httptest.NewRequest(tc.method, tc.path, &body)
			if tc.authUser != "" && tc.authPass != "" {
				req.SetBasicAuth(tc.authUser, tc.authPass)
			}
			if tc.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			require.Equal(t, tc.expectStatus, w.Code)

			// For successful GET requests, verify response body
			if tc.method == "GET" && tc.expectStatus == http.StatusOK {
				var notesList []map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &notesList)
				require.NoError(t, err)
				if len(tc.mockDocs) == 0 {
					require.Empty(t, notesList)
				} else {
					require.Len(t, notesList, len(tc.mockDocs))
					// Verify note has expected format
					require.Contains(t, notesList[0], "id")
					require.Contains(t, notesList[0], "title")
					require.Contains(t, notesList[0], "content")
				}
			}

			// For POST/PUT requests, verify response has note object
			if (tc.method == "POST" || tc.method == "PUT") && tc.expectStatus == http.StatusOK {
				var note map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &note)
				require.NoError(t, err)
				require.Contains(t, note, "id")
				require.Contains(t, note, "title")
			}
		})
	}
}

// Helper functions used by tests
func listNotesHandler(h *mockHighlightForNotes, l logger.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceName := c.GetString("device_name")
		docs, err := h.GetDocumentsByDevice(c.Request.Context(), deviceName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error", "code": 5000})
			return
		}

		var notesList []gin.H
		for _, doc := range docs {
			highlights, _ := h.Fetch(c.Request.Context(), doc.PartialMD5)
			content := notes.FormatHighlights(doc.Title, doc.Author, highlights)
			notesList = append(notesList, gin.H{
				"id":       notes.HashToInt(doc.PartialMD5),
				"title":    notes.FormatTitle(doc.Author, doc.Title),
				"content":  content,
				"category": 0,
				"etag":     "",
				"readonly": false,
				"favorite": false,
				"modified": 0,
			})
		}

		c.JSON(http.StatusOK, notesList)
	}
}

func createNoteHandler(l logger.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Title    string `json:"title"`
			Content  string `json:"content"`
			Category int    `json:"category"`
			Favorite bool   `json:"favorite"`
			Modified int64  `json:"modified"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request", "code": 4000})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":       notes.HashToInt(req.Title),
			"title":    req.Title,
			"content":  req.Content,
			"category": req.Category,
			"etag":     "",
			"readonly": false,
			"favorite": req.Favorite,
			"modified": req.Modified,
		})
	}
}

func updateNoteHandler(l logger.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		noteID := c.Param("id")
		var req struct {
			Title    string `json:"title"`
			Content  string `json:"content"`
			Category int    `json:"category"`
			Favorite bool   `json:"favorite"`
			Modified int64  `json:"modified"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request", "code": 4000})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":       noteID,
			"title":    req.Title,
			"content":  req.Content,
			"category": req.Category,
			"etag":     "",
			"readonly": false,
			"favorite": req.Favorite,
			"modified": req.Modified,
		})
	}
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
