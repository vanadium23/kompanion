# Phase 5: Standalone KOReader Plugin with Native Highlight Extraction - Context

**Gathered:** 2026-03-22
**Status:** Ready for planning

<domain>
## Phase Boundary

Create a standalone KOReader Lua plugin that directly extracts highlights from document sidecar files and syncs to Kompanion. Does NOT depend on the exporter.koplugin framework — reads DocSettings directly using the clip.lua patterns.

**Why this phase exists:** Phase 4's Provider-based plugin doesn't work reliably with the exporter system. This phase creates a self-contained solution.

**Key constraint:** Plugin must work without modifying KOReader core code.

</domain>

<decisions>
## Implementation Decisions

### Plugin Architecture
- **D-01:** Standalone plugin under Tools menu (not Export highlights submenu)
- **D-02:** WidgetContainer-based plugin, NOT BaseExporter subclass
- **D-03:** Uses `clip.lua` patterns to read highlights directly from DocSettings

### Sync Behavior
- **D-04:** Manual trigger only — "Sync highlights" menu item
- **D-05:** Current book only — reads from current document's sidecar file
- **D-06:** Syncs to existing `/syncs/highlights` API endpoint

### Highlight Extraction
- **D-07:** Read `annotations` setting first (newer KOReader format)
- **D-08:** Fallback to `highlight` + `bookmarks` settings (legacy format)
- **D-09:** Get document hash from `partial_md5_checksum` in DocSettings
- **D-10:** Transform to Kompanion API format matching Phase 1 contract

### Configuration UI
- **D-11:** Three menu items: Setup, Sync Now, Help
- **D-12:** Setup dialog with URL, Device Name, Device Password fields
- **D-13:** Settings stored in `G_reader_settings:readSetting("kompanion")`

### Error Handling
- **D-14:** Show InfoMessage toast on success with synced count
- **D-15:** Show InfoMessage toast on failure with error message
- **D-16:** Log errors to KOReader logger for debugging

### Claude's Discretion
- Exact toast message text
- Plugin version number
- Help text content
- HTTP timeout values

</decisions>

<specifics>
## Specific Ideas

- "I want a simple plugin that just works without the exporter dependency"
- Should feel like kosync.koplugin — straightforward Tools menu integration
- User shouldn't need to understand Provider systems or exporter frameworks

</specifics>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### KOReader Highlight Storage
- `/home/deploy/koreader/plugins/exporter.koplugin/clip.lua` — MyClipping class showing how to extract highlights from DocSettings (lines 362-385: getClippingsFromBook, parseAnnotations, parseHighlight)
- `/home/deploy/koreader/frontend/ui/widget/booklist.lua` — BookList.getDocSettings() for opening doc settings

### KOReader Plugin Patterns
- `/home/deploy/koreader/plugins/kosync.koplugin/main.lua` — Standalone plugin pattern with Tools menu integration, settings, HTTP API calls
- `/home/deploy/koreader/plugins/kosync.koplugin/_meta.lua` — Plugin metadata format

### Kompanion API
- `internal/controller/http/v1/highlight.go` — Highlight sync endpoint, request/response format
- `internal/entity/highlight.go` — Highlight entity structure (fields: text, note, page, chapter, time, drawer, color)

### Existing Non-Working Plugin (DO NOT COPY)
- `koreader/kompanion.koplugin/target.lua` — Previous attempt using BaseExporter/Provider system — this approach doesn't work

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `DocSettings:open(file):readSetting("annotations")` — New format for highlights
- `DocSettings:open(file):readSetting("highlight")` — Legacy format
- `DocSettings:open(file):readSetting("bookmarks")` — Notes linked to highlights
- `DocSettings:open(file):readSetting("partial_md5_checksum")` — Document hash
- `DocSettings:open(file):readSetting("doc_props")` — Title, author metadata

### Established Patterns (from kosync.koplugin)
- `WidgetContainer` base class for standalone plugins
- `UIManager:show(InfoMessage:new{...})` for toasts
- `MultiInputDialog` for setup dialogs
- `G_reader_settings:readSetting("plugin_name")` for persistent settings
- `socket.http.request` for HTTP calls with Basic Auth

### Integration Points
- Register plugin in Tools menu via `UIManager.menuItems:register()`
- Get current document from `UIManager:getCurrentInstance()` or similar
- No dependency on exporter.koplugin whatsoever

</code_context>

<deferred>
## Deferred Ideas

- Auto-sync on book close — could be future enhancement
- Sync all books from history — out of scope for now
- Sync progress back to KOReader — out of scope (one-way sync per PROJECT.md)
- Sync status screen with history — keep it minimal

</deferred>

---

*Phase: 05-standalone-koreader-plugin-with-native-highlight-extraction*
*Context gathered: 2026-03-22*
