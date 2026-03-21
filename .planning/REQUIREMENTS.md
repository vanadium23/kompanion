# Requirements: KOmpanion Highlights Sync

**Defined:** 2026-03-21
**Core Value:** Users can capture and review their book highlights in one place

## v1 Requirements

Requirements for highlights sync feature. Each maps to roadmap phases.

### API

- [x] **API-01**: KOReader can sync highlights via HTTP POST to `/syncs/highlights`
- [x] **API-02**: API accepts array of highlights in single request
- [x] **API-03**: API uses device authentication (MD5 hash, existing pattern)
- [x] **API-04**: API returns synced count and total count

### Data Storage

- [x] **DATA-01**: Highlights stored in PostgreSQL `highlight_annotations` table
- [x] **DATA-02**: Highlight text is stored (required)
- [x] **DATA-03**: User note is stored (optional)
- [x] **DATA-04**: Page/location is stored
- [x] **DATA-05**: Chapter is stored (optional)
- [x] **DATA-06**: Timestamp from KOReader is stored
- [x] **DATA-07**: Highlight style (drawer) and color are stored
- [x] **DATA-08**: Device name is stored
- [x] **DATA-09**: Document MD5 hash is stored for book matching
- [x] **DATA-10**: Content hash for deduplication is stored

### Sync Behavior

- [x] **SYNC-01**: Re-syncing same highlights does not create duplicates
- [x] **SYNC-02**: Highlights for books not in library are stored (orphan handling)
- [x] **SYNC-03**: Both KOReader data models supported (annotations + legacy)

### Web UI

- [x] **UI-01**: Highlights displayed on book detail page
- [x] **UI-02**: Highlights shown with text, page, chapter (when available)
- [x] **UI-03**: User notes displayed alongside highlight text
- [ ] **UI-04**: Highlights ordered chronologically or by page
- [x] **UI-05**: Read-only display (no editing in web UI)

## v2 Requirements

Deferred to future release.

### Export

- **EXPR-01**: Export highlights to JSON
- **EXPR-02**: Bulk export all highlights

### Enhancement

- **ENH-01**: Filter highlights by color/style
- **ENH-02**: Highlight count badge on book cards
- **ENH-03**: Dedicated highlights page across all books

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Two-way sync (Kompanion → KOReader) | KOReader exporter plugin is push-only |
| Highlight editing in web UI | Read-only for MVP, edit in KOReader |
| Image highlights | Complexity, deferred per PROJECT.md |
| Real-time sync | Over-engineering for reading use case |
| Highlight sharing | Out of scope for self-hosted tool |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| API-01 | Phase 1 | Complete |
| API-02 | Phase 1 | Complete |
| API-03 | Phase 1 | Complete |
| API-04 | Phase 1 | Complete |
| DATA-01 | Phase 1 | Complete |
| DATA-02 | Phase 1 | Complete |
| DATA-03 | Phase 1 | Complete |
| DATA-04 | Phase 1 | Complete |
| DATA-05 | Phase 1 | Complete |
| DATA-06 | Phase 1 | Complete |
| DATA-07 | Phase 1 | Complete |
| DATA-08 | Phase 1 | Complete |
| DATA-09 | Phase 1 | Complete |
| DATA-10 | Phase 1 | Complete |
| SYNC-01 | Phase 1 | Complete |
| SYNC-02 | Phase 1 | Complete |
| SYNC-03 | Phase 1 | Complete |
| UI-01 | Phase 2 | Complete |
| UI-02 | Phase 2 | Complete |
| UI-03 | Phase 2 | Complete |
| UI-04 | Phase 2 | Pending |
| UI-05 | Phase 2 | Complete |

**Coverage:**
- v1 requirements: 21 total
- Mapped to phases: 21
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-21*
*Last updated: 2026-03-21 after initial definition*
