# KOmpanion - Highlights Sync

## What This Is

A feature to synchronize book highlights and notes from KOReader to Kompanion.
Users can highlight text and add notes while reading in KOReader, and these annotations are automatically synced to Kompanion for storage and viewing.

## Core Value

Users can capture and review their book highlights in one place - Kompanion serves as the central repository for all reading annotations.

## Requirements

### Validated (Pre-existing)

- ✓ Book library management — existing
- ✓ KOReader progress sync — existing
- ✓ WebDAV statistics sync — existing
- ✓ OPDS catalog — existing
- ✓ User authentication — existing
- ✓ Device management — existing

### Validated (Phase 1: API & Storage — 2026-03-21)

- ✓ KOReader highlights sync via HTTP API — POST /syncs/highlights
- ✓ Store highlights in PostgreSQL database — highlight_annotations table
- ✓ Deduplication via content hash — ON CONFLICT DO NOTHING
- ✓ Device authentication — MD5 hash, matches progress sync pattern

### Validated (Phase 2: Web UI — 2026-03-21)

- ✓ Display highlights on book detail page — read-only section with text, note, page/chapter
- ✓ Highlight dependency wired through router to books handler
- ✓ CSS styling matching existing design system

### Validated (Phase 5: Standalone KOReader Plugin — 2026-03-22)

- ✓ Standalone KOReader plugin — `koreader/kompanion.koplugin/`
- ✓ WidgetContainer base class — Tools menu integration (not Export submenu)
- ✓ Setup dialog for URL/device credentials — persisted in G_reader_settings
- ✓ Dual format highlight extraction — annotations (new) + highlight/bookmarks (legacy)
- ✓ HTTP sync with Basic Auth — POST /syncs/highlights
- ✓ Success/error toasts — synced count or failure message

### Active

- None — all milestone requirements complete

### Out of Scope

- Two-way sync (Kompanion → KOReader) — deferred, KOReader exporter plugin doesn't support it
- Highlight editing in web UI — read-only display for now
- Highlight export from web UI — use KOReader's own export instead
- Image highlights (text only) — complexity, deferred

## Context

### Existing System

KOmpanion already has:
- Progress sync at `/syncs/progress` using MD5-hashed device credentials
- Statistics sync via WebDAV at `/webdav/statistics.sqlite3`
- Book storage with PostgreSQL backend
- Clean architecture with layered separation (entity, service, repository, controller)

### KOReader Exporter Plugin

KOReader has a built-in exporter plugin (`plugins/exporter.koplugin/`) that can export highlights to various targets:
- JSON file export
- Readwise API
- Joplin notes
- Nextcloud
- Markdown/text

The plugin uses a base exporter class with `makeJsonRequest` method for HTTP APIs.
Highlights are stored in document sidecar files and parsed via `MyClipping` class.

### Data Structure

From KOReader, each highlight has:
- `text`: the highlighted text
- `note`: optional user note
- `page`: page number or location
- `chapter`: optional chapter title
- `time`: Unix timestamp
- `drawer`: highlight style
- `color`: highlight color

## Constraints

- **Protocol**: Must use HTTP POST to match existing `/syncs/progress` pattern
- **Authentication**: Use existing device credential system (MD5 hash)
- **Database**: Store in PostgreSQL alongside existing data
- **UI**: Integrate into existing book detail page template

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| One-way sync only | KOReader exporter doesn't support fetching, user confirmed | ✓ Implemented |
| HTTP API over WebDAV | Simpler implementation, matches progress sync pattern | ✓ Implemented |
| Read-only UI | Editing adds complexity, can be added later | ✓ Implemented |
| WidgetContainer over Provider | Provider system unreliable, WidgetContainer stable like kosync | ✓ Implemented (Phase 5) |

---
*Last updated: 2026-03-22 after Phase 5 completion*
