# Technology Stack - Highlights Sync

**Project:** KOmpanion - Highlights Sync Feature
**Researched:** 2026-03-21

## Recommended Stack

### Core Framework

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.26 | Backend language | Matches existing codebase version requirement |
| gin-gonic/gin | v1.7.7 (existing) | HTTP web framework | Already in use; stable; consider upgrading to v1.12.0 for latest fixes |
| jackc/pgx/v5 | v5.6.0 (existing) | PostgreSQL driver | Already in use; excellent performance; supports advanced PostgreSQL features |

### Database

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| PostgreSQL | Existing | Primary data store | Matches existing infrastructure; JSONB support for flexible highlight metadata |
| golang-migrate/migrate | v4.15.1 (existing) | Database migrations | Already in use; create new migration for highlights table |

### Validation

| Library | Version | Purpose | Why |
|---------|---------|---------|-----|
| go-playground/validator/v10 | v10.9.0 (via Gin) | Request validation | Built into Gin binding; supports struct tags; current latest is v10.27.0 |

### Testing

| Library | Version | Purpose | Why |
|---------|---------|---------|-----|
| stretchr/testify | v1.11.1 (existing) | Assertions and mocks | Already in use |
| pashagolub/pgxmock/v4 | v4.2.0 (existing) | PostgreSQL mocking | Already in use; matches pgx driver |
| golang/mock | v1.6.0 (existing) | Mock generation | Already in use |

### Supporting Libraries

| Library | Purpose | When to Use |
|---------|---------|-------------|
| Existing logger (rs/zerolog v1.26.1) | Structured logging | All error/info logging in highlights sync |
| Existing UUID (moroz/uuidv7-go) | ID generation | If highlights need unique IDs |

## KOReader Integration

### Exporter Plugin Protocol

KOReader's exporter plugin sends HTTP POST requests with JSON payloads. Based on analysis of the [KOReader exporter plugin source](https://github.com/koreader/koreader/blob/master/plugins/exporter.koplugin/):

**Request Format (from Readwise target as reference):**
```json
{
  "highlights": [
    {
      "text": "highlighted text content",
      "title": "Book Title",
      "author": "Author Name",
      "source_type": "koreader",
      "category": "books",
      "note": "user's optional note",
      "location": "page number or location",
      "location_type": "order",
      "highlighted_at": "2026-03-21T10:30:00Z"
    }
  ]
}
```

**Base exporter provides:**
- `makeJsonRequest(endpoint, method, body, headers)` - HTTP POST with JSON body
- Settings storage per target (enabled, token, etc.)
- Timestamp handling

## New Components Required

### Entity Layer

```go
// internal/entity/highlight.go
type Highlight struct {
    ID           string    `json:"id"`
    DocumentID   string    `json:"document_id"`     // MD5 hash matching koreader_partial_md5
    Text         string    `json:"text" binding:"required"`
    Note         string    `json:"note"`
    Page         string    `json:"page"`
    Chapter      string    `json:"chapter"`
    Timestamp    int64     `json:"time"`
    Drawer       string    `json:"drawer"`          // highlight style
    Color        string    `json:"color"`           // highlight color
    DeviceName   string    `json:"device_name"`
    CreatedAt    time.Time `json:"created_at"`
}

type HighlightSyncRequest struct {
    DocumentID string      `json:"document_id" binding:"required"`
    Title      string      `json:"title"`
    Author     string      `json:"author"`
    Highlights []Highlight `json:"highlights" binding:"required,dive"`
}
```

### Database Schema

```sql
-- migrations/YYYYMMDD_highlights.up.sql
CREATE TABLE highlights (
    id BIGSERIAL PRIMARY KEY,
    koreader_partial_md5 TEXT NOT NULL,  -- links to library_book.koreader_partial_md5
    text TEXT NOT NULL,
    note TEXT,
    page TEXT,
    chapter TEXT,
    drawer TEXT,
    color TEXT,
    koreader_timestamp BIGINT NOT NULL,
    koreader_device TEXT NOT NULL,
    auth_device_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(koreader_partial_md5, text, koreader_timestamp)  -- dedupe same highlight
);

CREATE INDEX highlights_document ON highlights(koreader_partial_md5);
CREATE INDEX highlights_created ON highlights(created_at DESC);

COMMENT ON TABLE highlights IS 'Book highlights synced from KOReader devices';
```

### Repository Interface

```go
// internal/highlight/interfaces.go
type HighlightRepo interface {
    Store(ctx context.Context, h entity.Highlight) error
    StoreBatch(ctx context.Context, highlights []entity.Highlight) error
    GetByDocument(ctx context.Context, documentID string) ([]entity.Highlight, error)
    GetRecent(ctx context.Context, limit int) ([]entity.Highlight, error)
}

type Highlight interface {
    Sync(ctx context.Context, req entity.HighlightSyncRequest, deviceName string) error
    Fetch(ctx context.Context, documentID string) ([]entity.Highlight, error)
}
```

### Controller Routes

Following the existing `/syncs/progress` pattern:

```go
// internal/controller/http/v1/highlight.go
func newHighlightRoutes(handler *gin.RouterGroup, h highlight.Highlight, l logger.Interface) {
    r := &highlightRoutes{h, l}
    h := handler.Group("/highlights")
    {
        h.POST("/sync", r.syncHighlights)      // KOReader pushes highlights
        h.GET("/:document", r.fetchHighlights) // Fetch highlights for a book
    }
}
```

## API Endpoint Design

### POST /highlights/sync

**Request:**
```json
{
  "document_id": "abc123def456",
  "title": "Book Title",
  "author": "Author Name",
  "highlights": [
    {
      "text": "Highlighted text",
      "note": "My note",
      "page": "42",
      "chapter": "Chapter 3",
      "time": 1711022400,
      "drawer": "highlight_yellow",
      "color": "yellow"
    }
  ]
}
```

**Response:**
```json
{
  "message": "Highlights synced successfully",
  "count": 1,
  "document_id": "abc123def456"
}
```

**Authentication:** Uses existing device credential middleware (MD5 hash)

### GET /highlights/:document

**Response:**
```json
{
  "document_id": "abc123def456",
  "highlights": [
    {
      "text": "Highlighted text",
      "note": "My note",
      "page": "42",
      "chapter": "Chapter 3",
      "timestamp": 1711022400,
      "created_at": "2026-03-21T10:30:00Z"
    }
  ]
}
```

## Installation

No new dependencies required. All components use existing libraries.

```bash
# Run new migration
migrate -path migrations -database $KOMPANION_PG_URL up

# Generate mocks (if needed)
mockgen -source=internal/highlight/interfaces.go -destination=internal/highlight/mocks_test.go -package=highlight_test
```

## Upgrade Considerations

| Package | Current | Latest | Action |
|---------|---------|--------|--------|
| gin-gonic/gin | v1.7.7 | v1.12.0 | Consider upgrading post-feature for security fixes |
| go-playground/validator | v10.9.0 | v10.27.0 | Gin v1.12.0 includes updated validator |
| jackc/pgx/v5 | v5.6.0 | v5.7.x | Minor upgrade available if needed |

## Sources

- [KOReader Exporter Plugin](https://github.com/koreader/koreader/blob/master/plugins/exporter.koplugin/main.lua) - Plugin architecture and request format
- [KOReader Readwise Target](https://github.com/koreader/koreader/blob/master/plugins/exporter.koplugin/target/readwise.lua) - HTTP API integration reference
- [KOReader Base Exporter](https://github.com/koreader/koreader/blob/master/plugins/exporter.koplugin/base.lua) - makeJsonRequest implementation
- [Gin Releases](https://github.com/gin-gonic/gin/releases) - v1.12.0 latest
- [go-playground/validator Releases](https://github.com/go-playground/validator/releases) - v10.27.0 latest

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Framework (Gin) | HIGH | Already in use, stable API |
| Database (pgx) | HIGH | Already in use, straightforward |
| KOReader Protocol | MEDIUM | Based on source code analysis, may need testing with actual device |
| Validation | HIGH | Standard Gin binding patterns |
| API Design | HIGH | Follows existing progress sync pattern exactly |

---

*Stack research: 2026-03-21*
