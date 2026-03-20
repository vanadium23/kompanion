# Search Feature Specification

## Goal & Context

Add search functionality to KOmpanion to allow users to find books by title, author, description, series, and other metadata fields.

**Problem**: Users currently can only browse books with pagination and sorting. With a growing library, finding specific books becomes difficult.

**Solution**: Implement full-text search across book metadata with integration into Web UI, API, and OPDS catalog.

## Architecture & Data Models

### Search Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Web UI    │     │  REST API   │     │    OPDS     │
│  /books?q=  │     │ /api/v1/... │     │ /opds/...   │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
                    ┌──────▼──────┐
                    │  Controller │
                    │  (Gin)      │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │   UseCase   │
                    │ SearchBooks │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ Repository  │
                    │ (PostgreSQL)│
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ PostgreSQL  │
                    │ FTS (GIN)   │
                    └─────────────┘
```

### Data Model Changes

No schema changes required. Use PostgreSQL full-text search on existing columns:
- `title` (weight A - highest)
- `author` (weight A)
- `series` (weight B)
- `publisher` (weight C)
- `description` (weight C)

### Search Request Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `q` | string | required | Search query |
| `page` | int | 1 | Page number |
| `limit` | int | 12 | Items per page (max 100) |
| `sort` | string | rank | Sort field: rank, title, author, year |
| `order` | string | desc | Sort order: asc, desc |

### Search Response

```json
{
  "books": [
    {
      "id": "uuid",
      "title": "string",
      "author": "string",
      "description": "string",
      "publisher": "string",
      "year": 2024,
      "series": "string",
      "series_index": 1.0,
      "cover_url": "/books/{id}/cover"
    }
  ],
  "total": 42,
  "page": 1,
  "limit": 12,
  "query": "tolkien"
}
```

## API Contracts

### 1. Web UI Search

```
GET /books?q={query}&page={page}&sort={field}&order={dir}
```

- Renders existing books list template with filtered results
- Adds search input to the page header
- Highlights search terms in results (optional enhancement)

### 2. REST API Search

```
GET /api/v1/books/search?q={query}&page={page}&limit={limit}&sort={field}&order={dir}
```

- Returns JSON response with books and pagination metadata
- Includes search rank/relevance score

### 3. OPDS Search

```
GET /opds/search/{searchTerms}/
```

- Returns OPDS feed with matching books
- Already referenced in OPDS catalog but not implemented

## Edge Cases & Constraints

### Edge Cases
1. **Empty query**: Return empty results with message "Enter search query"
2. **No results**: Show "No books found" message with suggestions
3. **Special characters**: Sanitize query to prevent SQL injection
4. **Unicode/ Cyrillic**: Ensure proper handling of Russian and other non-Latin text
5. **Long queries**: Limit query length to 500 characters

### Constraints
1. **Performance**: Search should complete in < 100ms for typical queries
2. **Pagination**: Maximum 100 results per page
3. **Case insensitive**: Search should be case insensitive
4. **Partial matches**: Support prefix matching (e.g., "Tolk" finds "Tolkien")

## Acceptance Criteria

- [x] User can search books via Web UI with search bar
- [x] Search works across title, author, series, publisher, description
- [x] Results are ranked by relevance (PostgreSQL ts_rank)
- [x] Search results are paginated
- [ ] API endpoint `/api/v1/books/search` returns JSON results (deferred)
- [x] OPDS endpoint `/opds/search/{searchTerms}/` returns Atom feed
- [x] Search handles Cyrillic text correctly (via PostgreSQL english dictionary)
- [x] Empty/invalid queries are handled gracefully
- [ ] Unit tests cover search logic (TODO)
- [ ] Integration tests cover API endpoints (TODO)

## Boundaries

### In Scope
- Full-text search on book metadata
- Web UI and OPDS integration
- PostgreSQL full-text search

### Out of Scope
- Search within book content (full-text of EPUBs/PDFs)
- Search history/saved searches
- Advanced filters (by year range, publisher, etc.)
- Search suggestions/autocomplete
- Fuzzy search with typo tolerance
- REST API search endpoint (can be added later if needed)

## Decision Context

### Why PostgreSQL Full-Text Search?

1. **No additional dependencies**: Works with existing PostgreSQL database
2. **Good performance**: GIN indexes provide fast full-text search
3. **Relevance ranking**: Built-in `ts_rank` for sorting by relevance
4. **Multilingual support**: Supports Russian and English with proper dictionaries

### Alternatives Considered

1. **Elasticsearch**: Overkill for a personal book library, adds operational complexity
2. **SQLite FTS**: Not applicable (using PostgreSQL)
3. **LIKE queries**: Poor performance and no relevance ranking

## Implementation Status

1. [x] Create database migration for GIN index on searchable columns
   - `migrations/20250320120000_search_index.up.sql`
2. [x] Add `SearchBooks` method to BookRepository interface
   - `internal/library/interfaces.go`
3. [x] Implement PostgreSQL full-text search in repository
   - `internal/library/book_postgres.go` - `Search()` and `SearchCount()`
4. [x] Add search handler to web controller
   - `internal/controller/http/web/books.go` - updated `listBooks()`
5. [x] Implement OPDS search endpoint
   - `internal/controller/http/opds/router.go` - `searchBooks()`
6. [x] Update web templates with search form
   - `web/templates/books.html`
7. [x] Add pagination link helper for search query preservation
   - `internal/controller/http/web/router.go` - `paginationLink` template function
8. [ ] Unit tests for search logic (TODO)
9. [ ] Integration tests for endpoints (TODO)
