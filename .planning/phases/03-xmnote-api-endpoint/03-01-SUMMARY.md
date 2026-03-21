---
phase: 03-xmnote-api-endpoint
plan: 01
subsystem: notes
tags: [formatter, markdown, tdd]
dependency_graph:
  requires: [entity.Highlight]
  provides: [FormatHighlights, FormatTitle, HashToInt]
  affects: [notes-api]
tech_stack:
  added: [internal/notes package]
  patterns: [TDD, pure functions, strings.Builder]
key_files:
  created:
    - internal/notes/formatter.go
    - internal/notes/formatter_test.go
  modified: []
decisions:
  - Use CRC32 IEEE for hash-to-integer conversion (stable, fast)
  - Blockquote format for highlight text (D-07)
  - Plain text for user notes (D-07)
metrics:
  duration: 2min
  tasks: 2
  files: 2
  completed_date: 2026-03-21
---

# Phase 03 Plan 01: Markdown Formatter Summary

## One-Liner

Markdown formatter for converting KOReader highlights to Nextcloud Notes compatible format with stable integer ID generation.

## Changes Made

### Task 1: Create markdown formatter with highlight conversion

Created `internal/notes/formatter.go` with three exported functions:

- `FormatHighlights(title, author string, highlights []entity.Highlight) string` - Converts highlights to markdown matching KOReader's md.lua template
- `FormatTitle(author, title string) string` - Produces "{author} - {title}" format for note identification
- `HashToInt(hash string) int` - Converts string to stable integer using CRC32 IEEE

### Task 2: Write unit tests for formatter

Created `internal/notes/formatter_test.go` with 8 test functions covering:
- Single highlight formatting
- Multiple chapter grouping
- Note with separator
- Empty chapter handling
- Title formatting (single and multi-author)
- Hash stability and uniqueness

## Verification

All tests pass with 100% coverage:

```bash
go test ./internal/notes -v -cover
# 8 tests, 100.0% coverage
```

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None - formatter is fully implemented with no placeholder values.

---

## Self-Check: PASSED

- [x] internal/notes/formatter.go exists
- [x] internal/notes/formatter_test.go exists
- [x] Commit 7423bb7 exists in git log
