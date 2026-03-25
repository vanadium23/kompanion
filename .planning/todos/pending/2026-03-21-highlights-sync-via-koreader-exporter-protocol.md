---
created: 2026-03-21T15:27:26.457Z
title: Highlights sync via KOReader exporter protocol
area: api
files:
  - /home/deploy/koreader/plugins/exporter.koplugin/main.lua
  - /home/deploy/koreader/plugins/exporter.koplugin/clip.lua
  - /home/deploy/koreader/plugins/exporter.koplugin/base.lua
  - /home/deploy/koreader/plugins/exporter.koplugin/target/readwise.lua
  - /home/deploy/koreader/plugins/exporter.koplugin/target/json.lua
  - internal/controller/http/webdav/router.go
  - internal/sync/progress.go
---

## Problem

KOReader has a built-in highlights/annotations system with an exporter plugin that can send highlights to various targets (Readwise, JSON, Markdown, etc.). Kompanion currently syncs reading progress and statistics, but not highlights.

Users want to sync their highlights from KOReader to Kompanion so they can:
1. View all highlights in one place (web UI)
2. Sync highlights across devices
3. Export highlights later

## Solution

**Approach: Use XMNote exporter protocol (configurable IP)**

XMNote is the only KOReader exporter target that allows custom IP configuration.
Users can set their Kompanion server IP and port (8080) to send highlights.

**API Design:**
- `POST /send` - Accept highlights in XMNote format (same endpoint path as XMNote app)

**XMNote request format (from xmnote.lua):**
```json
{
  "title": "Book Title",
  "author": "Author Name",
  "type": 1,
  "locationUnit": 1,
  "readingStatus": 2,
  "readingStatusChangedDate": 1234567890,
  "source": "KOReader",
  "entries": [
    {
      "text": "highlighted text",
      "note": "user note",
      "chapter": "Chapter 1",
      "time": 1234567890,
      "page": 42
    }
  ],
  "fuzzyReadingDurations": [...]
}
```

**Mapping to highlight_annotations:**
- `entries[].text` → `text`
- `entries[].note` → `note`
- `entries[].page` → `page`
- `entries[].chapter` → `chapter`
- `entries[].time` → `highlight_time`
- `title + author` → lookup `koreader_partial_md5` from books table

**Implementation steps:**
1. Create `POST /send` endpoint (no auth initially, or IP-based)
2. Parse XMNote format and map to highlight_annotations
3. Lookup book by title/author to get document_id (MD5)
4. Store highlights with deduplication

**Reference files:**
- XMNote exporter: `/home/deploy/koreader/plugins/exporter.koplugin/target/xmnote.lua`
- Existing highlight sync: `internal/highlight/`
- Device auth pattern: `internal/auth/auth.go`
