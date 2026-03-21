# Codebase Structure

**Analysis Date:** 2026-03-21

## Directory Layout

```
/home/deploy/kompanion/
├── cmd/                    # Application entry points
│   └── app/                # Main application
│       └── main.go         # Binary entry point
├── config/                 # Configuration loading
│   └── config.go           # Environment-based config
├── internal/               # Private application code
│   ├── app/                # Application wiring and lifecycle
│   ├── auth/               # Authentication service
│   ├── controller/         # HTTP handlers
│   │   └── http/           # HTTP-specific controllers
│   │       ├── opds/       # OPDS catalog endpoints
│   │       ├── v1/         # JSON API v1 endpoints
│   │       ├── web/        # Web UI endpoints
│   │       └── webdav/     # WebDAV endpoints
│   ├── entity/             # Domain models
│   ├── library/            # Book library service
│   ├── stats/              # Reading statistics service
│   ├── storage/            # File storage abstraction
│   └── sync/               # Progress sync service
├── pkg/                    # Reusable packages (could be external)
│   ├── httpserver/         # HTTP server wrapper
│   ├── logger/             # Logging abstraction
│   ├── metadata/           # Book metadata extraction
│   ├── postgres/           # PostgreSQL connection
│   └── utils/              # Utility functions
├── web/                    # Web assets (embedded)
│   ├── static/             # CSS files
│   └── templates/          # HTML templates
├── migrations/             # Database migrations
├── test/                   # Test data fixtures
│   └── test_data/          # Sample files for tests
├── integration-test/       # Integration test suite
├── docs/                   # Documentation
│   └── adr/                # Architecture Decision Records
├── .github/                # GitHub workflows
│   └── workflows/          # CI/CD pipelines
├── assets.go               # Embedded assets (migrations, web)
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── Makefile                # Build automation
├── Dockerfile              # Container build
├── docker-compose.yml      # Local development stack
└── docker-compose-integration.yml  # Integration test stack
```

## Directory Purposes

**cmd/app:**
- Purpose: Application entry point
- Contains: `main.go` with minimal initialization logic
- Key files: `main.go` - calls config loading and `app.Run()`

**config:**
- Purpose: Environment variable configuration
- Contains: Config struct definitions and loaders
- Key files: `config.go` - all config structs with `KOMPANION_` prefix env vars

**internal/app:**
- Purpose: Dependency injection and application lifecycle
- Contains: Wire-up of all services, repos, and HTTP server
- Key files: `app.go` - main `Run()` function, `migrate.go` - DB migrations

**internal/entity:**
- Purpose: Core domain models
- Contains: Plain Go structs with domain methods
- Key files: `book.go`, `progress.go`

**internal/auth:**
- Purpose: User and device authentication
- Contains: AuthService, UserRepo, session management
- Key files: `auth.go` - service, `interface.go` - contracts, `repo_*.go` - storage

**internal/library:**
- Purpose: Book library management
- Contains: BookShelf service, pagination, BookRepo
- Key files: `shelf.go` - service, `interfaces.go` - contracts, `book_postgres.go` - repo

**internal/sync:**
- Purpose: Reading progress synchronization
- Contains: ProgressSyncUseCase, ProgressRepo
- Key files: `progress.go` - service, `interfaces.go` - contracts

**internal/stats:**
- Purpose: KOReader reading statistics
- Contains: ReadingStats service for importing/querying stats
- Key files: `interface.go` - contracts, `stats.go` - implementation

**internal/storage:**
- Purpose: File storage abstraction
- Contains: Memory, filesystem, and PostgreSQL storage implementations
- Key files: `storage.go` - factory, `interfaces.go` - Storage interface

**internal/controller/http:**
- Purpose: HTTP routing and request handling
- Contains: Route groups for different interfaces
- Subdirectories:
  - `v1/` - JSON API (healthcheck, metrics, sync)
  - `web/` - HTML web UI (books, auth, stats, devices)
  - `opds/` - OPDS XML feed for e-readers
  - `webdav/` - WebDAV for stats upload

**pkg:**
- Purpose: Reusable infrastructure packages
- Contains: Could theoretically be imported by other projects
- Key packages:
  - `httpserver/` - Graceful shutdown wrapper
  - `logger/` - Zerolog wrapper with interface
  - `postgres/` - Connection pool setup
  - `metadata/` - EPUB/PDF/FB2 metadata extraction
  - `utils/` - MD5 hash, KOReader utilities

**web:**
- Purpose: Embedded web assets
- Contains: Static CSS and HTML templates
- Key files: `templates/*.html` - Go templates, `static/*.css`

**migrations:**
- Purpose: Database schema versioning
- Contains: SQL up/down migration files
- Naming: `YYYYMMDDHHMMSS_description.up.sql`

## Key File Locations

**Entry Points:**
- `cmd/app/main.go`: Application binary entry point

**Configuration:**
- `config/config.go`: All configuration structs and env var loading
- `.env.example`: Example environment variables

**Core Logic:**
- `internal/app/app.go`: Dependency injection and server startup
- `internal/entity/book.go`: Book domain model
- `internal/entity/progress.go`: Progress domain model

**Business Services:**
- `internal/library/shelf.go`: Book management service
- `internal/sync/progress.go`: Progress sync service
- `internal/auth/auth.go`: Authentication service
- `internal/stats/stats.go`: Statistics service

**HTTP Controllers:**
- `internal/controller/http/v1/router.go`: API v1 routing
- `internal/controller/http/web/router.go`: Web UI routing
- `internal/controller/http/opds/router.go`: OPDS routing
- `internal/controller/http/webdav/router.go`: WebDAV routing

**Database:**
- `migrations/*.up.sql`: Schema migrations
- `pkg/postgres/postgres.go`: Connection management

**Testing:**
- `internal/**/`_test.go: Co-located unit tests
- `integration-test/`: Integration test suite
- `test/test_data/`: Test fixtures (books, covers, koreader data)

## Naming Conventions

**Files:**
- Go files: `snake_case.go` (e.g., `book_postgres.go`, `interfaces.go`)
- Test files: `*_test.go` (e.g., `shelf_test.go`, `mocks_test.go`)
- Migrations: `YYYYMMDDHHMMSS_description.{up,down}.sql`

**Interfaces:**
- Single-method interfaces: Verb noun (e.g., `Storage`, `Shelf`, `Progress`)
- Multi-method interfaces: Noun describing capability (e.g., `BookRepo`, `UserRepo`, `AuthInterface`)

**Structs:**
- Services: UseCase suffix or noun (e.g., `ProgressSyncUseCase`, `BookShelf`, `AuthService`)
- Repositories: Repo suffix (e.g., `BookRepo`, `UserRepo`)

**Functions:**
- Constructors: `New` prefix (e.g., `NewBookShelf()`, `NewRouter()`)
- Factory functions: `New` prefix with options pattern (e.g., `postgres.New(url, postgres.MaxPoolSize(2))`)

**Packages:**
- Single word when possible (e.g., `auth`, `sync`, `stats`)
- Domain-driven naming matching internal context

## Where to Add New Code

**New Feature (e.g., new entity like bookmarks):**
1. Domain model: `internal/entity/bookmark.go`
2. Repository interface: `internal/bookmark/interfaces.go`
3. Repository implementation: `internal/bookmark/bookmark_postgres.go`
4. Service: `internal/bookmark/service.go`
5. HTTP routes: `internal/controller/http/v1/bookmark.go` (API) or `internal/controller/http/web/bookmark.go` (UI)
6. Wire in: `internal/app/app.go`
7. Migration: `migrations/YYYYMMDDHHMMSS_bookmark.up.sql`

**New HTTP Endpoint:**
1. Add route in appropriate router file (e.g., `internal/controller/http/v1/router.go`)
2. Create handler function in same package
3. Add service method if needed

**New Storage Backend (e.g., S3):**
1. Implement `storage.Storage` interface in `internal/storage/s3.go`
2. Add case in `storage.NewStorage()` factory in `internal/storage/storage.go`
3. Add config option in `config/config.go`

**New Book Format Support:**
1. Add extractor in `pkg/metadata/` (e.g., `pkg/metadata/mobi.go`)
2. Update `guessExtension()` in `pkg/metadata/metadata.go`
3. Add format to `entity.Book.MimeType()` in `internal/entity/book.go`

**New Configuration Option:**
1. Add field to appropriate config struct in `config/config.go`
2. Add reader function (e.g., `readNewConfig()`)
3. Update `.env.example` with documentation

## Special Directories

**.github/workflows:**
- Purpose: GitHub Actions CI/CD
- Contains: Build, test, and release workflows
- Committed: Yes

**docs/adr:**
- Purpose: Architecture Decision Records
- Contains: Design decision documentation
- Committed: Yes

**test/test_data:**
- Purpose: Test fixtures for unit and integration tests
- Contains: Sample EPUB, PDF, FB2 files; KOReader database files
- Committed: Yes

**data:**
- Purpose: Local development data storage (when using filesystem storage)
- Contains: Uploaded books and covers
- Committed: No (in .gitignore)

**.planning:**
- Purpose: GSD planning documents
- Contains: Codebase analysis, phase plans
- Committed: No (typically)

---

*Structure analysis: 2026-03-21*
