# Phase 1: API & Storage - Research

**Researched:** 2026-03-21
**Domain:** KOReader Highlights Synchronization via HTTP API with PostgreSQL Storage
**Confidence:** HIGH

## Summary

This phase implements the server-side API and storage infrastructure for syncing book highlights from KOReader devices to Kompanion. The implementation follows existing patterns from the progress sync feature and requires no new external dependencies.

**Primary recommendation:** Follow the existing `internal/sync/` architecture exactly - create `internal/highlight/` package with entity, interfaces, use case, and PostgreSQL repository layers. Use unique constraint on `(koreader_partial_md5, text_hash, timestamp)` for deduplication.

<phase_requirements>

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| API-01 | KOReader can sync highlights via HTTP POST to `/syncs/highlights` | Existing `/syncs/progress` pattern; `authDeviceMiddleware` for auth |
| API-02 | API accepts array of highlights in single request | Gin JSON binding; batch Store method in repository |
| API-03 | API uses device authentication (MD5 hash, existing pattern) | `authDeviceMiddleware` in `internal/controller/http/v1/users.go:31-48` |
| API-04 | API returns synced count and total count | Response pattern from `internal/controller/http/v1/sync.go` |
| DATA-01 | Highlights stored in PostgreSQL `highlight_annotations` table | Migration pattern from `migrations/20250211190954_sync.up.sql` |
| DATA-02 | Highlight text is stored (required) | `text TEXT NOT NULL` in schema |
| DATA-03 | User note is stored (optional) | `note TEXT` nullable in schema |
| DATA-04 | Page/location is stored | `page TEXT NOT NULL` in schema |
| DATA-05 | Chapter is stored (optional) | `chapter TEXT` nullable in schema |
| DATA-06 | Timestamp from KOReader is stored | `highlight_time TIMESTAMPTZ NOT NULL` in schema |
| DATA-07 | Highlight style (drawer) and color are stored | `drawer TEXT`, `color TEXT` in schema |
| DATA-08 | Device name is stored | `koreader_device TEXT NOT NULL`, `auth_device_name TEXT NOT NULL` |
| DATA-09 | Document MD5 hash is stored for book matching | `koreader_partial_md5 TEXT NOT NULL` matches library_book |
| DATA-10 | Content hash for deduplication is stored | `highlight_hash TEXT NOT NULL` with unique index |
| SYNC-01 | Re-syncing same highlights does not create duplicates | UPSERT with `ON CONFLICT DO NOTHING` using unique hash |
| SYNC-02 | Highlights for books not in library are stored (orphan handling) | Nullable book reference; `koreader_partial_md5` stored directly |
| SYNC-03 | Both KOReader data models supported (annotations + legacy) | Request struct accepts both `note` field (annotations) and matches bookmarks |

</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| gin-gonic/gin | v1.7.7 (existing) | HTTP web framework | Already in use; stable; established pattern in codebase |
| jackc/pgx/v5 | v5.6.0 (existing) | PostgreSQL driver | Already in use; excellent performance; supports advanced features |
| golang-migrate/migrate | v4.15.1 (existing) | Database migrations | Already in use; migration pattern established |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| stretchr/testify | v1.11.1 (existing) | Assertions and mocks | All test files |
| pashagolub/pgxmock/v4 | v4.2.0 (existing) | PostgreSQL mocking | Repository unit tests |
| golang/mock | v1.6.0 (existing) | Mock generation | Interface mocking via `go:generate` |
| rs/zerolog | v1.26.1 (existing) | Structured logging | Error/info logging |
| moroz/uuidv7-go | (existing) | UUIDv7 generation | Not required - use BIGSERIAL or gen_random_uuid() |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| gen_random_uuid() for ID | UUIDv7 from moroz/uuidv7-go | UUIDv7 is time-sortable but adds complexity for minimal benefit |
| UNIQUE constraint for dedup | Application-level dedup check | Database constraint is simpler and guarantees integrity |

**Installation:**
No new dependencies required. All components use existing libraries.

```bash
# Run new migration
make migrate-create MIGRATE_NAME=highlights
make migrate-up
```

## Architecture Patterns

### Recommended Project Structure

```
internal/
├── highlight/                    # NEW - highlight sync package
│   ├── interfaces.go             # HighlightRepo, Highlight interfaces
│   ├── sync.go                   # HighlightSyncUseCase
│   ├── highlight_postgres.go     # PostgreSQL repository
│   ├── highlight_postgres_test.go
│   └── mocks_test.go             # Generated mocks
├── entity/
│   └── highlight.go              # NEW - Highlight entity
├── controller/http/v1/
│   ├── highlight.go              # NEW - HTTP routes
│   └── router.go                 # MODIFY - add highlight routes
└── app/
    └── app.go                    # MODIFY - wire dependencies
migrations/
└── YYYYMMDDHHMMSS_highlights.up.sql   # NEW - table creation
```

### Pattern 1: Repository Pattern with Interface

**What:** Define repository interface in service layer, implement in data layer.
**When to use:** All data access operations.

**Example:**
```go
// internal/highlight/interfaces.go
package highlight

import (
    "context"
    "github.com/vanadium23/kompanion/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=highlight_test

type HighlightRepo interface {
    Store(ctx context.Context, h entity.Highlight) error
    GetByDocumentID(ctx context.Context, documentID string) ([]entity.Highlight, error)
}

type Highlight interface {
    Sync(ctx context.Context, documentID string, highlights []entity.Highlight, deviceName string) (int, error)
    Fetch(ctx context.Context, documentID string) ([]entity.Highlight, error)
}
```

**Source:** Existing pattern from `internal/sync/interfaces.go`

### Pattern 2: Use Case / Service Layer

**What:** Encapsulate business logic in dedicated service structs.
**When to use:** All business operations beyond CRUD.

**Example:**
```go
// internal/highlight/sync.go
package highlight

import (
    "context"
    "crypto/md5"
    "encoding/hex"
    "fmt"
    "time"

    "github.com/vanadium23/kompanion/internal/entity"
)

type HighlightSyncUseCase struct {
    repo HighlightRepo
}

func NewHighlightSync(r HighlightRepo) *HighlightSyncUseCase {
    return &HighlightSyncUseCase{repo: r}
}

func (uc *HighlightSyncUseCase) Sync(ctx context.Context, documentID string, highlights []entity.Highlight, deviceName string) (int, error) {
    synced := 0
    for i := range highlights {
        highlights[i].DocumentID = documentID
        highlights[i].AuthDeviceName = deviceName
        highlights[i].CreatedAt = time.Now()
        highlights[i].HighlightHash = generateHash(highlights[i].Text, highlights[i].Page, highlights[i].Timestamp)

        if err := uc.repo.Store(ctx, highlights[i]); err != nil {
            // Log and continue - unique constraint violation is expected for duplicates
            continue
        }
        synced++
    }
    return synced, nil
}

func generateHash(text, page string, timestamp int64) string {
    data := fmt.Sprintf("%s:%s:%d", text, page, timestamp)
    hash := md5.Sum([]byte(data))
    return hex.EncodeToString(hash[:])
}
```

**Source:** Pattern from `internal/sync/progress.go`

### Pattern 3: HTTP Route Handler

**What:** Bind JSON, call service, return appropriate response.
**When to use:** All HTTP endpoints.

**Example:**
```go
// internal/controller/http/v1/highlight.go
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
    Document  string            `json:"document" binding:"required"`
    Title     string            `json:"title"`
    Author    string            `json:"author"`
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
```

**Source:** Pattern from `internal/controller/http/v1/sync.go`

### Anti-Patterns to Avoid

- **Tight coupling between layers:** Never call database directly from controller. Always use service interface.
- **Duplicate storage without deduplication:** KOReader sends ALL highlights every time. Use unique constraint on content hash.
- **Ignoring existing auth patterns:** Do NOT create new auth. Reuse `authDeviceMiddleware`.
- **Using timestamp as primary identifier:** Timestamps can collide. Use composite hash of `(text, page, timestamp)`.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Authentication middleware | New auth system | `authDeviceMiddleware` | Already exists, MD5 hash matching KOReader |
| Deduplication logic | Application-level checking | PostgreSQL UNIQUE constraint | Database guarantees integrity, handles race conditions |
| ID generation | Custom UUID logic | `gen_random_uuid()` or BIGSERIAL | PostgreSQL built-in, simpler |
| JSON binding | Manual parsing | Gin's `ShouldBindJSON` | Built-in validation, well-tested |
| Content hash | Custom hash function | `crypto/md5` + `hex.EncodeToString` | Standard library, matches KOReader's approach |

**Key insight:** The existing codebase has well-established patterns for sync operations. Following them exactly reduces risk and accelerates implementation.

## Common Pitfalls

### Pitfall 1: Duplicate Highlights on Re-Sync

**What goes wrong:** Every sync creates duplicate highlights instead of updating existing ones.

**Why it happens:** KOReader sends ALL highlights every time, not just new ones. The exporter plugin has no concept of "sync state".

**How to avoid:**
1. Generate deterministic hash: `MD5(text + page + timestamp)`
2. Use PostgreSQL `INSERT ... ON CONFLICT DO NOTHING`
3. Create unique constraint: `UNIQUE (koreader_partial_md5, highlight_hash)`

**Warning signs:** Highlights table grows faster than reading rate; same text appears multiple times.

### Pitfall 2: Orphan Highlights Without Book Match

**What goes wrong:** Highlights arrive for books not in Kompanion library, causing errors.

**Why it happens:** KOReader can highlight any document, including sideloaded files.

**How to avoid:**
1. Store `koreader_partial_md5` directly (nullable book reference)
2. Do NOT require foreign key to library_book
3. Handle lookup failure gracefully

**Warning signs:** Foreign key violation errors; NULL book_id in queries.

### Pitfall 3: Two KOReader Data Models

**What goes wrong:** Notes lost for users on older KOReader versions.

**Why it happens:** KOReader has TWO models:
- **Annotations (newer):** `item.note` directly on annotation
- **Highlights + Bookmarks (legacy):** Note stored in separate bookmarks table

**How to avoid:**
1. Request struct accepts `note` field directly
2. Store note as nullable - don't require it
3. KOReader exporter handles legacy conversion before sending

**Warning signs:** Empty note field when KOReader shows note exists.

### Pitfall 4: Character Encoding Corruption

**What goes wrong:** Non-ASCII characters corrupted during sync.

**How to avoid:**
1. PostgreSQL uses UTF-8 encoding: `ENCODING 'UTF8'`
2. Content-Type: `application/json; charset=utf-8`
3. Go's `json.Marshal` handles UTF-8 correctly

**Warning signs:** Mojibake in stored text; database encoding errors.

### Pitfall 5: Timestamp Collision for Identity

**What goes wrong:** Using only timestamp as unique ID causes missed highlights.

**Why it happens:** Multiple highlights created in rapid succession can have identical timestamps.

**How to avoid:**
1. Do NOT use timestamp as primary identifier
2. Use composite key with content hash
3. Accept timestamp as metadata, not identity

**Warning signs:** Highlights with same-second timestamps merged or skipped.

## Code Examples

### Entity Definition

```go
// internal/entity/highlight.go
package entity

import "time"

type Highlight struct {
    ID             string    `json:"id"`
    DocumentID     string    `json:"document"`          // MD5 hash (koreader_partial_md5)
    Text           string    `json:"text" binding:"required"`
    Note           string    `json:"note"`
    Page           string    `json:"page"`
    Chapter        string    `json:"chapter"`
    Timestamp      int64     `json:"time"`
    Drawer         string    `json:"drawer"`            // highlight style
    Color          string    `json:"color"`             // highlight color
    Device         string    `json:"device"`
    DeviceID       string    `json:"device_id"`
    AuthDeviceName string    `json:"-"`
    HighlightHash  string    `json:"-"`
    CreatedAt      time.Time `json:"created_at"`
}
```

**Source:** Pattern from `internal/entity/progress.go`

### Database Migration

```sql
-- migrations/YYYYMMDDHHMMSS_highlights.up.sql
CREATE TABLE highlight_annotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    koreader_partial_md5 TEXT NOT NULL,

    -- Highlight content
    text TEXT NOT NULL,
    note TEXT,
    page TEXT NOT NULL,
    chapter TEXT,

    -- Highlight metadata
    drawer TEXT,
    color TEXT,

    -- Timestamps
    highlight_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Device info
    koreader_device TEXT NOT NULL,
    koreader_device_id TEXT,
    auth_device_name TEXT NOT NULL,

    -- Deduplication
    highlight_hash TEXT NOT NULL
);

CREATE INDEX idx_highlights_document ON highlight_annotations(koreader_partial_md5);
CREATE INDEX idx_highlights_time ON highlight_annotations(highlight_time DESC);
CREATE UNIQUE INDEX idx_highlights_unique ON highlight_annotations(koreader_partial_md5, highlight_hash);

COMMENT ON TABLE highlight_annotations IS 'Stores book highlights synced from KOReader devices';
COMMENT ON COLUMN highlight_annotations.highlight_hash IS 'MD5 hash of (text:page:timestamp) for deduplication';
```

**Source:** Pattern from `migrations/20250211190954_sync.up.sql`

### PostgreSQL Repository

```go
// internal/highlight/highlight_postgres.go
package highlight

import (
    "context"
    "fmt"
    "time"

    "github.com/vanadium23/kompanion/internal/entity"
    "github.com/vanadium23/kompanion/pkg/postgres"
)

type HighlightDatabaseRepo struct {
    *postgres.Postgres
}

func NewHighlightDatabaseRepo(pg *postgres.Postgres) *HighlightDatabaseRepo {
    return &HighlightDatabaseRepo{pg}
}

func (r *HighlightDatabaseRepo) Store(ctx context.Context, h entity.Highlight) error {
    sql := `INSERT INTO highlight_annotations
        (koreader_partial_md5, text, note, page, chapter, drawer, color,
         highlight_time, koreader_device, koreader_device_id, auth_device_name, highlight_hash)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        ON CONFLICT (koreader_partial_md5, highlight_hash) DO NOTHING`

    args := []interface{}{
        h.DocumentID, h.Text, h.Note, h.Page, h.Chapter, h.Drawer, h.Color,
        time.Unix(h.Timestamp, 0), h.Device, h.DeviceID, h.AuthDeviceName, h.HighlightHash,
    }

    _, err := r.Pool.Exec(ctx, sql, args...)
    if err != nil {
        return fmt.Errorf("HighlightDatabaseRepo - Store - r.Pool.Exec: %w", err)
    }
    return nil
}

func (r *HighlightDatabaseRepo) GetByDocumentID(ctx context.Context, documentID string) ([]entity.Highlight, error) {
    sql := `SELECT id, koreader_partial_md5, text, note, page, chapter, drawer, color,
            highlight_time, koreader_device, koreader_device_id, auth_device_name, created_at
            FROM highlight_annotations
            WHERE koreader_partial_md5 = $1
            ORDER BY highlight_time ASC`

    rows, err := r.Pool.Query(ctx, sql, documentID)
    if err != nil {
        return nil, fmt.Errorf("HighlightDatabaseRepo - GetByDocumentID - r.Pool.Query: %w", err)
    }
    defer rows.Close()

    var highlights []entity.Highlight
    for rows.Next() {
        var h entity.Highlight
        var highlightTime time.Time
        err = rows.Scan(&h.ID, &h.DocumentID, &h.Text, &h.Note, &h.Page, &h.Chapter,
            &h.Drawer, &h.Color, &highlightTime, &h.Device, &h.DeviceID, &h.AuthDeviceName, &h.CreatedAt)
        if err != nil {
            return nil, fmt.Errorf("HighlightDatabaseRepo - GetByDocumentID - rows.Scan: %w", err)
        }
        h.Timestamp = highlightTime.Unix()
        highlights = append(highlights, h)
    }
    return highlights, nil
}
```

**Source:** Pattern from `internal/sync/progress_postgres.go`

### Router Integration

```go
// In internal/controller/http/v1/router.go - add to NewRouter function:

// Add highlight routes under /syncs with device auth
highlightRoutes := handler.Group("/syncs")
highlightRoutes.Use(authDeviceMiddleware(a, l))
newHighlightRoutes(highlightRoutes, h, l)
```

```go
// In internal/app/app.go - add to Run function:

import "github.com/vanadium23/kompanion/internal/highlight"

// After existing services:
highlightSync := highlight.NewHighlightSync(highlight.NewHighlightDatabaseRepo(pg))

// Update router call:
v1.NewRouter(handler, l, authService, progress, shelf, highlightSync)
```

**Source:** Pattern from `internal/app/app.go`

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Application-level deduplication | Database UNIQUE constraint | This implementation | Simpler, race-condition safe |
| Single highlight per request | Batch array of highlights | This implementation | Efficient for heavy readers |
| Foreign key to books required | Nullable book reference | This implementation | Handles orphan highlights |

**Deprecated/outdated:**
- Legacy KOReader highlights+bookmarks model: KOReader exporter converts to annotations before sending, so API only needs to handle annotation format.

## Open Questions

1. **Should we store `koreader_device_id`?**
   - What we know: KOReader sends it; progress sync stores it
   - What's unclear: Is it useful for any queries?
   - Recommendation: Store it for consistency with progress sync pattern

2. **What HTTP status for partial sync failure?**
   - What we know: Individual highlight failures should not fail the whole request
   - What's unclear: Should we return 207 Multi-Status or just 200 with counts?
   - Recommendation: Return 200 with synced/total counts - simpler and matches KOReader expectations

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | stretchr/testify v1.11.1 + pashagolub/pgxmock/v4 |
| Config file | None - tests self-contained |
| Quick run command | `go test -v -cover ./internal/highlight/...` |
| Full suite command | `make test` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| API-01 | POST to /syncs/highlights | integration | `make integration-test` | ❌ Wave 0 |
| API-02 | Accept array of highlights | unit | `go test ./internal/controller/http/v1/...` | ❌ Wave 0 |
| API-03 | Device authentication | unit | `go test ./internal/controller/http/v1/...` | ❌ Wave 0 |
| API-04 | Return synced/total counts | unit | `go test ./internal/controller/http/v1/...` | ❌ Wave 0 |
| DATA-01 | Store in highlight_annotations | unit | `go test ./internal/highlight/...` | ❌ Wave 0 |
| DATA-02 | Text field required | unit | `go test ./internal/highlight/...` | ❌ Wave 0 |
| DATA-03 | Note field optional | unit | `go test ./internal/highlight/...` | ❌ Wave 0 |
| DATA-04 | Page field stored | unit | `go test ./internal/highlight/...` | ❌ Wave 0 |
| DATA-05 | Chapter field optional | unit | `go test ./internal/highlight/...` | ❌ Wave 0 |
| DATA-06 | Timestamp stored | unit | `go test ./internal/highlight/...` | ❌ Wave 0 |
| DATA-07 | Drawer/color stored | unit | `go test ./internal/highlight/...` | ❌ Wave 0 |
| DATA-08 | Device name stored | unit | `go test ./internal/highlight/...` | ❌ Wave 0 |
| DATA-09 | Document MD5 stored | unit | `go test ./internal/highlight/...` | ❌ Wave 0 |
| DATA-10 | Content hash stored | unit | `go test ./internal/highlight/...` | ❌ Wave 0 |
| SYNC-01 | No duplicates on re-sync | unit | `go test ./internal/highlight/...` | ❌ Wave 0 |
| SYNC-02 | Orphan highlights stored | unit | `go test ./internal/highlight/...` | ❌ Wave 0 |
| SYNC-03 | Both data models supported | integration | `make integration-test` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test -v -cover ./internal/highlight/...`
- **Per wave merge:** `make test`
- **Phase gate:** `make test && make integration-test` - all green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/highlight/highlight_postgres_test.go` - unit tests for repository (DATA-01 through DATA-10)
- [ ] `internal/highlight/sync_test.go` - unit tests for use case (SYNC-01, SYNC-02)
- [ ] `internal/controller/http/v1/highlight_test.go` - unit tests for HTTP handlers (API-01 through API-04)
- [ ] `integration-test/highlight_test.go` - integration test for full sync flow (SYNC-03)
- [ ] `internal/highlight/mocks_test.go` - generated mocks (run `go generate ./internal/highlight/...`)

## Sources

### Primary (HIGH confidence)
- `internal/sync/progress.go` - Existing sync use case pattern (exact template to follow)
- `internal/sync/progress_postgres.go` - Existing repository pattern (exact template to follow)
- `internal/controller/http/v1/sync.go` - Existing HTTP handler pattern (exact template to follow)
- `migrations/20250211190954_sync.up.sql` - Existing migration pattern (exact template to follow)
- `/home/deploy/koreader/frontend/apps/reader/modules/readerannotation.lua` - KOReader annotation data model

### Secondary (MEDIUM confidence)
- `/home/deploy/koreader/plugins/exporter.koplugin/clip.lua` - KOReader clipping parser
- `/home/deploy/koreader/plugins/exporter.koplugin/target/readwise.lua` - HTTP API integration reference

### Tertiary (LOW confidence)
- None - all critical patterns verified from source code

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries already in use, no new dependencies
- Architecture: HIGH - Exact pattern exists in `internal/sync/` to follow
- Pitfalls: HIGH - Analyzed KOReader source code directly for data models

**Research date:** 2026-03-21
**Valid until:** 30 days (stable Go/PostgreSQL patterns)

---

*Research synthesized from existing project documentation and KOReader source code analysis*
