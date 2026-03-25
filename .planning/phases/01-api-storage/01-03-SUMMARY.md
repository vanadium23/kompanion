---
phase: 01-api-storage
plan: 03
subsystem: highlight
tags: [sync, repository, use-case, deduplication, tdd]
requires:
  - 01
  - 02
provides:
  - HighlightSyncUseCase implementation
  - HighlightDatabaseRepo implementation
  - Unit tests with mocks
affects: []
tech_stack:
  added: []
  patterns:
    - Repository pattern with interface
    - Use case / service layer
    - MD5 hash for deduplication
    - ON CONFLICT DO NOTHING for idempotency
key_files:
  created:
    - internal/highlight/sync.go
    - internal/highlight/sync_test.go
    - internal/highlight/highlight_postgres.go
    - internal/highlight/highlight_postgres_test.go
  modified:
    - internal/highlight/mocks_test.go
decisions:
  - Use MD5 hash of (text:page:timestamp) for deduplication (simple, deterministic)
  - Continue on individual Store errors to support idempotent batch sync
  - ON CONFLICT DO NOTHING for UPSERT deduplication (database guarantees integrity)
---

# Phase 01 Plan 03: Highlight Sync Use Case & Repository Summary

**One-liner:** HighlightSyncUseCase and HighlightDatabaseRepo implemented with MD5 hash deduplication and ON CONFLICT DO NOTHING for idempotent highlight storage.

## Tasks Completed

| Task | Name | Status | Commit |
|------|------|--------|--------|
| 1 | Implement HighlightSyncUseCase | DONE | 65cec0b |
| 2 | Implement HighlightDatabaseRepo | DONE | 3b66cd7 |
| 3 | Generate mocks for testing | DONE | 65cec0b (included) |

## Implementation Details

### Task 1: HighlightSyncUseCase

Created `internal/highlight/sync.go` with:
- `HighlightSyncUseCase` struct with `HighlightRepo` dependency
- `Sync()` method that iterates over highlights, sets fields, generates hash, and stores
- `Fetch()` method that delegates to repository
- `generateHash()` function creating MD5 hash from `text:page:timestamp`

Key behaviors:
- Sets `DocumentID`, `AuthDeviceName`, `CreatedAt`, `HighlightHash` on each highlight
- Continues on individual `Store` errors (supports idempotent re-sync)
- Returns count of successfully synced highlights

### Task 2: HighlightDatabaseRepo

Created `internal/highlight/highlight_postgres.go` with:
- `HighlightDatabaseRepo` struct embedding `*postgres.Postgres`
- `Store()` method using `ON CONFLICT DO NOTHING` for deduplication
- `GetByDocumentID()` method retrieving highlights ordered by `highlight_time ASC`
- Full 12-field storage including metadata (drawer, color, chapter)

Key behaviors:
- Idempotent storage via unique constraint on `(koreader_partial_md5, highlight_hash)`
- Timestamp conversion from Unix epoch to PostgreSQL TIMESTAMPTZ
- Handles orphan highlights (no foreign key to books)

### Task 3: Mock Generation

Generated mocks via `go generate ./internal/highlight/...`:
- `MockHighlightRepo` with `Store` and `GetByDocumentID` methods
- `MockHighlight` with `Sync` and `Fetch` methods

## Files Created/Modified

| File | Purpose | Lines |
|------|---------|-------|
| internal/highlight/sync.go | Use case implementation | 52 |
| internal/highlight/sync_test.go | Use case unit tests | 162 |
| internal/highlight/highlight_postgres.go | Repository implementation | 76 |
| internal/highlight/highlight_postgres_test.go | Repository unit tests | 138 |
| internal/highlight/mocks_test.go | Generated mocks | 119 |

## Requirements Satisfied

| ID | Description | Status |
|----|-------------|--------|
| SYNC-01 | Re-syncing same highlights does not create duplicates | DONE |
| SYNC-02 | Highlights for books not in library are stored (orphan handling) | DONE |

## Test Results

```
=== RUN   TestHighlightRepo_Store
--- PASS: TestHighlightRepo_Store (0.00s)
=== RUN   TestHighlightRepo_GetByDocumentID
--- PASS: TestHighlightRepo_GetByDocumentID (0.00s)
=== RUN   TestHighlightRepo_GetByDocumentID_Empty
--- PASS: TestHighlightRepo_GetByDocumentID_Empty (0.00s)
=== RUN   TestHighlightSync_NewHighlightSync
--- PASS: TestHighlightSync_NewHighlightSync (0.00s)
=== RUN   TestHighlightSync_Sync_EmptyHighlights
--- PASS: TestHighlightSync_Sync_EmptyHighlights (0.00s)
=== RUN   TestHighlightSync_Sync_SetsFields
--- PASS: TestHighlightSync_Sync_SetsFields (0.00s)
=== RUN   TestHighlightSync_Sync_ContinuesOnError
--- PASS: TestHighlightSync_Sync_ContinuesOnError (0.00s)
=== RUN   TestHighlightSync_Sync_GeneratesHash
--- PASS: TestHighlightSync_Sync_GeneratesHash (0.00s)
=== RUN   TestHighlightSync_Fetch
--- PASS: TestHighlightSync_Fetch (0.00s)
=== RUN   TestHighlightSync_Fetch_Error
--- PASS: TestHighlightSync_Fetch_Error (0.00s)
PASS
ok  	github.com/vanadium23/kompanion/internal/highlight	0.004s
```

## Deviations from Plan

None - plan executed exactly as written.

## Metrics

| Metric | Value |
|--------|-------|
| Duration | ~5 min |
| Tasks | 3 |
| Files | 5 |
| Test Coverage | 10 tests |

## Self-Check: PASSED

- [x] All created files exist
- [x] All commits exist in git log
- [x] All tests pass
