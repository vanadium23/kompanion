# Phase 2: Web UI - Research

**Researched:** 2026-03-21
**Domain:** Go web templates (goview), Gin HTTP handlers, HTML rendering
**Confidence:** HIGH

## Summary

This phase adds highlight display functionality to the existing book detail page. The infrastructure from Phase 1 is complete: `entity.Highlight` struct exists, `highlight.Highlight` use case with `Fetch()` method is implemented, PostgreSQL repository `GetByDocumentID()` works, and the book detail page (`web/templates/book.html`) already renders reading stats. The implementation requires minimal changes: inject the highlight use case into the books routes handler, fetch highlights by `book.DocumentID`, and add a new section to the existing book template.

**Primary recommendation:** Extend the existing `booksRoutes` struct and `viewBook` handler to include highlights, following the same pattern used for reading stats.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| foolin/goview | v0.3.0 | Template rendering | Already configured in web/router.go |
| gin-gonic/gin | v1.7.7 | HTTP routing | Existing web framework |
| stretchr/testify | v1.11.1 | Testing assertions | Project standard for all tests |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| golang/mock | v1.6.0 | Mock generation | Already used in highlight package |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| goview templates | React/Vue SPA | SPA is overkill for read-only display; goview matches existing pattern |
| Separate highlights page | Embed on book page | Separate page adds navigation complexity; book page is natural location |

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── controller/http/web/
│   ├── books.go          # Extend viewBook handler
│   └── router.go         # Pass highlight use case
├── highlight/
│   ├── interfaces.go     # Already has Fetch method
│   └── sync.go           # Already implemented
├── entity/
│   └── highlight.go      # Already exists
web/
├── templates/
│   └── book.html         # Add highlights section
└── static/
    └── static.css        # Add highlight styling
```

### Pattern 1: Handler Extension Pattern
**What:** Extend existing route handler to include additional data from another use case
**When to use:** When adding display-only data to an existing page
**Example:**
```go
// Source: internal/controller/http/web/books.go (existing pattern for stats)
func (r *booksRoutes) viewBook(c *gin.Context) {
    bookID := c.Param("bookID")

    book, err := r.shelf.ViewBook(c.Request.Context(), bookID)
    // ... error handling ...

    // Existing pattern: fetch stats
    bookStats, err := r.stats.GetBookStats(c.Request.Context(), book.DocumentID)
    if err != nil {
        r.logger.Error(err, "failed to get book stats")
        bookStats = &stats.BookStats{} // Use empty on error
    }

    // NEW: Fetch highlights using same pattern
    highlights, err := r.highlight.Fetch(c.Request.Context(), book.DocumentID)
    if err != nil {
        r.logger.Error(err, "failed to fetch highlights")
        highlights = []entity.Highlight{} // Use empty on error
    }

    c.HTML(200, "book", passStandartContext(c, gin.H{
        "book":       book,
        "stats":      bookStats,
        "highlights": highlights,  // NEW
    }))
}
```

### Pattern 2: Template Section Pattern
**What:** Add a self-contained section to an existing template using goview blocks
**When to use:** When adding a new feature display to an existing page
**Example:**
```html
<!-- Source: web/templates/book.html (existing pattern for stats section) -->
{{ with $.highlights }}
<section class="highlights-section">
    <hgroup>
        <h3>Highlights</h3>
        <p>{{ len . }} highlights synced from KOReader</p>
    </hgroup>
    {{ range . }}
    <blockquote class="highlight-item">
        <p>{{ .Text }}</p>
        {{ with .Note }}<footer><em>Note: {{ . }}</em></footer>{{ end }}
        <cite>
            {{ with .Page }}Page {{ . }}{{ end }}
            {{ with .Chapter }} - {{ . }}{{ end }}
        </cite>
    </blockquote>
    {{ end }}
</section>
{{ end }}
```

### Anti-Patterns to Avoid
- **Adding highlight dependency to web.NewRouter without updating app.go:** The highlight use case is already instantiated in app.go; just pass it through to web.NewRouter
- **Creating a separate highlights route handler:** Unnecessary complexity; embed on existing book page
- **Using JavaScript for dynamic loading:** Keep it simple with server-side rendering like the rest of the app
- **Adding edit/delete functionality:** Requirements explicitly state read-only display (UI-05)

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Highlight display | Custom highlight component | goview template with range | Consistent with existing book.html pattern |
| Highlight retrieval | New repository method | highlight.Fetch() | Already implemented and tested in Phase 1 |
| Date formatting | Custom time formatting | template FuncMap if needed | Existing pattern in router.go |

**Key insight:** All infrastructure exists from Phase 1. This phase is purely about wiring and display.

## Common Pitfalls

### Pitfall 1: Wrong DocumentID Field
**What goes wrong:** Using `book.ID` instead of `book.DocumentID` to fetch highlights
**Why it happens:** Book has both ID (UUID for internal use) and DocumentID (MD5 hash for KOReader sync)
**How to avoid:** Always use `book.DocumentID` which matches `koreader_partial_md5` in highlights table
**Warning signs:** No highlights displayed even though sync worked

### Pitfall 2: Nil vs Empty Slice
**What goes wrong:** Template range fails or shows unexpected behavior with nil slice
**Why it happens:** Go templates handle nil and empty slices differently
**How to avoid:** Initialize to `[]entity.Highlight{}` on error, not nil
**Warning signs:** Template rendering errors or nothing rendered

### Pitfall 3: Missing CSS Styling
**What goes wrong:** Highlights appear unstyled or break page layout
**Why it happens:** New content needs corresponding CSS
**How to avoid:** Add highlight-specific CSS classes to static.css before testing
**Warning signs:** Content visible but poorly formatted

### Pitfall 4: Forgetting to Update Function Signature
**What goes wrong:** Build fails after adding highlight parameter
**Why it happens:** newBooksRoutes call in router.go needs updated signature
**How to avoid:** Update both the function signature AND the call site in router.go
**Warning signs:** Compile error about wrong number of arguments

## Code Examples

### Extending booksRoutes Struct
```go
// Source: internal/controller/http/web/books.go pattern
type booksRoutes struct {
    shelf     library.Shelf
    stats     stats.ReadingStats
    progress  syncpkg.Progress
    highlight highlight.Highlight  // NEW
    logger    logger.Interface
}

func newBooksRoutes(
    handler *gin.RouterGroup,
    shelf library.Shelf,
    stats stats.ReadingStats,
    progress syncpkg.Progress,
    h highlight.Highlight,  // NEW parameter
    l logger.Interface,
) {
    r := &booksRoutes{
        shelf:     shelf,
        stats:     stats,
        progress:  progress,
        highlight: h,  // NEW
        logger:    l,
    }
    // ... route registration ...
}
```

### Updating Router Call
```go
// Source: internal/controller/http/web/router.go
func NewRouter(
    handler *gin.Engine,
    l logger.Interface,
    a auth.AuthInterface,
    p sync.Progress,
    shelf library.Shelf,
    stats stats.ReadingStats,
    h highlight.Highlight,  // NEW parameter
    version string,
) {
    // ... existing setup ...

    bookGroup := handler.Group("/books")
    bookGroup.Use(authMiddleware(a))
    newBooksRoutes(bookGroup, shelf, stats, p, h, l)  // Add h parameter
}
```

### Updating app.go
```go
// Source: internal/app/app.go
web.NewRouter(handler, l, authService, progress, shelf, rs, highlightSync, cfg.Version)
```

### Highlight Template Section
```html
<!-- Add after stats section in web/templates/book.html -->
{{ with $.highlights }}
<section class="highlights-section">
    <hgroup>
        <h3>Highlights</h3>
        <p>{{ len . }} notes from your reading</p>
    </hgroup>
    {{ range . }}
    <article class="highlight-card">
        <blockquote class="highlight-text">
            {{ .Text }}
        </blockquote>
        {{ with .Note }}
        <div class="highlight-note">
            <strong>Note:</strong> {{ . }}
        </div>
        {{ end }}
        <footer class="highlight-meta">
            {{ with .Page }}<span>Page {{ . }}</span>{{ end }}
            {{ with .Chapter }}<span> / {{ . }}</span>{{ end }}
        </footer>
    </article>
    {{ end }}
</section>
{{ else }}
<section class="highlights-section">
    <hgroup>
        <h3>Highlights</h3>
        <p>No highlights synced yet. Use KOReader to highlight text in this book.</p>
    </hgroup>
</section>
{{ end }}
```

### CSS for Highlights
```css
/* Add to web/static/static.css */
.highlights-section {
    margin-top: 2rem;
    padding: 1rem;
}

.highlight-card {
    border-left: 4px solid var(--text-color);
    padding-left: 1rem;
    margin-bottom: 1rem;
}

.highlight-text {
    font-style: italic;
    margin: 0;
}

.highlight-note {
    margin-top: 0.5rem;
    color: var(--text-color-alt);
}

.highlight-meta {
    font-size: 0.875rem;
    color: var(--text-color-alt);
    margin-top: 0.25rem;
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| N/A | goview with embedded templates | Project inception | All templates use goview with master layout |

**Deprecated/outdated:**
- None relevant to this phase

## Open Questions

1. **Highlight ordering - chronological vs page order?**
   - What we know: Repository returns highlights ordered by `highlight_time ASC`
   - What's unclear: Requirements say "chronologically or by page" (UI-04)
   - Recommendation: Use existing chronological order; page order would require repository change

2. **Empty state message wording?**
   - What we know: No highlights may exist for a book
   - What's unclear: Exact wording for empty state
   - Recommendation: Use helpful message suggesting user sync from KOReader

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | stretchr/testify v1.11.1 |
| Config file | none - standard Go testing |
| Quick run command | `go test -v ./internal/controller/http/...` |
| Full suite command | `go test -v -cover ./internal/...` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| UI-01 | Highlights displayed on book detail page | integration | `go test -v ./integration-test/... -run TestHTTPKompanionShelf` | Needs extension |
| UI-02 | Highlights shown with text, page, chapter | integration | Same as above | Needs extension |
| UI-03 | User notes displayed alongside highlight text | integration | Same as above | Needs extension |
| UI-04 | Highlights ordered chronologically or by page | unit | `go test -v ./internal/highlight/... -run TestHighlightSync_Fetch` | Existing test |
| UI-05 | Read-only display (no editing in web UI) | manual | N/A - verify no edit controls in template | Manual verification |

### Sampling Rate
- **Per task commit:** `go test -v ./internal/highlight/... ./internal/controller/http/web/...`
- **Per wave merge:** `go test -v -cover ./internal/...`
- **Phase gate:** Full suite + integration test green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `integration-test/integration_test.go` - Add TestHTTPKompanionHighlights test case (sync highlights, view book, verify display)
- [ ] No new mock files needed - existing mocks_test.go covers highlight.HighlightRepo

## Sources

### Primary (HIGH confidence)
- Project source code analysis - books.go, router.go, app.go patterns
- Existing highlight package - interfaces.go, sync.go, highlight_postgres.go
- Existing templates - book.html, stats.html patterns

### Secondary (MEDIUM confidence)
- goview documentation - https://github.com/foolin/goview
- Gin template rendering patterns - project conventions

### Tertiary (LOW confidence)
- N/A - all research based on direct code analysis

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All infrastructure exists, same libraries as rest of project
- Architecture: HIGH - Clear pattern from existing stats implementation
- Pitfalls: HIGH - Based on direct code analysis of existing patterns

**Research date:** 2026-03-21
**Valid until:** 30 days - stable codebase with established patterns

---

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| UI-01 | Highlights displayed on book detail page | Extend viewBook handler, add highlights section to book.html template |
| UI-02 | Highlights shown with text, page, chapter (when available) | entity.Highlight has Text, Page, Chapter fields; use template with range |
| UI-03 | User notes displayed alongside highlight text | entity.Highlight.Note field exists; display conditionally with {{ with .Note }} |
| UI-04 | Highlights ordered chronologically or by page | Repository GetByDocumentID already orders by highlight_time ASC |
| UI-05 | Read-only display (no editing in web UI) | No form elements in template section; display-only blockquote/cite structure |
