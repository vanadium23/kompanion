# Architecture

**Analysis Date:** 2026-03-21

## Pattern Overview

**Overall:** Clean Architecture / Hexagonal-style

**Key Characteristics:**
- Layered separation with interfaces defining boundaries between layers
- Dependency injection via constructors in `internal/app/app.go`
- Repository pattern for data access with multiple storage backends
- Interface-based abstractions enabling testability with mocks
- Domain entities in `internal/entity/` with no external dependencies

## Layers

**Entity Layer:**
- Purpose: Core domain models with business logic methods
- Location: `internal/entity/`
- Contains: `Book`, `Progress` structs with domain methods like `Filename()`, `MimeType()`
- Depends on: Only standard library and minimal deps (decimal for series indexing)
- Used by: All other layers

**Use Case / Service Layer:**
- Purpose: Business logic orchestration and domain operations
- Location: `internal/library/`, `internal/sync/`, `internal/stats/`, `internal/auth/`
- Contains: Service structs (`BookShelf`, `ProgressSyncUseCase`, `AuthService`) implementing domain interfaces
- Depends on: Entity layer, Repository interfaces, Storage interfaces
- Used by: Controller layer

**Repository Layer:**
- Purpose: Data persistence abstraction
- Location: `internal/library/book_postgres.go`, `internal/sync/progress_postgres.go`, `internal/auth/repo_*.go`
- Contains: PostgreSQL implementations of repository interfaces
- Depends on: `pkg/postgres`, Entity layer
- Used by: Service layer via interfaces

**Controller Layer:**
- Purpose: HTTP request handling and routing
- Location: `internal/controller/http/`
- Contains: Gin router groups and route handlers
- Depends on: Service layer interfaces, Logger, Auth
- Used by: HTTP server in `pkg/httpserver`

**Infrastructure Layer:**
- Purpose: External concerns (database, HTTP server, logging, storage)
- Location: `pkg/`
- Contains: Configurable components that can be swapped
- Depends on: External packages (gin, pgx, zerolog)
- Used by: Application layer

## Data Flow

**Book Upload Flow:**

1. HTTP POST to `/books/upload` received by `internal/controller/http/web/books.go`
2. File saved to temp location, passed to `library.BookShelf.StoreBook()`
3. `BookShelf` calculates partial MD5 hash via `pkg/utils/koreader.go`
4. Metadata extracted via `pkg/metadata/metadata.go` (supports PDF, EPUB, FB2)
5. Book file and cover stored via `storage.Storage` interface
6. Book record persisted via `library.BookRepo` to PostgreSQL
7. User redirected to book detail page

**Progress Sync Flow:**

1. KOReader sends PUT to `/syncs/progress` with `entity.Progress` JSON
2. Auth middleware extracts device name from validated credentials
3. `sync.ProgressSyncUseCase.Sync()` stores progress via `sync.ProgressRepo`
4. Timestamp returned to KOReader for confirmation

**Reading Stats Flow:**

1. KOReader sends PUT to `/webdav/statistics.sqlite3` with SQLite database
2. `stats.KOReaderPGStats.Write()` parses SQLite and upserts to PostgreSQL
3. Stats viewable via web UI at `/stats`

**State Management:**
- Stateless HTTP server using PostgreSQL for persistence
- Session-based authentication stored in `auth_session` table
- Device credentials stored as MD5 hash (KOReader compatibility requirement)

## Key Abstractions

**Storage Interface:**
- Purpose: Abstract file storage for books and covers
- Examples: `internal/storage/interfaces.go`
- Pattern: Strategy pattern with three implementations:
  - `MemoryStorage` - testing only
  - `FilesystemStorage` - local disk storage
  - `PostgresStorage` - binary data in PostgreSQL (default)
```go
type Storage interface {
    Write(ctx context.Context, source string, filepath string) error
    Read(ctx context.Context, filepath string) (*os.File, error)
}
```

**Repository Interfaces:**
- Purpose: Abstract database operations for testability
- Examples: `internal/library/interfaces.go`, `internal/sync/interfaces.go`, `internal/auth/interface.go`
- Pattern: Repository pattern with PostgreSQL implementations
```go
type BookRepo interface {
    Store(context.Context, entity.Book) error
    List(ctx context.Context, sortBy, sortOrder string, page, perPage int) ([]entity.Book, error)
    Count(ctx context.Context) (int, error)
    GetById(context.Context, string) (entity.Book, error)
    GetByFileHash(context.Context, string) (entity.Book, error)
    Update(context.Context, entity.Book) error
}
```

**Service Interfaces:**
- Purpose: Define business capability contracts
- Examples: `library.Shelf`, `sync.Progress`, `auth.AuthInterface`, `stats.ReadingStats`
- Pattern: Interface segregation - each service exposes focused methods

## Entry Points

**Application Entry:**
- Location: `cmd/app/main.go`
- Triggers: Process start
- Responsibilities: Load config, invoke `app.Run()`

**HTTP Server:**
- Location: `internal/app/app.go`
- Triggers: Called from `main.go`
- Responsibilities: Wire dependencies, start Gin server, handle graceful shutdown

**HTTP Routes:**
- `/` - Redirects to `/books`
- `/healthcheck` - Kubernetes liveness probe
- `/metrics` - Prometheus metrics
- `/auth/*` - Login/logout web UI
- `/books/*` - Book library web UI (session auth required)
- `/stats/*` - Reading statistics web UI (session auth required)
- `/devices/*` - Device management web UI (session auth required)
- `/syncs/progress` - KOReader sync API (device auth required)
- `/opds/*` - OPDS catalog for e-readers (basic auth required)
- `/webdav/*` - WebDAV for KOReader stats sync (device auth required)

## Error Handling

**Strategy:** Explicit error returns with wrapped errors

**Patterns:**
- Errors wrapped with `fmt.Errorf` including context: `fmt.Errorf("BookShelf - StoreBook - PartialMD5: %w", err)`
- Custom domain errors in entity packages: `entity.ErrBookAlreadyExists`
- Auth errors as package-level variables: `auth.ErrAuth`, `auth.UserNotFound`
- HTTP responses use JSON for API endpoints, HTML for web UI

## Cross-Cutting Concerns

**Logging:** `pkg/logger/logger.go` wraps zerolog with interface for testability
```go
type Interface interface {
    Info(message string, args ...interface{})
    Error(err error, args ...interface{})
    Fatal(err error)
}
```

**Validation:** Form binding via Gin's `ShouldBind` and `ShouldBindJSON`

**Authentication:**
- Session-based for web UI: `auth.AuthService.Login()` creates session, middleware checks `isAuthenticated`
- Basic auth for OPDS and WebDAV: Device credentials validated via `auth.CheckDevicePassword()`
- Device auth middleware for sync API: Extracts device name into context

**Middleware Chain:**
1. `gin.Logger()` - Request logging
2. `gin.Recovery()` - Panic recovery
3. Route-specific auth middleware (`authMiddleware`, `authDeviceMiddleware`, `basicAuth`)

---

*Architecture analysis: 2026-03-21*
