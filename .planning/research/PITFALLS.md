# Domain Pitfalls: Highlights Sync

**Domain:** KOReader to Kompanion highlights synchronization
**Researched:** 2026-03-21
**Confidence:** HIGH (source code analysis of KOReader exporter plugin and Kompanion sync implementation)

---

## Critical Pitfalls

Mistakes that cause rewrites, data corruption, or major issues.

### Pitfall 1: Duplicate Highlights on Re-Sync

**What goes wrong:** Every sync creates duplicate highlights in the database instead of updating existing ones.

**Why it happens:** KOReader sends ALL highlights every time, not just new ones. The exporter plugin has no concept of "sync state" - it just exports everything in the current document or history.

**Root cause from source analysis:**
```lua
-- From clip.lua:260-276 - parseAnnotations iterates ALL annotations
for _, item in ipairs(annotations) do
    -- Every annotation is exported every time
end
```

**Consequences:**
- Database bloat with duplicate entries
- User sees multiple copies of same highlight in UI
- De-duplication becomes a painful migration later

**Prevention:**
1. Generate a deterministic ID for each highlight using `document_hash + text_hash + page + timestamp`
2. Use PostgreSQL `INSERT ... ON CONFLICT DO NOTHING` or `ON CONFLICT UPDATE`
3. Create a unique constraint: `(book_id, content_hash, page, created_at)`

**Warning signs:**
- Highlights table grows faster than user reading rate
- Same highlight text appears multiple times in queries
- No unique constraint on highlights table in schema

**Phase mapping:** Phase 1 (API Implementation) - Must design ID strategy before first sync

---

### Pitfall 2: Timestamp Collision for Identity

**What goes wrong:** Using only `datetime` field as unique identifier causes missed highlights or false duplicates.

**Why it happens:** KOReader stores `datetime` as string like "2024-04-21 10:08:07". Multiple highlights created in rapid succession can have identical timestamps. Additionally, the timestamp parsing in KOReader is locale-dependent and fragile.

**Root cause from source analysis:**
```lua
-- From clip.lua:168-200 - getTime parsing is complex and locale-dependent
-- Multiple date formats are tried, some can fail silently
if not year or not month or not day then
    -- Fallback logic that may produce nil or duplicate values
end
```

**Consequences:**
- Highlights with same-second timestamps get merged or skipped
- Parsing failures create highlights with `nil` or `0` timestamps
- Timezone handling creates off-by-hour issues

**Prevention:**
1. Do NOT use timestamp as primary identifier
2. Create composite key: `(book_document_id, page, content_hash)`
3. Content hash = MD5 or SHA256 of `text + note` fields
4. Accept timestamp as metadata, not identity

**Warning signs:**
- Highlight with same text but different notes appears as one entry
- User reports "missing" highlights that were actually merged
- Queries show highlights with NULL or epoch timestamps

**Phase mapping:** Phase 1 (API Implementation) - Schema design must address this

---

### Pitfall 3: Orphan Highlights Without Book Match

**What goes wrong:** Highlights arrive for a book that doesn't exist in Kompanion's library, creating foreign key violations or orphan data.

**Why it happens:** KOReader can highlight any document, including sideloaded files not in Kompanion's library. The exporter sends `title` and `author` as strings, not IDs. Kompanion uses MD5 hash for book identification.

**Root cause from source analysis:**
```lua
-- From clip.lua:355-360 - Title/author extraction from file path
function MyClipping:getTitleAuthor(filepath, props)
    -- Falls back to parsing filename if metadata missing
    return isEmpty(props.title) and parsed_title or props.title,
           isEmpty(props.authors) and parsed_author or props.authors
end
```

**Consequences:**
- Highlights stored without valid `book_id` reference
- UI can't display highlights properly (no book info)
- Data integrity issues with foreign key constraints

**Prevention:**
1. Use KOReader's `partial_md5` (first 1MB of file) as `document` field - this IS sent in the sync
2. Match against Kompanion's `document_id` field in books table
3. Store highlights with nullable `book_id` - allow "unlinked" highlights
4. Provide UI to manually link orphan highlights to books

**Warning signs:**
- Foreign key violation errors during sync
- Highlights stored with NULL book_id
- Users report highlights "disappearing" for certain books

**Phase mapping:** Phase 1 (API Implementation) - Handle book resolution before storing

---

### Pitfall 4: Missing `note` Field Handling

**What goes wrong:** Highlights with notes are stored without the note content, or notes are stored incorrectly.

**Why it happens:** KOReader has TWO data models:
1. **Annotations** (newer): `item.note` directly on annotation
2. **Highlights + Bookmarks** (legacy): Note stored in separate bookmarks table, matched by datetime

**Root cause from source analysis:**
```lua
-- From clip.lua:260-276 - New annotations model
clipping = {
    note = item.note and self:getText(item.note), -- Direct field
}

-- From clip.lua:279-316 - Legacy highlights + bookmarks model
-- Notes must be matched from bookmarks by datetime:
if bookmark.datetime == item.datetime then
    clipping.note = bookmark_quote or bookmark.text
end
```

**Consequences:**
- Notes lost for users on older KOReader versions
- Empty `note` fields when bookmarks not matched
- Incorrect note content when bookmark text differs from highlight

**Prevention:**
1. Request KOReader send both `annotations` format (preferred) and handle legacy
2. Store note as nullable text field, don't require it
3. Log warning when highlight has no note but expected one

**Warning signs:**
- User reports missing notes for highlights
- Notes appearing in wrong highlights
- Empty note field when KOReader shows note exists

**Phase mapping:** Phase 1 (API Implementation) - Must handle both data models

---

### Pitfall 5: Character Encoding Corruption

**What goes wrong:** Non-ASCII characters in highlight text (accents, CJK, emojis) are corrupted during sync.

**Why it happens:** KOReader uses UTF-8, but JSON encoding/decoding can corrupt if:
- HTTP headers don't specify charset
- Database column isn't UTF-8
- Go string handling assumes ASCII

**Root cause from source analysis:**
```lua
-- From base.lua:167 - JSON encoding uses rapidjson
body_json, err = rapidjson.encode(body)
-- rapidjson handles UTF-8 correctly, but receiver might not
```

**Consequences:**
- Mojibake in stored highlights (e.g., "cafÃ©" instead of "cafe")
- CJK characters become question marks or garbage
- Database errors on invalid UTF-8 sequences

**Prevention:**
1. Ensure PostgreSQL database uses UTF-8 encoding: `ENCODING 'UTF8'`
2. Set Content-Type header: `application/json; charset=utf-8`
3. Use Go's `json.Marshal` which handles UTF-8 correctly
4. Test with non-ASCII characters: accented, CJK, emoji

**Warning signs:**
- Highlight text contains replacement characters ()
- Database errors: "invalid byte sequence for encoding"
- User reports with specific example of corrupted text

**Phase mapping:** Phase 1 (API Implementation) - Test early with non-ASCII

---

## Moderate Pitfalls

### Pitfall 6: Large Payload Timeout

**What goes wrong:** Sync requests with many highlights (heavy readers) timeout or fail.

**Why it happens:** KOReader sends ALL highlights for a book, not incremental changes. A heavily annotated book can have hundreds of highlights.

**Root cause from source analysis:**
```lua
-- From base.lua:173 - Uses LARGE_BLOCK_TIMEOUT
socketutil:set_timeout(socketutil.LARGE_BLOCK_TIMEOUT, socketutil.LARGE_TOTAL_TIMEOUT)
```

**Prevention:**
1. Set server timeout appropriately (30-60 seconds)
2. Process highlights in batches if count > 100
3. Return 202 Accepted for large payloads, process async

**Warning signs:**
- Sync fails for users with many highlights
- Timeout errors in server logs
- Successful syncs take >10 seconds

**Phase mapping:** Phase 1 (API Implementation) - Set timeout config

---

### Pitfall 7: Chapter Information Loss

**What goes wrong:** Chapter context for highlights is lost, making it hard to find where a highlight came from.

**Why it happens:** The `chapter` field is optional and populated differently by different document formats. PDFs often lack chapter metadata.

**Root cause from source analysis:**
```lua
-- From clip.lua:269 - Chapter is optional
chapter = item.chapter, -- Can be nil
```

**Prevention:**
1. Store chapter as nullable field
2. Don't require chapter for storage
3. UI should handle highlights without chapter gracefully

**Phase mapping:** Phase 2 (UI Display) - Design UI to handle missing chapter

---

### Pitfall 8: Page Number Inconsistency

**What goes wrong:** Page numbers change between syncs or don't match the actual book page numbers.

**Why it happens:** KOReader has multiple page number concepts:
- `page` (display page)
- `pn_xp` (internal position)
- `pageref` or `pageno` (different annotations models)

**Root cause from source analysis:**
```lua
-- From clip.lua:265 - Multiple page sources
page = item.pageref or item.pageno, -- Newer annotations
-- vs
page = page, -- Legacy: key in highlights table
```

**Prevention:**
1. Store both `display_page` and `position` fields
2. Document which page source is used
3. Don't assume page numbers are stable across book versions

**Phase mapping:** Phase 1 (API Implementation) - Store multiple page fields

---

### Pitfall 9: Network Error Retry Loop

**What goes wrong:** Transient network errors cause duplicate highlights when user retries sync.

**Why it happens:** KOReader doesn't know if server received the request. User taps "Export" again, and server receives duplicate data.

**Prevention:**
1. Make sync endpoint idempotent (see Pitfall 1)
2. Return clear success/error responses
3. KOReader side: implement request-level idempotency key (not in current code)

**Phase mapping:** Phase 1 (API Implementation) - Idempotency is essential

---

### Pitfall 10: Filter Settings Ignored

**What goes wrong:** Server receives highlights that user filtered out (e.g., only yellow highlights), but stores them anyway.

**Why it happens:** KOReader has client-side filtering by style/color, but server doesn't know about these preferences.

**Root cause from source analysis:**
```lua
-- From clip.lua:251-258 - Client-side filter
function MyClipping:doesHighlightMatch(item)
    if filter.style and not filter.style[item.drawer] then return end
    if filter.color and not filter.color[item.color] then return end
    return true
end
```

**Prevention:**
1. Trust KOReader's filter - only receive what it sends
2. Optionally store style/color metadata for UI filtering
3. Don't re-implement filtering on server side

**Phase mapping:** Phase 2 (UI Display) - Consider storing style/color for UI

---

## Minor Pitfalls

### Pitfall 11: Highlight Style/Color Metadata Ignored

**What goes wrong:** `drawer` (underline vs highlight) and `color` fields are discarded, losing user's visual organization.

**Prevention:**
- Store `style` and `color` as optional metadata
- UI can filter/highlight by color

**Phase mapping:** Phase 2 (UI Display) - Nice-to-have metadata

---

### Pitfall 12: Image Highlights Not Supported

**What goes wrong:** Highlights of images (in PDFs with reflow) are lost or cause errors.

**Why it happens:** KOReader can highlight images, but the project scope explicitly defers image highlights.

**Prevention:**
- Log warning when image highlight received
- Store placeholder indicating image highlight exists
- Defer full implementation

**Phase mapping:** Out of scope per PROJECT.md

---

### Pitfall 13: Sync Progress vs Highlights Confusion

**What goes wrong:** Developers confuse progress sync endpoint (`/syncs/progress`) with highlights sync, causing architectural issues.

**Why it happens:** Both involve KOReader sync, but different data models and timing.

**Prevention:**
- Clear naming: `/syncs/progress` for reading position, `/syncs/highlights` for annotations
- Separate tables: `sync_progress` vs `book_highlights`
- Document the distinction clearly

**Phase mapping:** Phase 1 (API Implementation) - Use distinct endpoint/table names

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| API Design | Duplicate on re-sync (Pitfall 1) | Idempotent design with content hash |
| API Design | Timestamp collision (Pitfall 2) | Composite key, not timestamp-only |
| API Design | Orphan highlights (Pitfall 3) | Handle unmatched books gracefully |
| API Design | Missing notes (Pitfall 4) | Support both KOReader data models |
| Data Storage | Encoding corruption (Pitfall 5) | UTF-8 everywhere, test early |
| Error Handling | Network retry duplicates (Pitfall 9) | Idempotent endpoints |
| UI Display | Missing chapter (Pitfall 7) | Nullable field, graceful UI |
| UI Display | Page inconsistency (Pitfall 8) | Store position, not just page |

---

## Sources

- **KOReader exporter plugin source code** (HIGH confidence)
  - `/home/deploy/koreader/plugins/exporter.koplugin/clip.lua` - Highlight parsing logic
  - `/home/deploy/koreader/plugins/exporter.koplugin/base.lua` - JSON request handling
  - `/home/deploy/koreader/plugins/exporter.koplugin/main.lua` - Export coordination
  - `/home/deploy/koreader/plugins/exporter.koplugin/target/readwise.lua` - Remote API example
  - `/home/deploy/koreader/plugins/exporter.koplugin/target/nextcloud.lua` - Another remote API with duplicate handling

- **Kompanion existing sync implementation** (HIGH confidence)
  - `/home/deploy/kompanion/internal/sync/progress.go` - Progress sync use case
  - `/home/deploy/kompanion/internal/sync/progress_postgres.go` - Storage pattern
  - `/home/deploy/kompanion/internal/controller/http/v1/sync.go` - HTTP handler pattern
  - `/home/deploy/kompanion/migrations/20250211190954_sync.up.sql` - Schema pattern

- **Project requirements** (HIGH confidence)
  - `/home/deploy/kompanion/.planning/PROJECT.md` - Scope and constraints
