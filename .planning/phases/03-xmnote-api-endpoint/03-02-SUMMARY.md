---
phase: 03-xmnote-api-endpoint
plan: 02
subsystem: notes-api
tags: [api, handlers, basic-auth, tdd]
dependency_graph:
  requires: [entity.Highlight, highlight.HighlightRepo]
  provides: [NoteResponse, newNotesRoutes, notesBasicAuth]
  affects: [highlight-sync, notes-viewer]
tech_stack:
  added: [internal/controller/http/v1/notes.go, internal/controller/http/v1/notes_test.go, internal/highlight/interfaces.go, internal/highlight/highlight_postgres.go, internal/highlight/sync.go]
  patterns: [TDD, Basic Auth, gin.Router, mock-based testing]
decisions:
  - One note per book with deterministic integer ID from document hash
  - Basic Auth with MD5-hashed device credentials
  - Note title uses "{author} - {title}" format for KOReader update detection
metrics:
  duration: 5min
  tasks: 4
  files: 7
  completed_date: 2026-03-21
---

# Phase 03 Plan 02: Notes API Endpoints Summary

## One-Liner

Nextcloud Notes API compatible endpoints with Basic Auth enabling KOReader exporter to sync highlights without modification.

## Changes Made

### Task 1: Add GetDocumentsByDevice to repository

- Added `DocumentInfo` struct to `internal/highlight/interfaces.go`
- Extended `HighlightRepo` interface with `GetDocumentsByDevice(ctx, deviceName) ([]DocumentInfo, error)`
- Extended `Highlight` interface with same method
- Implemented in `internal/highlight/highlight_postgres.go` using LEFT JOIN with books table
- Added implementation to `internal/highlight/sync.go`

### Task 2: Create Notes API handlers

Created `internal/controller/http/v1/notes.go`:
- `NoteResponse` struct with Nextcloud Notes API fields
- `listNotes` handler - GET /notes returns notes filtered by device
- `createNote` handler - POST /notes acknowledges note creation
- `updateNote` handler - PUT /notes/:id acknowledges note update
- `notesBasicAuth` middleware - Basic Auth using device credentials

### Task 3: Write integration tests for Notes API

Created `internal/controller/http/v1/notes_test.go` with 5 test cases:
- TestNotesList_Unauthorized - 401 without auth
- TestNotesList_Empty - empty array for device with no highlights
- TestNotesList_Success - returns notes for device with highlights
- TestNotesCreate_Success - returns 200 with note object
- TestNotesUpdate_Success - returns 200 with note object

### Task 4: Wire Notes routes in router

Modified `internal/controller/http/v1/router.go`:
- Added Notes routes at `/index.php/apps/notes/api/v1`
- Uses `notesBasicAuth` middleware
- Connected to existing `highlight.Highlight` instance

## Verification

```bash
go build ./...
go test ./internal/controller/http/v1/... -v
go test ./internal/highlight/... -v
```

All builds and tests pass.

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None - all functionality is fully implemented.

---

## Self-Check: PASSED

- [x] internal/controller/http/v1/notes.go exists
- [x] internal/controller/http/v1/notes_test.go exists
- [x] internal/highlight/interfaces.go has GetDocumentsByDevice
- [x] internal/highlight/sync.go has GetDocumentsByDevice
- [x] internal/controller/http/v1/router.go has Notes routes
