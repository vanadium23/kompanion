# Phase 3: Nextcloud Notes API Endpoint - Research

**Researched:** 2026-03-21
**Domain:** Nextcloud Notes API compatibility for KOReader exporter
**Confidence:** HIGH

## Summary

This phase implements a Nextcloud Notes API-compatible endpoint that allows KOReader's built-in Nextcloud Notes exporter to sync highlights to Kompanion without modification. The implementation requires HTTP Basic Authentication (different from existing device auth middleware), three API endpoints (GET/POST/PUT), and markdown formatting that matches KOReader's expected output format.

**Primary recommendation:** Reuse existing `basicAuth` middleware pattern from OPDS router, create new route group at `/index.php/apps/notes/api/v1/notes`, and format highlights using the established markdown template from KOReader's md.lua.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**API Format Compatibility**
- D-01: Implement full Nextcloud Notes API: `GET /notes`, `POST /notes`, `PUT /notes/{id}`
- D-02: One note per book containing all highlights (matches KOReader's Nextcloud exporter behavior)
- D-03: Support batch export of multiple books in one session
- D-04: URL path: `/index.php/apps/notes/api/v1/notes` (full Nextcloud compatibility)

**Content Format**
- D-05: Group highlights by chapter with `## Chapter Name` headings
- D-06: Metadata per highlight: Page/location + Timestamp
- D-07: Visual format: Blockquote (`>`) for highlight text, plain text for user note
- D-08: Separator between highlights: `---` horizontal rule

**Authentication**
- D-09: Use HTTP Basic Auth (username:password in Authorization header)
- D-10: Map KOReader devices to Kompanion users via existing `devices` table
- D-11: Device name from Basic Auth username maps to `koreader_device_id`
- D-12: Validate credentials against device password (MD5 hash, existing pattern)

**Storage & Updates**
- D-13: Reuse existing `highlight_annotations` table, filtered by device
- D-14: Document hash (partial_md5) used as identifier for update detection
- D-15: On PUT: Replace note content entirely (no partial merge)
- D-16: Category parameter: Accept but store as metadata (category grouping is client-side)

### Claude's Discretion

- Exact markdown template implementation
- Error response format and status codes
- Category handling in responses
- Timestamp formatting in markdown output

### Deferred Ideas (OUT OF SCOPE)

- Note sharing between devices — future enhancement
- Note editing in web UI — out of scope (read-only display)
- Rich text formatting beyond markdown — not in Nextcloud Notes spec
- Image highlights — explicitly out of scope per PROJECT.md

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| NC-01 | GET /notes endpoint returns all notes for device | Nextcloud API spec section + KOReader nextcloud.lua lines 81-91 |
| NC-02 | POST /notes creates new note from highlights | KOReader nextcloud.lua lines 93-111 for expected request format |
| NC-03 | PUT /notes/{id} updates existing note | KOReader nextcloud.lua lines 113-133 for update detection logic |
| NC-04 | HTTP Basic Auth validates device credentials | Existing `basicAuth` middleware in opds/router.go |
| NC-05 | Markdown format matches KOReader expectations | md.lua template for exact formatting structure |
| NC-06 | Note ID based on document hash | partial_md5 field used as stable identifier |

</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| gin-gonic/gin | v1.7.7 | HTTP routing, middleware | Already in use, route groups pattern |
| jackc/pgx/v5 | v5.6.0 | PostgreSQL access | Existing highlight repository uses this |
| stretchr/testify | v1.11.1 | Testing | Existing test pattern |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| encoding/base64 | stdlib | Basic Auth decoding | Parsing Authorization header |
| encoding/json | stdlib | API response encoding | Nextcloud API JSON responses |
| strconv | stdlib | Integer ID parsing | Extracting note ID from URL path |

### Existing Code to Reuse
| Component | Location | Purpose |
|-----------|----------|---------|
| `basicAuth` middleware | `internal/controller/http/opds/router.go` | HTTP Basic Auth pattern |
| `CheckDevicePassword` | `internal/auth/auth.go` | Device credential validation |
| `Highlight` interface | `internal/highlight/sync.go` | Sync and Fetch methods |
| `HighlightRepo` | `internal/highlight/highlight_postgres.go` | Database operations |

**No new dependencies required.**

## Architecture Patterns

### Recommended Project Structure

```
internal/
├── controller/http/
│   └── v1/
│       └── notes.go           # NEW: Nextcloud Notes API handlers
├── notes/                     # NEW: Notes formatting logic
│   ├── formatter.go           # Highlight to markdown conversion
│   └── formatter_test.go      # Markdown output tests
└── highlight/                 # EXISTING: Reuse for data access
    ├── sync.go                # Sync/Fetch methods
    └── highlight_postgres.go  # Repository
```

### Pattern 1: Route Group with Basic Auth

**What:** Separate route group for Nextcloud API path with Basic Auth middleware
**When to use:** Nextcloud Notes API endpoints
**Example:**
```go
// From opds/router.go - reuse this pattern
func basicAuth(auth auth.AuthInterface) gin.HandlerFunc {
    return func(c *gin.Context) {
        username, password, ok := c.Request.BasicAuth()
        if !ok {
            c.Header("WWW-Authenticate", `Basic realm="KOmpanion Notes"`)
            c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized", "code": 2001})
            c.Abort()
            return
        }
        if !auth.CheckDevicePassword(c.Request.Context(), username, password, true) {
            c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized", "code": 2001})
            c.Abort()
            return
        }
        c.Set("device_name", username)
        c.Next()
    }
}

// In router.go
notesRoutes := handler.Group("/index.php/apps/notes/api/v1")
notesRoutes.Use(basicAuth(a))
newNotesRoutes(notesRoutes, h, l)
```

### Pattern 2: Note Response Structure

**What:** JSON response format matching Nextcloud Notes API
**When to use:** All API responses
**Example:**
```go
type NoteResponse struct {
    ID           int    `json:"id"`           // Document hash as integer suffix or unique ID
    Etag         string `json:"etag"`         // Content hash for caching
    ReadOnly     bool   `json:"readonly"`     // Always false for our use case
    Content      string `json:"content"`      // Markdown-formatted highlights
    Title        string `json:"title"`        // "{author} - {title}"
    Category     int    `json:"category"`     // From request or 0
    Favorite     bool   `json:"favorite"`     // Always false
    Modified     int64  `json:"modified"`     // Unix timestamp
}

// Title format from KOReader nextcloud.lua line 76
// string.format("%s - %s", string.gsub(booknotes.author, "\n", ", "), booknotes.title)
```

### Pattern 3: Markdown Formatting

**What:** Convert highlights to markdown matching KOReader's md.lua template
**When to use:** Generating note content for GET/POST/PUT responses
**Example:**
```go
// Based on md.lua template structure
func FormatHighlights(title, author string, highlights []entity.Highlight) string {
    var sb strings.Builder

    // Header
    sb.WriteString("# " + title + "\n")
    author = strings.ReplaceAll(author, "\n", ", ")
    sb.WriteString("##### " + author + "\n\n")

    // Group by chapter
    currentChapter := ""
    for _, hl := range highlights {
        if hl.Chapter != currentChapter && hl.Chapter != "" {
            currentChapter = hl.Chapter
            sb.WriteString("## " + currentChapter + "\n")
        }

        // Page and timestamp
        timestamp := time.Unix(hl.Timestamp, 0).Format("02 January 2006 03:04:05 PM")
        sb.WriteString("### Page " + hl.Page + " @ " + timestamp + "\n")

        // Highlight text with drawer formatting (blockquote)
        sb.WriteString("> " + hl.Text + "\n")

        // Note if present
        if hl.Note != "" {
            sb.WriteString("\n---\n" + hl.Note + "\n")
        }
        sb.WriteString("\n")
    }

    return sb.String()
}
```

### Anti-Patterns to Avoid

- **Using authDeviceMiddleware instead of basicAuth:** authDeviceMiddleware uses x-auth-user/x-auth-key headers, but KOReader's Nextcloud exporter sends Basic Auth. Must use `c.Request.BasicAuth()` pattern.
- **Storing notes separately:** Notes are views over existing highlights. Don't create a new notes table.
- **Partial merge on PUT:** KOReader replaces entire note content. Don't implement incremental updates.
- **Ignoring category parameter:** Accept it in POST/PUT but don't filter on GET unless explicitly requested.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Basic Auth parsing | Custom header parsing | `c.Request.BasicAuth()` | Standard library handles base64 decoding |
| Device password validation | New auth logic | `auth.CheckDevicePassword(ctx, username, password, true)` | Existing MD5 hash comparison |
| Note ID generation | UUID or random | `partial_md5` as stable identifier | KOReader uses document hash for update detection |
| Markdown escaping | Manual string building | `strings.Builder` with careful escaping | Avoid injection issues |

**Key insight:** The Notes API is a read/write view over the existing highlights table. No new storage required.

## Common Pitfalls

### Pitfall 1: Wrong Authentication Method

**What goes wrong:** Using `authDeviceMiddleware` which expects `x-auth-user` and `x-auth-key` headers instead of HTTP Basic Auth.

**Why it happens:** Both are device authentication, but different protocols. The existing highlight sync uses headers, Nextcloud uses Basic Auth.

**How to avoid:** Use `c.Request.BasicAuth()` from the standard library. KOReader sends:
```lua
local auth = mime.b64(self.settings.username .. ":" .. self.settings.password)
["Authorization"] = "Basic " .. auth
```

**Warning signs:** 401 Unauthorized responses from KOReader even with correct device credentials.

### Pitfall 2: Missing OCS-APIRequest Header

**What goes wrong:** KOReader sends `OCS-APIRequest: true` header. Some Nextcloud implementations require this.

**Why it happens:** The header is part of Nextcloud's OCS API specification.

**How to avoid:** Our implementation should accept but can ignore this header. Don't require it for compatibility.

**Warning signs:** KOReader logs showing "401 Unauthorized" despite correct credentials.

### Pitfall 3: Note ID as Integer vs String

**What goes wrong:** Using string document hashes as IDs when Nextcloud expects integers.

**Why it happens:** The API spec shows `"id": 123` but we want to use document hashes.

**How to avoid:** Generate a deterministic integer from partial_md5 hash:
```go
// Use CRC32 or similar to convert hash to int32
func hashToInt(hash string) int {
    h := crc32.ChecksumIEEE([]byte(hash))
    return int(h)
}
```

**Warning signs:** KOReader failing to update existing notes (treats all as new).

### Pitfall 4: Timestamp Format Mismatch

**What goes wrong:** Using different timestamp formats than expected.

**Why it happens:** KOReader uses `os.date("%d %B %Y %I:%M:%S %p", entry.time)` which is locale-dependent.

**How to avoid:** Use Go's time formatting with explicit format string:
```go
timestamp := time.Unix(hl.Timestamp, 0).Format("02 January 2006 03:04:05 PM")
```

**Warning signs:** Timestamps appearing in wrong format in markdown output.

### Pitfall 5: Update Detection Without Proper ID

**What goes wrong:** POST always creates new notes instead of updating existing ones.

**Why it happens:** KOReader checks for existing note by title before deciding POST vs PUT (see nextcloud.lua lines 81-91).

**How to avoid:** GET /notes must return notes with titles matching `{author} - {title}` format so KOReader can detect existing notes. Use consistent title formatting.

**Warning signs:** Duplicate notes appearing after multiple exports of same book.

## Code Examples

### GET /notes Handler

```go
// Source: Based on Nextcloud API spec and KOReader nextcloud.lua
func (r *notesRoutes) listNotes(c *gin.Context) {
    deviceName := c.GetString("device_name")
    category := c.Query("category") // Optional filter

    // Get all highlights grouped by document
    documents, err := r.highlight.GetDocumentsByDevice(c, deviceName)
    if err != nil {
        r.l.Error(err)
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal error"})
        return
    }

    notes := []NoteResponse{}
    for _, doc := range documents {
        highlights, _ := r.highlight.Fetch(c, doc.PartialMD5)
        content := FormatHighlights(doc.Title, doc.Author, highlights)

        notes = append(notes, NoteResponse{
            ID:       hashToInt(doc.PartialMD5),
            Title:    formatTitle(doc.Author, doc.Title),
            Content:  content,
            Modified: getLatestModified(highlights),
            Category: 0,
            Favorite: false,
            ReadOnly: false,
            Etag:     computeEtag(content),
        })
    }

    c.JSON(http.StatusOK, notes)
}
```

### POST /notes Handler

```go
// Source: KOReader nextcloud.lua lines 93-111
func (r *notesRoutes) createNote(c *gin.Context) {
    deviceName := c.GetString("device_name")

    var req struct {
        Title    string `json:"title"`
        Content  string `json:"content"`
        Category int    `json:"category"`
        Favorite bool   `json:"favorite"`
        Modified int64  `json:"modified"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
        return
    }

    // Parse title to extract author and book title
    // Format: "{author} - {title}"
    author, title := parseTitle(req.Title)

    // Store highlights (content already formatted by KOReader)
    // Note: Content is markdown, we may need to parse it back to highlights
    // OR we can store the raw content and return it as-is

    note := NoteResponse{
        ID:       generateID(),
        Title:    req.Title,
        Content:  req.Content,
        Modified: req.Modified,
        Category: req.Category,
        Favorite: req.Favorite,
        ReadOnly: false,
        Etag:     computeEtag(req.Content),
    }

    c.JSON(http.StatusOK, note)
}
```

### PUT /notes/:id Handler

```go
// Source: KOReader nextcloud.lua lines 113-133
func (r *notesRoutes) updateNote(c *gin.Context) {
    noteID := c.Param("id")
    deviceName := c.GetString("device_name")

    var req struct {
        Title    string `json:"title"`
        Content  string `json:"content"`
        Category int    `json:"category"`
        Favorite bool   `json:"favorite"`
        Modified int64  `json:"modified"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
        return
    }

    // Find document by note ID (reverse hash lookup)
    // Replace note content entirely

    note := NoteResponse{
        ID:       parseNoteID(noteID),
        Title:    req.Title,
        Content:  req.Content,
        Modified: req.Modified,
        Category: req.Category,
        Favorite: req.Favorite,
        ReadOnly: false,
        Etag:     computeEtag(req.Content),
    }

    c.JSON(http.StatusOK, note)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| XMNote API (no auth) | Nextcloud Notes API with Basic Auth | 2026-03-21 | Security improvement, standard protocol |
| Custom sync endpoint | Standard API compatibility | Phase 3 | Works with existing KOReader exporter |

**Deprecated/outdated:**
- XMNote API approach: No authentication mechanism, unsuitable for multi-user system

## Open Questions

1. **Should we parse markdown content back to highlights?**
   - What we know: KOReader sends formatted markdown content
   - What's unclear: Do we need to extract individual highlights for storage, or just store the markdown blob?
   - Recommendation: Store markdown blob for simplicity. Highlights are already stored via the existing `/syncs/highlights` endpoint. Notes API is a read/write view over that data.

2. **How to handle note ID generation?**
   - What we know: Nextcloud uses integers, we have partial_md5 strings
   - What's unclear: Best way to convert hash to stable integer ID
   - Recommendation: Use CRC32 checksum of partial_md5 for deterministic integer ID.

3. **Category filtering implementation?**
   - What we know: KOReader sends category parameter, Nextcloud API supports filtering
   - What's unclear: Do we need to implement category storage and filtering?
   - Recommendation: Accept category in POST/PUT, return it in GET, but don't implement filtering unless requested. KOReader doesn't appear to use category filtering.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | testify + go-hit |
| Config file | None - tests in `*_test.go` files |
| Quick run command | `go test ./internal/controller/http/v1/... -run TestNotes -v` |
| Full suite command | `go test ./... -v` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| NC-01 | GET /notes returns device notes | integration | `go test ./internal/controller/http/v1 -run TestNotesList -v` | Wave 0 |
| NC-02 | POST /notes creates note | integration | `go test ./internal/controller/http/v1 -run TestNotesCreate -v` | Wave 0 |
| NC-03 | PUT /notes/:id updates note | integration | `go test ./internal/controller/http/v1 -run TestNotesUpdate -v` | Wave 0 |
| NC-04 | Basic Auth validates device | unit | `go test ./internal/auth -run TestCheckDevicePassword -v` | Exists |
| NC-05 | Markdown formatting matches expected | unit | `go test ./internal/notes -run TestFormatHighlights -v` | Wave 0 |
| NC-06 | Note ID from document hash | unit | `go test ./internal/notes -run TestHashToInt -v` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/notes/... -v`
- **Per wave merge:** `go test ./internal/controller/http/v1/... -v`
- **Phase gate:** `go test ./... -v` full suite green

### Wave 0 Gaps
- [ ] `internal/controller/http/v1/notes_test.go` - Notes API handler tests
- [ ] `internal/notes/formatter.go` - Markdown formatting logic
- [ ] `internal/notes/formatter_test.go` - Formatting unit tests
- [ ] Repository method `GetDocumentsByDevice` in highlight_postgres.go

## Sources

### Primary (HIGH confidence)
- `/home/deploy/koreader/plugins/exporter.koplugin/target/nextcloud.lua` - KOReader's Nextcloud exporter implementation (exact API expectations)
- `/home/deploy/koreader/plugins/exporter.koplugin/template/md.lua` - Markdown formatting template
- `internal/controller/http/opds/router.go` - Existing Basic Auth middleware pattern
- `internal/auth/auth.go` - Device password validation implementation
- Nextcloud Notes API v1 specification (GitHub) - Official API spec

### Secondary (MEDIUM confidence)
- `internal/highlight/sync.go` - Existing highlight sync logic to reuse
- `internal/highlight/highlight_postgres.go` - Repository pattern to extend
- `migrations/20260321120000_highlights.up.sql` - Database schema

### Tertiary (LOW confidence)
- None - all critical information verified from primary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - No new dependencies, reuses existing patterns
- Architecture: HIGH - Clear existing patterns to follow, KOReader code is explicit
- Pitfalls: HIGH - Identified from actual KOReader implementation

**Research date:** 2026-03-21
**Valid until:** 30 days - API spec is stable, KOReader exporter rarely changes
