# External Integrations

**Analysis Date:** 2026-03-21

## APIs & External Services

**KOReader Ecosystem:**
- KOReader Sync Server API - Reading progress synchronization
  - Endpoint: `PUT /syncs/progress` - Sync reading progress
  - Endpoint: `GET /syncs/progress/{document}` - Fetch reading progress
  - Auth: `X-Auth-User` and `X-Auth-Key` headers (MD5 hashed password)
  - Implementation: `internal/controller/http/v1/sync.go`

- KOReader Statistics via WebDAV - Reading statistics upload
  - Endpoint: `PUT /webdav/statistics.sqlite3` - Upload SQLite stats file
  - Auth: HTTP Basic Auth with device credentials
  - Implementation: `internal/controller/http/webdav/router.go`
  - Parser: `internal/stats/syncer.go` - Syncs KOReader SQLite to PostgreSQL

- OPDS Catalog - Open Publication Distribution System feed
  - Endpoint: `GET /opds/` - Main catalog
  - Endpoint: `GET /opds/newest/` - Newest books feed
  - Endpoint: `GET /opds/book/{bookID}/download` - Book download
  - Auth: HTTP Basic Auth
  - Implementation: `internal/controller/http/opds/opds.go`

## Data Storage

**Databases:**
- PostgreSQL
  - Connection: `KOMPANION_PG_URL` environment variable
  - Client: pgx/v5 with connection pooling
  - Pool config: `pkg/postgres/postgres.go`
  - Migrations: `migrations/` directory

**Database Schema:**
- `library_book` - Book metadata (title, author, ISBN, series, etc.)
- `sync_progress` - Reading progress sync data
- `auth_user` - User accounts
- `auth_device` - Device registrations
- `auth_session` - User sessions
- `stats_book` - Reading statistics per book
- `stats_page_stat_data` - Per-page reading statistics

**File Storage:**
- Primary: PostgreSQL (binary storage in database)
- Alternative: Filesystem via `KOMPANION_BSTORAGE_PATH`
- Alternative: In-memory (for testing)
- Implementation: `internal/storage/`

**Caching:**
- None (stateless application)

## Authentication & Identity

**Auth Provider:**
- Custom implementation
  - Session-based auth for web interface (cookies)
  - HTTP Basic Auth for API endpoints
  - Device authentication with MD5-hashed passwords (KOReader compatibility)
  - Implementation: `internal/auth/auth.go`

**Password Handling:**
- Users: bcrypt with cost 14
- Devices: MD5 hash (KOReader kosync plugin compatibility)
- See: `internal/auth/auth.go` functions `hashPassword()` and `hashSyncPassword()`

**Session Management:**
- UUIDv7 session keys stored in database
- Sessions include user agent and client IP tracking
- See: `internal/auth/repo_postgres.go`

## Monitoring & Observability

**Error Tracking:**
- None (logs only)

**Logs:**
- zerolog structured JSON logging to stdout
- Levels: debug, info, warn, error, fatal
- Config: `KOMPANION_LOG_LEVEL` environment variable
- Implementation: `pkg/logger/logger.go`

**Metrics:**
- Prometheus metrics exposed at `/metrics` endpoint
- See: `internal/controller/http/v1/router.go`

**Health Checks:**
- `/healthcheck` endpoint returns HTTP 200
- Used by Kubernetes/Docker health probes

## CI/CD & Deployment

**Hosting:**
- Docker Hub: `vanadium23/kompanion`
- GitHub Container Registry: `ghcr.io/vanadium23/kompanion`
- Railway (one-click deploy template)

**CI Pipeline:**
- GitHub Actions (`.github/workflows/ci.yml`)
  - Linting: hadolint, dotenv-linter
  - Unit tests with coverage upload to Codecov
  - Integration tests via Docker Compose
  - Build and push to GHCR on push to any branch

**Release Pipeline:**
- GitHub Actions triggers:
  - `publish.yml` - On tag push, builds and pushes to Docker Hub
  - `build-release-binaries.yml` - On release, builds cross-platform binaries

**Supported Platforms:**
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## Environment Configuration

**Required env vars:**
- `KOMPANION_PG_URL` - PostgreSQL connection URL
- `KOMPANION_AUTH_USERNAME` - Admin username
- `KOMPANION_AUTH_PASSWORD` - Admin password

**Optional env vars:**
- `KOMPANION_AUTH_STORAGE` - postgres or memory
- `KOMPANION_HTTP_PORT` - default 8080
- `KOMPANION_LOG_LEVEL` - default info
- `KOMPANION_PG_POOL_MAX` - default 2
- `KOMPANION_BSTORAGE_TYPE` - default postgres
- `KOMPANION_BSTORAGE_PATH` - for filesystem storage

**Secrets location:**
- Docker Hub: `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN` (GitHub secrets)
- GHCR: Uses `GITHUB_TOKEN` (automatic)

## Webhooks & Callbacks

**Incoming:**
- None explicitly defined

**Outgoing:**
- None (application does not make external API calls)

## KOReader Integration Details

**Progress Sync Protocol:**
- Compatible with KOReader kosync plugin
- Uses `X-Auth-User` (device name) and `X-Auth-Key` (MD5 password) headers
- Document identified by MD5 hash in `document` field
- Returns timestamp, percentage, progress, device info

**Statistics Sync Protocol:**
- KOReader uploads `statistics.sqlite3` file via WebDAV PUT
- Server parses SQLite and syncs to PostgreSQL
- Device name passed via authenticated user
- Processing is async (goroutine)

**OPDS Feed Format:**
- Atom XML format with OPDS catalog profile
- Supports pagination
- Book entries include download links

---

*Integration audit: 2026-03-21*
