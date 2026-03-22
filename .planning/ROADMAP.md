# Roadmap: KOmpanion Highlights Sync

**Project:** KOmpanion - Highlights Sync
**Core Value:** Users can capture and review their book highlights in one place
**Created:** 2026-03-21

---

## Overview

This roadmap delivers the highlights sync feature in 4 phases. Each phase delivers a complete, verifiable capability.

**Total v1 Requirements:** 31 (including Phase 4 Lua plugin)
**Granularity:** Coarse

---

## Phases

- [x] **Phase 1: API & Storage** - KOReader can sync highlights via HTTP API, stored in PostgreSQL
- [x] **Phase 2: Web UI** - Users can view their synced highlights on the book detail page
- [x] **Phase 3: Nextcloud Notes API** - KOReader Nextcloud Notes exporter compatibility
- [ ] **Phase 4: KOReader Lua Plugin** - Dedicated plugin for Kompanion highlights sync

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
- [x] 01-01-PLAN.md - Create Highlight entity and database migration
- [x] 01-02-PLAN.md - Create highlight package interfaces
- [x] 01-03-PLAN.md - Implement use case and PostgreSQL repository
- [x] 01-04-PLAN.md - Implement HTTP handler and wire dependencies

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
**Plans:** 3 plans

Plans:
- [x] 02-01-PLAN.md - Wire highlight dependency through router and fetch highlights in viewBook handler
- [x] 02-02-PLAN.md - Add highlights template section to book.html
- [x] 02-03-PLAN.md - Add CSS styling for highlights

### Phase 3: Nextcloud Notes API Endpoint
**Goal:** Implement Nextcloud Notes API-compatible endpoint for KOReader exporter (replaces XMNote due to auth security)
**Depends on:** Phase 2
**Requirements:** NC-01, NC-02, NC-03, NC-04, NC-05, NC-06

**Success Criteria** (what must be TRUE):
  1. KOReader can connect using Nextcloud Notes exporter with Basic Auth
  2. GET /notes returns notes filtered by authenticated device
  3. POST /notes creates note with highlights formatted as markdown
  4. PUT /notes/{id} updates existing note by document hash
  5. One note per book containing all highlights

**Plans:** 2 plans in 2 waves

Plans:
- [x] 03-01-PLAN.md - Create notes package with markdown formatter (Wave 1)
- [x] 03-02-PLAN.md - Implement Notes API handlers and wire routes (Wave 2, depends on 03-01)

### Phase 4: KOReader Lua Plugin for Highlights Sync
**Goal:** Create a KOReader Lua plugin that exports highlights to Kompanion's existing `/syncs/highlights` API. The built-in exporter plugin is marked deprecated, so this provides a dedicated integration path.
**Depends on:** Phase 3
**Requirements:** LUA-01, LUA-02, LUA-03, LUA-04

**Success Criteria** (what must be TRUE):
  1. KOReader user sees Kompanion option in Export highlights menu
  2. User can configure server URL and device credentials via Setup dialog
  3. Export sends highlights to Kompanion /syncs/highlights endpoint
  4. Success/failure shows as toast notification in KOReader

**Plans:** 1 plan in 1 wave

Plans:
- [ ] 04-01-PLAN.md - Create KOReader Lua plugin with Provider registration (Wave 1)

---

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|-----------------|--------|-----------|
| 1. API & Storage | 4/4 | Complete | 2026-03-21 |
| 2. Web UI | 3/3 | Complete | 2026-03-21 |
| 3. Nextcloud Notes API | 2/2 | Complete | 2026-03-21 |
| 4. KOReader Lua Plugin | 0/1 | Not started | - |

---

## Coverage Map

| Category | Requirements | Phase |
|----------|--------------|-------|
| API | API-01, API-02, API-03, API-04 | 1 |
| Data Storage | DATA-01 through DATA-10 | 1 |
| Sync Behavior | SYNC-01, SYNC-02, SYNC-03 | 1 |
| Web UI | UI-01 through UI-05 | 2 |
| Notes API | NC-01, NC-02, NC-03, NC-04, NC-05, NC-06 | 3 |
| Lua Plugin | LUA-01, LUA-02, LUA-03, LUA-04 | 4 |

**Total:** 31 requirements mapped to 4 phases
**Orphaned:** 0

---

*Roadmap created: 2026-03-21*
*Last updated: 2026-03-22 after Phase 4 planning*
