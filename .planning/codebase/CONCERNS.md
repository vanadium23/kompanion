# Codebase Concerns

**Analysis Date:** 2026-03-21

## Tech Debt

**Session Key Type:**
- Issue: Session key is plain string, should be separate type for type safety
- Files: `internal/auth/interface.go:19`
- Impact: Harder to distinguish session keys from other strings, potential for accidental misuse
- Fix approach: Create `SessionKey` type alias as noted in TODO comment

**Error Response Handling:**
- Issue: Error responses mix JSON and HTML templates inconsistently
- Files: `internal/controller/http/web/books.go:155-168`
- Impact: API consumers may receive different content types for errors, harder to handle consistently
- Fix approach: Standardize on JSON errors for API endpoints, HTML error pages for web routes

**OPDS Search Not Implemented:**
- Issue: OPDS search endpoint marked TODO but not implemented
- Files: `internal/controller/http/opds/router.go:34`
- Impact: Users cannot search books via OPDS protocol
- Fix approach: Implement OPDS search spec or remove TODO comment

**File Extension Detection:**
- Issue: File extensions handled by string literals instead of enum
- Files: `pkg/metadata/metadata.go:53`
- Impact: Magic strings scattered, no compile-time safety for supported formats
- Fix approach: Create `FileFormat` enum type for pdf/epub/fb2

**Metadata Extraction Switch:**
- Issue: Switch statement for metadata extraction does not handle unknown formats gracefully (returns empty)
- Files: `pkg/metadata/metadata.go:31-48`
- Impact: Unknown file formats return empty Metadata without error indication
- Fix approach: Return explicit error for unsupported formats

## Known Bugs

**No known open bugs identified in code comments.**
- Note: Recent commit `29fdd41` fixed NULL values in book database queries

## Security Considerations

**MD5 for Device Password Hashing:**
- Risk: MD5 is cryptographically broken, used for KOReader compatibility
- Files: `internal/auth/auth.go:115-118`
- Current mitigation: MD5 only used for KOReader sync protocol compatibility, bcrypt used for web auth
- Recommendations: Document that MD5 is only for legacy KOReader protocol, consider migration path

**Plaintext Password in Config:**
- Risk: Password read from environment variable without validation
- Files: `config/config.go:97-101`
- Current mitigation: Environment variables are standard practice
- Recommendations: Add password complexity validation, consider hashing before storage

**Session Key in Cookie:**
- Risk: Session key stored client-side
- Files: `internal/auth/auth.go:56-57`
- Current mitigation: UUIDv7 provides unpredictable session keys
- Recommendations: Add secure/httponly cookie flags, consider CSRF protection

**SQL Query Construction:**
- Risk: Dynamic SQL in List query uses fmt.Sprintf
- Files: `internal/library/book_postgres.go:103-109`
- Current mitigation: sortBy and sortOrder are validated against whitelist (lines 82-92)
- Recommendations: Safe as-is due to whitelist validation, could add parameterized ORDER BY

## Performance Bottlenecks

**N+1 Query in Book List:**
- Problem: Each book requires separate progress fetch
- Files: `internal/controller/http/web/books.go:49-65`
- Cause: Loop over books fetching progress individually
- Improvement path: Batch fetch all progress in single query using IN clause

**Pagination Uses OFFSET:**
- Problem: OFFSET pagination degrades on large datasets
- Files: `internal/library/book_postgres.go:101-109`
- Cause: Comment acknowledges this ("yes, it's not the best way")
- Improvement path: Implement cursor-based pagination for large libraries

**Stats Processing Wait:**
- Problem: Integration test uses 2-second sleep waiting for stats
- Files: `integration-test/integration_test.go:379-380`
- Cause: No notification mechanism for async processing
- Improvement path: Add webhook or polling mechanism for processing completion

## Fragile Areas

**Book Database Scanning:**
- Files: `internal/library/book_postgres.go:119-153`, `internal/library/book_postgres.go:168-200`, `internal/library/book_postgres.go:223-247`
- Why fragile: Extensive NULL handling with sql.NullString/sql.NullInt64 duplicated across three methods
- Safe modification: Extract scanning to shared helper function
- Test coverage: Unit tests exist with mocks but no NULL edge case tests for all methods

**Stats Syncer:**
- Files: `internal/stats/syncer.go`
- Why fragile: Complex nullableToInterface helper, string sanitization, multiple database operations
- Safe modification: Add integration tests for sync scenarios
- Test coverage: Only syncer_test.go exists, may not cover all NULL combinations

**Metadata Extraction:**
- Files: `pkg/metadata/epub.go`, `pkg/metadata/fb2.go`, `pkg/metadata/pdf.go`
- Why fragile: External file parsing, charset handling (FB2), XML parsing
- Safe modification: Add more format-specific tests
- Test coverage: metadata_test.go, series_test.go exist

## Scaling Limits

**Connection Pool Size:**
- Current capacity: Default 2 connections (config/config.go:142)
- Limit: May bottleneck under concurrent load
- Scaling path: Increase PG_POOL_MAX based on expected concurrent users

**Book List Page Size:**
- Current capacity: Hardcoded 12 for web (books.go:36), 10 for OPDS (opds.go:63)
- Limit: No user-configurable page size
- Scaling path: Add query parameter for page size with max limit

**Temp File Handling:**
- Current capacity: Temp files created during upload, removed after
- Limit: May fill disk under heavy upload load
- Scaling path: Add temp file cleanup on startup, monitor disk usage

## Dependencies at Risk

**go-sqlite3 (C Library):**
- Risk: CGO dependency for SQLite in stats syncer
- Impact: Cross-compilation complexity, deployment size
- Migration plan: Stats syncer is only SQLite consumer, consider pure-Go alternative (modernc.org/sqlite)

**gin-gonic v1.7.7:**
- Risk: Older version (current is 1.9+)
- Impact: Missing security fixes and features
- Migration plan: Test upgrade path, check breaking changes in middleware

## Missing Critical Features

**OPDS Search:**
- Problem: Search not implemented in OPDS feed
- Blocks: Full KOReader integration, catalog search

**User Registration:**
- Problem: Only single user supported via config
- Blocks: Multi-user deployments

**Rate Limiting:**
- Problem: No rate limiting on API endpoints
- Blocks: Protection against brute force attacks

## Test Coverage Gaps

**Integration Test Hardcoded Credentials:**
- What's not tested: Tests use hardcoded "user"/"password" from `integration-test/integration_test.go:465`
- Files: `integration-test/integration_test.go:464-466`
- Risk: Tests may not work with different configurations
- Priority: Medium

**HTTP Handler Error Paths:**
- What's not tested: Error handling in web handlers when shelf/stats fail
- Files: `internal/controller/http/web/*.go`
- Risk: Error responses untested, may expose internals
- Priority: High

**File Upload Edge Cases:**
- What's not tested: Large files, corrupted files, unsupported formats in upload
- Files: `internal/controller/http/web/books.go:83-111`
- Risk: May crash or leak resources on malformed uploads
- Priority: High

**Concurrent Access:**
- What's not tested: Multiple devices syncing simultaneously
- Files: `internal/sync/`, `internal/stats/syncer.go`
- Risk: Race conditions in progress sync
- Priority: Medium

**NULL Value Handling:**
- What's not tested: All NULL combinations in book database operations
- Files: `internal/library/book_postgres.go`
- Risk: Crash on unexpected NULL combinations
- Priority: Medium

---

*Concerns audit: 2026-03-21*
