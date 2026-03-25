# Feature Landscape

**Domain:** Book Highlights Sync System
**Researched:** 2026-03-21

## Table Stakes

Features users expect. Missing = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **HTTP API endpoint for receiving highlights** | Standard integration pattern, matches existing progress sync | Low | POST endpoint similar to existing `/syncs/progress` |
| **Store highlight text** | Core data - users want to see what they highlighted | Low | Text field, potentially long |
| **Store page/location** | Essential context for where highlight appeared | Low | Integer or string for flexibility |
| **Store timestamp** | Ordering, deduplication, sync conflict resolution | Low | Unix timestamp from KOReader |
| **Associate highlight with book** | Must link highlights to specific books | Low | Foreign key to books table via document ID (MD5) |
| **Associate highlight with device** | Track which device sent the highlight | Low | Device name from auth context |
| **Display highlights on book detail page** | Users want to review their highlights in context | Medium | Integrate into existing web UI template |
| **Handle user notes attached to highlights** | KOReader supports adding notes to highlights | Low | Optional text field |
| **Authentication via device credentials** | Matches existing progress sync pattern | Low | MD5-hashed device credentials |
| **Idempotent sync (deduplication)** | Re-syncing should not create duplicates | Medium | Hash text + page + timestamp |

## Differentiators

Features that set product apart. Not expected, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Chapter information storage** | Contextual organization, better browsing | Low | Optional chapter title from KOReader |
| **Highlight color/style preservation** | Visual organization, matches KOReader experience | Low | Store drawer and color fields |
| **Bulk highlight sync** | Efficient for users with many highlights | Medium | Accept array of highlights in one request |
| **Sync timestamp tracking** | Know when last sync occurred per book | Low | Store `last_highlight_at` per book |
| **Highlight count on book cards** | Quick indicator of annotated books | Low | Display count badge on book list |
| **Chronological ordering** | Browse highlights in reading order | Low | Order by location/page or timestamp |
| **Filter highlights by style/color** | Find specific types of highlights | Medium | UI filtering capability |
| **Export highlights to JSON** | Portability, backup, integration with other tools | Low | Leverage existing JSON export pattern |

## Anti-Features

Features to explicitly NOT build.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **Two-way sync (Kompanion -> KOReader)** | KOReader exporter plugin architecture is push-only; implementing fetch would require KOReader core changes | One-way sync only (KOReader -> Kompanion) |
| **Highlight editing in web UI** | Adds significant complexity; editing could cause sync conflicts | Read-only display; edit in KOReader only |
| **Highlight export from web UI** | KOReader has robust export features; avoid duplication | Users export directly from KOReader |
| **Image highlights** | Requires binary storage, significantly more complex | Text-only highlights for MVP |
| **Real-time sync (WebSocket/push)** | Over-engineering for reading use case; battery drain on e-ink | On-demand sync when KOReader initiates |
| **Highlight sharing to social media** | Out of scope for self-hosted personal tool | Export to JSON if user wants to share |
| **OCR for image-based PDFs** | Extremely complex; many edge cases | Text-based highlights only |
| **Highlight tagging system** | Adds data model complexity; KOReader has limited tagging | Rely on color/style for categorization |

## Feature Dependencies

```
HTTP API endpoint -> Store highlight text (API must accept data)
Store highlight text -> Associate with book (need book reference)
Associate with book -> Display on book detail page (need data to display)
Authentication -> All sync operations (security foundation)
```

## Data Model from KOReader

Based on KOReader exporter plugin source code analysis (`plugins/exporter.koplugin/clip.lua`):

```lua
-- Each clipping/highlight contains:
{
    sort    = "highlight",     -- type of annotation
    page    = 123,             -- page number or location
    time    = 1398127554,      -- Unix timestamp
    text    = "highlighted text", -- the actual highlighted text
    note    = "user note",     -- optional user annotation
    chapter = "Chapter I",     -- optional chapter title
    drawer  = "lighten",       -- highlight style (underline, lighten, etc.)
    color   = "yellow",        -- highlight color
    pn_xp   = "1/100",         -- precise position in document
}
```

The `booknotes` structure from KOReader:
```lua
{
    title = "Book Title",
    author = "Author Name",
    file = "/path/to/book.epub",
    number_of_pages = 300,
    [1] = { {clipping1}, {clipping2}, ... }, -- chapter 1 highlights
    [2] = { {clipping3}, ... },               -- chapter 2 highlights
}
```

## MVP Recommendation

Prioritize:
1. **HTTP API endpoint** - Core infrastructure, matches existing `/syncs/progress` pattern
2. **Store highlights in PostgreSQL** - Text, page, timestamp, book association, device
3. **Display highlights on book detail page** - Read-only list view

Defer:
- **Bulk sync optimization**: Start with single-highlight sync, optimize later
- **Chapter/color preservation**: Nice to have but not critical for MVP
- **Highlight filtering**: Add when highlight volume justifies it

## Sources

- KOReader exporter plugin source code (local: `/home/deploy/koreader/plugins/exporter.koplugin/`)
  - `main.lua` - Plugin architecture, menu integration
  - `base.lua` - BaseExporter class with `makeJsonRequest` for HTTP APIs
  - `clip.lua` - Clipping parser, data structure definitions
  - `target/readwise.lua` - Reference HTTP API integration
  - `target/json.lua` - JSON export format reference
- Readwise API documentation (https://readwise.io/api_deets) - Industry standard for highlight APIs
- Kompanion existing codebase:
  - `internal/sync/progress.go` - Progress sync use case pattern
  - `internal/sync/progress_postgres.go` - PostgreSQL repository pattern
  - `internal/controller/http/v1/sync.go` - HTTP route pattern
  - `internal/entity/progress.go` - Entity structure pattern
  - `.planning/PROJECT.md` - Project requirements and constraints

## Confidence Assessment

| Area | Confidence | Reason |
|------|------------|--------|
| KOReader data structure | HIGH | Analyzed source code directly |
| API pattern | HIGH | Existing progress sync provides clear template |
| Table stakes | HIGH | Based on Readwise API and KOReader capabilities |
| Anti-features | HIGH | Confirmed by PROJECT.md constraints |
| Differentiators | MEDIUM | Based on industry analysis, may need user validation |
