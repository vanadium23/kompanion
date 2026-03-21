# Roadmap: KOmpanion Highlights Sync

**Project:** KOmpanion - Highlights Sync
**Core Value:** Users can capture and review their book highlights in one place
**Created:** 2026-03-21

---

## Overview

This roadmap delivers the highlights sync feature in 2 phases. Each phase delivers a complete, verifiable capability.

**Total v1 Requirements:** 21
**Granularity:** Coarse

---

## Phases

- [ ] **Phase 1: API & Storage** - KOReader can sync highlights via HTTP API, stored in PostgreSQL
- [ ] **Phase 2: Web UI** - Users can view their synced highlights on the book detail page

---

## Phase Details

### Phase 1: API & Storage
**Goal:** KOReader devices can sync highlights to Kompanion via HTTP API, with persistent storage
**Depends on:** Nothing (first phase)
**Requirements:** API-01, API-02, API-03, API-04, DATA-01, DATA-02, DATA-03, DATA-04, DATA-05, DATA-06, DATA-07, DATA-08, DATA-09, DATA-10, SYNC-01, SYNC-02, SYNC-03
**Success Criteria** (what must be TRUE):
  1. KOReader can POST highlights to `/syncs/highlights` and receive synced count
  2. Device authentication works (MD5 hash, matches existing progress sync pattern)
  3. Highlights are stored in PostgreSQL with all metadata (text, note, page, chapter, timestamp, drawer, color, device, document hash)
  4. Re-syncing same highlights does not create duplicates (idempotent via content hash)
  5. Highlights for unknown books are stored without errors (orphan handling)

Plans:
- [ ] 01-01-PLAN.md - Create Highlight entity and database migration
- [ ] 01-02-PLAN.md - Create highlight package interfaces
- [ ] 01-03-PLAN.md - Implement use case and PostgreSQL repository
- [ ] 01-04-PLAN.md - Implement HTTP handler and wire dependencies

### Phase 2: Web UI
**Goal:** Users can view their synced highlights on the book detail page
**Depends on:** Phase 1 (API & Storage)
**Requirements:** UI-01, UI-02, UI-03, UI-04, UI-05
**Success Criteria** (what must be TRUE):
  1. Book detail page displays all highlights for that book
  2. Each highlight shows text, page, and chapter (when available)
  3. User notes appear alongside highlight text
  4. Highlights are ordered chronologically or by page number
  5. Display is read-only (no editing controls shown)
**Plans:** TBD

---

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|-----------------|--------|-----------|
| 1. API & Storage | 0/4 | Not started | - |
| 2. Web UI | 0/1 | Not started | - |

---

## Coverage Map

| Category | Requirements | Phase |
|----------|--------------|-------|
| API | API-01, API-02, API-03, API-04 | 1 |
| Data Storage | DATA-01 through DATA-10 | 1 |
| Sync Behavior | SYNC-01, SYNC-02, SYNC-03 | 1 |
| Web UI | UI-01 through UI-05 | 2 |

**Total:** 21 requirements mapped to 2 phases
**Orphaned:** 0

---

*Roadmap created: 2026-03-21*
