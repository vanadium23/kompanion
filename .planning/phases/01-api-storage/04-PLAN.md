---
phase: 01-api-storage
plan: 04
type: execute
wave: 3
depends_on:
  - 03
files_modified:
  - internal/controller/http/v1/highlight.go
  - internal/controller/http/v1/router.go
  - internal/app/app.go
autonomous: true
requirements:
  - API-01
  - API-02
  - API-03
  - API-04
  - SYNC-03

must_haves:
  truths:
    - "POST /syncs/highlights accepts array of highlights and returns synced count"
    - "Device authentication via authDeviceMiddleware is applied to highlight routes"
    - "Both annotation and legacy KOReader data models work (note field handling)"
  artifacts:
    - path: "internal/controller/http/v1/highlight.go"
      provides: "HTTP handler for highlight sync"
      min_lines: 60
      contains: "func newHighlightRoutes"
    - path: "internal/controller/http/v1/router.go"
      provides: "Router configuration"
      contains: "newHighlightRoutes"
    - path: "internal/app/app.go"
      provides: "Dependency wiring"
      contains: "highlight"
  key_links:
    - from: "internal/controller/http/v1/highlight.go"
      to: "internal/highlight/sync.go"
      via: "Highlight interface"
      pattern: "r.highlight.Sync"
    - from: "internal/controller/http/v1/router.go"
      to: "internal/controller/http/v1/users.go"
      via: "authDeviceMiddleware"
      pattern: "authDeviceMiddleware"
---

<objective>
Implement HTTP handler for highlight sync API and wire dependencies into the application.

Purpose: Expose highlight sync functionality via HTTP POST endpoint with device authentication.
Output: HTTP routes, handler, and dependency wiring in app.go.
</objective>

<execution_context>
@~/.claude/get-shit-done/workflows/execute-plan.md
@~/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/01-api-storage/01-RESEARCH.md
</context>

<interfaces>

From internal/controller/http/v1/sync.go (pattern to follow):
```go
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
    // ...
    c.AsciiJSON(http.StatusOK, gin.H{"timestamp": savedDoc.Timestamp, "document": savedDoc})
}
```

From internal/controller/http/v1/router.go (pattern to follow):
```go
func NewRouter(handler *gin.Engine, l logger.Interface, a auth.AuthInterface, p sync.Progress, shelf library.Shelf) {
    // ...
    syncRoutes := handler.Group("/syncs")
    syncRoutes.Use(authDeviceMiddleware(a, l))
    newSyncRoutes(syncRoutes, p, l)
}
```

From internal/controller/http/v1/users.go (auth middleware pattern):
```go
func authDeviceMiddleware(auth auth.AuthInterface, l logger.Interface) gin.HandlerFunc {
    return func(c *gin.Context) {
        username := c.GetHeader("x-auth-user")
        hashed_password := c.GetHeader("x-auth-key")
        // ...
        c.Set("device_name", username)
        c.Next()
    }
}
```

From internal/app/app.go (dependency wiring pattern):
```go
progress := sync.NewProgressSync(sync.NewProgressDatabaseRepo(pg))
// ...
v1.NewRouter(handler, l, authService, progress, shelf)
```

</interfaces>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Create highlight HTTP handler</name>
  <files>internal/controller/http/v1/highlight.go</files>
  <read_first>
    - internal/controller/http/v1/sync.go (handler pattern to follow)
    - internal/controller/http/v1/users.go (authDeviceMiddleware reference)
    - internal/highlight/interfaces.go (Highlight interface)
    - internal/entity/highlight.go (Highlight entity with JSON tags)
    - .planning/phases/01-api-storage/01-RESEARCH.md (research with handler example)
  </read_first>
  <behavior>
    - Test 1: syncHighlights binds JSON request with document and highlights array
    - Test 2: syncHighlights returns 400 for invalid JSON
    - Test 3: syncHighlights calls highlight.Sync with documentID, highlights, deviceName
    - Test 4: syncHighlights returns 200 with {"synced": N, "total": M}
    - Test 5: syncHighlights extracts device_name from context (set by middleware)
    - Test 6: fetchHighlights calls highlight.Fetch with document parameter
    - Test 7: fetchHighlights returns 200 with highlights array
  </behavior>
  <action>
Create `internal/controller/http/v1/highlight.go`:

```go
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
```

Key points:
- POST endpoint at /highlights (API-01)
- Accepts array of highlights (API-02)
- Returns synced and total counts (API-04)
- Uses c.GetString("device_name") from authDeviceMiddleware (API-03)
- Request struct includes title/author for future use but not required
  </action>
  <verify>
    <automated>go build ./internal/controller/http/v1/...</automated>
  </verify>
  <acceptance_criteria>
    - File internal/controller/http/v1/highlight.go exists
    - File contains "type highlightRoutes struct"
    - File contains "func newHighlightRoutes"
    - File contains "hl.POST(\"/highlights\", r.syncHighlights)"
    - File contains "hl.GET(\"/highlights/:document\", r.fetchHighlights)"
    - File contains "type highlightSyncRequest struct"
    - highlightSyncRequest has `Document string \`json:"document" binding:"required"\``
    - highlightSyncRequest has `Highlights []entity.Highlight \`json:"highlights" binding:"required"\``
    - syncHighlights extracts device_name: "c.GetString(\"device_name\")"
    - syncHighlights returns gin.H{"synced": ..., "total": ...}
    - Command `go build ./internal/controller/http/v1/...` exits 0
  </acceptance_criteria>
  <done>HTTP handler created with POST /highlights endpoint satisfying API-01, API-02, API-03, API-04</done>
</task>

<task type="auto">
  <name>Task 2: Update router to include highlight routes</name>
  <files>internal/controller/http/v1/router.go</files>
  <read_first>
    - internal/controller/http/v1/router.go (current router configuration)
    - internal/controller/http/v1/highlight.go (new handler to wire)
  </read_first>
  <action>
Modify `internal/controller/http/v1/router.go` to add highlight routes:

1. Add import for highlight package (if not already present - check existing imports pattern)

2. Update NewRouter function signature to accept highlight parameter:
```go
func NewRouter(handler *gin.Engine, l logger.Interface, a auth.AuthInterface, p sync.Progress, shelf library.Shelf, h highlight.Highlight) {
```

3. Add highlight routes under /syncs with device auth (after existing syncRoutes):
```go
// Highlight sync routes (uses same device auth as progress sync)
highlightRoutes := handler.Group("/syncs")
highlightRoutes.Use(authDeviceMiddleware(a, l))
newHighlightRoutes(highlightRoutes, h, l)
```

The key pattern is:
- Highlight routes go under /syncs (same as progress)
- Uses authDeviceMiddleware for device authentication (API-03)
- Pass highlight.Highlight interface to newHighlightRoutes
  </action>
  <verify>
    <automated>go build ./internal/controller/http/v1/...</automated>
  </verify>
  <acceptance_criteria>
    - File internal/controller/http/v1/router.go contains "highlight.Highlight" in NewRouter signature
    - File contains "newHighlightRoutes(highlightRoutes, h, l)"
    - File contains "highlightRoutes.Use(authDeviceMiddleware(a, l))"
    - Command `go build ./internal/controller/http/v1/...` exits 0
  </acceptance_criteria>
  <done>Router updated to wire highlight routes with device authentication</done>
</task>

<task type="auto">
  <name>Task 3: Wire dependencies in app.go</name>
  <files>internal/app/app.go</files>
  <read_first>
    - internal/app/app.go (current dependency wiring)
    - internal/highlight/sync.go (HighlightSyncUseCase)
    - internal/highlight/highlight_postgres.go (HighlightDatabaseRepo)
    - internal/controller/http/v1/router.go (updated router signature)
  </read_first>
  <action>
Modify `internal/app/app.go` to wire highlight dependencies:

1. Add import for highlight package:
```go
import "github.com/vanadium23/kompanion/internal/highlight"
```

2. After existing progress := line, add highlight service initialization:
```go
progress := sync.NewProgressSync(sync.NewProgressDatabaseRepo(pg))
highlightSync := highlight.NewHighlightSync(highlight.NewHighlightDatabaseRepo(pg))
```

3. Update v1.NewRouter call to include highlightSync:
```go
v1.NewRouter(handler, l, authService, progress, shelf, highlightSync)
```

The pattern follows existing code - create repository, create use case, pass to router.
  </action>
  <verify>
    <automated>go build ./internal/app/...</automated>
  </verify>
  <acceptance_criteria>
    - File internal/app/app.go contains `"github.com/vanadium23/kompanion/internal/highlight"` import
    - File contains "highlightSync := highlight.NewHighlightSync"
    - File contains "highlight.NewHighlightDatabaseRepo(pg)"
    - File contains "v1.NewRouter(handler, l, authService, progress, shelf, highlightSync)"
    - Command `go build ./internal/app/...` exits 0
  </acceptance_criteria>
  <done>Dependencies wired in app.go, highlight service available to router</done>
</task>

</tasks>

<verification>
After completing all tasks:
1. Full build succeeds: `go build ./...`
2. Highlight routes registered at /syncs/highlights
3. Device authentication enforced via authDeviceMiddleware
</verification>

<success_criteria>
- POST /syncs/highlights endpoint accepts highlight array (API-01, API-02)
- Device authentication via x-auth-user and x-auth-key headers (API-03)
- Response includes synced and total counts (API-04)
- Both annotation and legacy data models work (note field optional) (SYNC-03)
</success_criteria>

<output>
After completion, create `.planning/phases/01-api-storage/01-04-SUMMARY.md`
</output>
