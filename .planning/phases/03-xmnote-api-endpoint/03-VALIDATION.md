---
phase: 03
slug: nextcloud-notes-api-endpoint
status: ready
nyquist_compliant: true
wave_0_complete: false
created: 2026-03-21
---

# Phase 03 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test + testify + go-hit |
| **Config file** | none — existing test infrastructure |
| **Quick run command** | `go test ./internal/notes/... -v` |
| **Full suite command** | `go test ./... -v` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/notes/... -v` (formatter tests) or `go test ./internal/controller/http/v1 -run TestNotes -v` (handler tests)
- **After every plan wave:** Run `go test ./internal/controller/http/v1/... -v`
- **Before `/gsd:verify-work`:** Full suite `go test ./... -v` must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | NC-05 | unit | `go test ./internal/notes -run TestFormat -v` | W0 (plan creates) | pending |
| 03-01-02 | 01 | 1 | NC-06 | unit | `go test ./internal/notes -run TestHashToInt -v` | W0 (plan creates) | pending |
| 03-02-01 | 02 | 2 | NC-04 | integration | `go test ./internal/auth -run TestCheckDevicePassword -v` | exists | pending |
| 03-02-02 | 02 | 2 | NC-01 | integration | `go test ./internal/controller/http/v1 -run TestNotesList -v` | W0 (plan creates) | pending |
| 03-02-03 | 02 | 2 | NC-02 | integration | `go test ./internal/controller/http/v1 -run TestNotesCreate -v` | W0 (plan creates) | pending |
| 03-02-04 | 02 | 2 | NC-03 | integration | `go test ./internal/controller/http/v1 -run TestNotesUpdate -v` | W0 (plan creates) | pending |

*Status: pending / green / red / flaky*

---

## Wave 0 Requirements

Wave 0 files are created by the plans themselves (no pre-existing test scaffolding needed):

- [x] `internal/notes/formatter.go` — Created by Plan 01 Task 1
- [x] `internal/notes/formatter_test.go` — Created by Plan 01 Task 2
- [x] `internal/controller/http/v1/notes.go` — Created by Plan 02 Task 2
- [x] `internal/controller/http/v1/notes_test.go` — Created by Plan 02 Task 3
- [x] Repository method `GetDocumentsByDevice` in `internal/highlight/highlight_postgres.go` — Created by Plan 02 Task 1

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| KOReader exporter integration | NC-01, NC-02, NC-03 | Requires real KOReader device | 1. Configure KOReader Nextcloud exporter with Kompanion URL and device credentials. 2. Export highlights from a book. 3. Verify note appears in GET /notes. |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references (plans create test files)
- [x] No watch-mode flags
- [x] Feedback latency < 15s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** ready
