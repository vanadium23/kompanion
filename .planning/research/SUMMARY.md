# Research Summary: Highlights Sync

**Project:** KOmpanion - Highlights Sync
**Date:** 2026-03-21

---

## Key Findings

### Stack
- **No new dependencies required** - Use existing Gin, pgx/v5, golang-migrate
- **Follow existing patterns** - Progress sync architecture is the template
- **PostgreSQL with unique constraint** - For deduplication on (document, text_hash, timestamp)

### Table Stakes
1. HTTP API endpoint for receiving highlights
2. Store highlight text, page, timestamp, book association, device
3. Display highlights on book detail page (read-only)
4. Authentication via device credentials (MD5)
5. Idempotent sync with deduplication

### Architecture
- **New package:** `internal/highlight/` with entity, interfaces, usecase, repository
- **New table:** `highlight_annotations` with unique constraint on deduplication hash
- **Modified:** `router.go`, `app.go`, `books.go`, `book.html` template
- **Reuse:** `authDeviceMiddleware` for device authentication

### Critical Pitfalls
1. **Duplicate highlights** - KOReader sends ALL highlights every time; use UPSERT with content hash
2. **Timestamp collision** - Don't use timestamp as unique ID; use composite key
3. **Orphan highlights** - Handle books not in Kompanion library; allow nullable book_id
4. **Two KOReader data models** - Support both annotations and legacy highlights+bookmarks
5. **UTF-8 encoding** - Ensure charset handling throughout

---

## Recommended Phase Structure

| Phase | Focus | Key Deliverables |
|-------|-------|------------------|
| 1 | API & Storage | Entity, migration, repository, sync usecase, HTTP endpoint |
| 2 | Web UI | Display highlights on book detail page |

---

## Data Model

```go
type Highlight struct {
    ID             string    // UUID
    DocumentID     string    // MD5 hash (koreader_partial_md5)
    Text           string    // Highlighted text
    Note           string    // Optional user note
    Page           string    // Page number or location
    Chapter        string    // Optional chapter
    Timestamp      int64     // Unix timestamp from KOReader
    Drawer         string    // Highlight style
    Color          string    // Highlight color
    Device         string    // Device name
    HighlightHash  string    // MD5(text + page + timestamp) for dedup
    CreatedAt      time.Time // Server timestamp
}
```

---

## API Contract

```
POST /syncs/highlights
Authorization: Basic device:password (MD5)
Content-Type: application/json

{
  "document": "abc123...",
  "title": "Book Title",
  "author": "Author",
  "highlights": [
    {
      "text": "Highlighted text",
      "note": "User note",
      "page": "123",
      "chapter": "Chapter 1",
      "time": 1711034567,
      "drawer": "highlight",
      "color": "yellow"
    }
  ]
}

Response: 200 OK
{
  "synced": 5,
  "total": 5
}
```

---

## Sources

- KOReader exporter plugin: `/home/deploy/koreader/plugins/exporter.koplugin/`
- Existing sync pattern: `internal/sync/progress.go`
- Project requirements: `.planning/PROJECT.md`
