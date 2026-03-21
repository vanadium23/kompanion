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

**Approach: HTTP API endpoint (like Readwise)**

Create a new target in Kompanion that accepts highlights via HTTP POST, similar to how Readwise accepts them. This allows creating a custom KOReader exporter target or using the existing JSON exporter with a custom endpoint.

**API Design:**
- `PUT /syncs/highlights` - Receive highlights from device
- `GET /syncs/highlights/{document}` - Return highlights for a book (optional, for bidirectional sync)

**Data model from KOReader exporter (clip.lua):**
```json
{
  "title": "Book Title",
  "author": "Author Name",
  "entries": [
    {
      "page": 42,
      "time": 1398127554,
      "text": "Highlighted text...",
      "note": "User note (optional)",
      "chapter": "Chapter 1",
      "sort": "highlight"
    }
  ]
}
```

**Implementation steps:**
1. Create `internal/highlights/` package with entity, repo, usecase
2. Add `highlights` table migration
3. Create `/syncs/highlights` endpoint with device auth
4. Add Web UI to view highlights per book
5. (Optional) Create KOReader exporter plugin for Kompanion

**Reference files:**
- KOReader exporter: `/home/deploy/koreader/plugins/exporter.koplugin/`
- Existing sync pattern: `internal/sync/progress.go`
- Device auth: `internal/auth/auth.go`
