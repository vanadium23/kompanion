# Technology Stack

**Analysis Date:** 2026-03-21

## Languages

**Primary:**
- Go 1.26 - Backend web application

**Secondary:**
- SQL - Database migrations and queries

## Runtime

**Environment:**
- Go 1.26 runtime

**Package Manager:**
- Go modules
- Lockfile: `go.sum` (present)

## Frameworks

**Core:**
- gin-gonic/gin v1.7.7 - HTTP web framework
- jackc/pgx/v5 v5.6.0 - PostgreSQL driver and toolkit

**Testing:**
- stretchr/testify v1.11.1 - Testing assertions and mocks
- Eun/go-hit v0.5.23 - HTTP integration testing
- pashagolub/pgxmock/v4 v4.2.0 - PostgreSQL mocking
- golang/mock v1.6.0 - Mock generation

**Build/Dev:**
- golang-migrate/migrate/v4 v4.15.1 - Database migrations
- foolin/goview v0.3.0 - Template rendering engine

## Key Dependencies

**Critical:**
- golang.org/x/crypto v0.49.0 - bcrypt password hashing
- rs/zerolog v1.26.1 - Structured logging
- prometheus/client_golang v1.23.2 - Metrics exposition
- moroz/uuidv7-go - UUIDv7 generation for session keys
- shopspring/decimal v1.4.0 - Decimal handling via pgx integration
- wcharczuk/go-chart/v2 v2.1.0 - Chart generation for statistics

**Infrastructure:**
- mattn/go-sqlite3 v1.14.19 - SQLite support (KOReader stats import)
- dustinkirkland/golang-petname v0.0.0-20240428194347-eebcea082ee0 - Random device name generation

## Configuration

**Environment:**
- All config via environment variables with `KOMPANION_` prefix
- Configuration loaded in `config/config.go`

**Key Environment Variables:**
- `KOMPANION_AUTH_USERNAME` - Admin username (required)
- `KOMPANION_AUTH_PASSWORD` - Admin password (required)
- `KOMPANION_AUTH_STORAGE` - postgres or memory (default: postgres)
- `KOMPANION_HTTP_PORT` - Server port (default: 8080)
- `KOMPANION_LOG_LEVEL` - debug, info, error (default: info)
- `KOMPANION_PG_URL` - PostgreSQL connection string (required)
- `KOMPANION_PG_POOL_MAX` - Connection pool size (default: 2)
- `KOMPANION_BSTORAGE_TYPE` - Book storage type: postgres, memory, filesystem (default: postgres)
- `KOMPANION_BSTORAGE_PATH` - File path for filesystem storage

**Build:**
- `Makefile` - Development commands
- `Dockerfile` - Multi-stage Docker build
- `build_release.sh` - Cross-platform binary release script

## Platform Requirements

**Development:**
- Go 1.26+
- Docker and Docker Compose (for PostgreSQL)
- golangci-lint (optional, for linting)
- migrate CLI (for database migrations)

**Production:**
- PostgreSQL database
- Docker container or standalone binary
- Deployable to Railway, Docker Hub, or GHCR

---

*Stack analysis: 2026-03-21*
