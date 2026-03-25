---
phase: 01-api-storage
plan: 03
type: execute
wave: 2
depends_on:
  - 01
  - 02
files_modified:
  - internal/highlight/sync.go
  - internal/highlight/highlight_postgres.go
  - internal/highlight/mocks_test.go
  - internal/highlight/sync_test.go
  - internal/highlight/highlight_postgres_test.go
autonomous: true
requirements:
  - SYNC-01
  - SYNC-02

must_haves:
  truths:
    - "HighlightSyncUseCase.Sync stores highlights and returns synced count"
    - "HighlightDatabaseRepo.Store uses ON CONFLICT DO NOTHING for deduplication"
    - "Highlights for unknown books are stored without errors"
    - "Content hash is generated from (text:page:timestamp)"
  artifacts:
    - path: "internal/highlight/sync.go"
      provides: "HighlightSyncUseCase implementation"
      min_lines: 50
      contains: "func NewHighlightSync"
    - path: "internal/highlight/highlight_postgres.go"
      provides: "PostgreSQL repository implementation"
      min_lines: 60
      contains: "ON CONFLICT"
  key_links:
    - from: "internal/highlight/sync.go"
      to: "internal/highlight/highlight_postgres.go"
      via: "HighlightRepo interface"
      pattern: "uc.repo.Store"
    - from: "internal/highlight/sync.go"
      to: "crypto/md5"
      via: "generateHash function"
      pattern: "md5.Sum"
---

<objective>
Implement the use case layer and PostgreSQL repository for highlight synchronization with deduplication support.

Purpose: Enable highlight storage with idempotent behavior (no duplicates on re-sync) and orphan handling.
Output: HighlightSyncUseCase, HighlightDatabaseRepo, and unit tests with mocks.
</objective>

<execution_context>
@~/.claude/get-shit-done/workflows/execute-plan.md
@~/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/01-api-storage/01-RESEARCH.md
</context>

<interfaces>

From internal/sync/progress.go (pattern to follow):
```go
type ProgressSyncUseCase struct {
    repo ProgressRepo
}

func NewProgressSync(r ProgressRepo) *ProgressSyncUseCase {
    return &ProgressSyncUseCase{repo: r}
}

func (uc *ProgressSyncUseCase) Sync(ctx context.Context, doc entity.Progress) (entity.Progress, error) {
    if doc.Timestamp == 0 {
        doc.Timestamp = time.Now().Unix()
    }
    err := uc.repo.Store(ctx, doc)
    if err != nil {
        return doc, fmt.Errorf("ProgressSyncUseCase - Sync - s.repo.Sync: %w", err)
    }
    return doc, nil
}
```

From internal/sync/progress_postgres.go (pattern to follow):
```go
type ProgressDatabaseRepo struct {
    *postgres.Postgres
}

func NewProgressDatabaseRepo(pg *postgres.Postgres) *ProgressDatabaseRepo {
    return &ProgressDatabaseRepo{pg}
}

func (r *ProgressDatabaseRepo) Store(ctx context.Context, t entity.Progress) error {
    sql := `INSERT INTO sync_progress ...`
    args := []interface{}{...}
    _, err := r.Pool.Exec(ctx, sql, args...)
    if err != nil {
        return fmt.Errorf("TranslationRepo - Store - r.Pool.Exec: %w", err)
    }
    return nil
}
```

From internal/sync/progress_test.go (test pattern to follow):
```go
func mockedProgress(t *testing.T) (*sync.ProgressSyncUseCase, *MockProgressRepo) {
    t.Helper()
    mockCtl := gomock.NewController(t)
    repo := NewMockProgressRepo(mockCtl)
    progress := sync.NewProgressSync(repo)
    return progress, repo
}
```

</interfaces>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement HighlightSyncUseCase</name>
  <files>internal/highlight/sync.go, internal/highlight/sync_test.go</files>
  <read_first>
    - internal/sync/progress.go (use case pattern to follow)
    - internal/sync/progress_test.go (test pattern to follow)
    - internal/highlight/interfaces.go (interfaces from plan 02)
    - internal/entity/highlight.go (entity from plan 01)
    - .planning/phases/01-api-storage/01-RESEARCH.md (research with implementation example)
  </read_first>
  <behavior>
    - Test 1: NewHighlightSync creates use case with repository
    - Test 2: Sync returns synced count for valid highlights
    - Test 3: Sync sets DocumentID, AuthDeviceName, CreatedAt, HighlightHash on each highlight
    - Test 4: Sync continues on individual Store errors (does not fail entire batch)
    - Test 5: generateHash produces consistent MD5 hash from text:page:timestamp
    - Test 6: Fetch calls repository GetByDocumentID
  </behavior>
  <action>
Create `internal/highlight/sync.go` with HighlightSyncUseCase:

```go
package highlight

import (
    "context"
    "crypto/md5"
    "encoding/hex"
    "fmt"
    "time"

    "github.com/vanadium23/kompanion/internal/entity"
)

// HighlightSyncUseCase implements highlight synchronization logic.
type HighlightSyncUseCase struct {
    repo HighlightRepo
}

// NewHighlightSync creates a new highlight sync use case.
func NewHighlightSync(r HighlightRepo) *HighlightSyncUseCase {
    return &HighlightSyncUseCase{repo: r}
}

// Sync stores highlights and returns count of successfully synced items.
func (uc *HighlightSyncUseCase) Sync(ctx context.Context, documentID string, highlights []entity.Highlight, deviceName string) (int, error) {
    synced := 0
    for i := range highlights {
        highlights[i].DocumentID = documentID
        highlights[i].AuthDeviceName = deviceName
        highlights[i].CreatedAt = time.Now()
        highlights[i].HighlightHash = generateHash(highlights[i].Text, highlights[i].Page, highlights[i].Timestamp)

        if err := uc.repo.Store(ctx, highlights[i]); err != nil {
            // Log and continue - unique constraint violation is expected for duplicates (SYNC-01)
            continue
        }
        synced++
    }
    return synced, nil
}

// Fetch retrieves all highlights for a document.
func (uc *HighlightSyncUseCase) Fetch(ctx context.Context, documentID string) ([]entity.Highlight, error) {
    return uc.repo.GetByDocumentID(ctx, documentID)
}

// generateHash creates MD5 hash from text:page:timestamp for deduplication.
func generateHash(text, page string, timestamp int64) string {
    data := fmt.Sprintf("%s:%s:%d", text, page, timestamp)
    hash := md5.Sum([]byte(data))
    return hex.EncodeToString(hash[:])
}
```

Create `internal/highlight/sync_test.go` following pattern from internal/sync/progress_test.go with tests for:
- Sync with empty highlights returns 0
- Sync sets all required fields on highlights
- Sync continues on Store errors
- Fetch calls repository
  </action>
  <verify>
    <automated>go test -v ./internal/highlight/... -run TestHighlightSync</automated>
  </verify>
  <acceptance_criteria>
    - File internal/highlight/sync.go exists
    - File contains "type HighlightSyncUseCase struct"
    - File contains "func NewHighlightSync(r HighlightRepo)"
    - File contains "func (uc *HighlightSyncUseCase) Sync"
    - File contains "func generateHash(text, page string, timestamp int64) string"
    - File contains "md5.Sum([]byte(data))"
    - File contains "hex.EncodeToString(hash[:])"
    - Sync method signature: "(ctx context.Context, documentID string, highlights []entity.Highlight, deviceName string) (int, error)"
    - Sync sets DocumentID, AuthDeviceName, CreatedAt, HighlightHash
    - Sync continues on Store errors (does not return error for duplicates)
    - File internal/highlight/sync_test.go exists
    - Command `go test -v ./internal/highlight/... -run TestHighlightSync` exits 0
  </acceptance_criteria>
  <done>HighlightSyncUseCase implemented with deduplication hash generation satisfying SYNC-01, SYNC-02</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Implement HighlightDatabaseRepo</name>
  <files>internal/highlight/highlight_postgres.go, internal/highlight/highlight_postgres_test.go</files>
  <read_first>
    - internal/sync/progress_postgres.go (repository pattern to follow)
    - internal/sync/progress_postgres_test.go (test pattern to follow)
    - internal/highlight/interfaces.go (interfaces from plan 02)
    - internal/entity/highlight.go (entity from plan 01)
    - .planning/phases/01-api-storage/01-RESEARCH.md (research with implementation example)
  </read_first>
  <behavior>
    - Test 1: Store inserts highlight with all fields
    - Test 2: Store uses ON CONFLICT DO NOTHING for deduplication
    - Test 3: GetByDocumentID retrieves highlights ordered by highlight_time
    - Test 4: GetByDocumentID returns empty slice when no highlights found
    - Test 5: Store converts Timestamp (int64) to time.Time for highlight_time column
  </behavior>
  <action>
Create `internal/highlight/highlight_postgres.go`:

```go
package highlight

import (
    "context"
    "fmt"
    "time"

    "github.com/vanadium23/kompanion/internal/entity"
    "github.com/vanadium23/kompanion/pkg/postgres"
)

// HighlightDatabaseRepo implements HighlightRepo using PostgreSQL.
type HighlightDatabaseRepo struct {
    *postgres.Postgres
}

// NewHighlightDatabaseRepo creates a new highlight repository.
func NewHighlightDatabaseRepo(pg *postgres.Postgres) *HighlightDatabaseRepo {
    return &HighlightDatabaseRepo{pg}
}

// Store inserts a highlight with ON CONFLICT DO NOTHING for deduplication.
func (r *HighlightDatabaseRepo) Store(ctx context.Context, h entity.Highlight) error {
    sql := `INSERT INTO highlight_annotations
        (koreader_partial_md5, text, note, page, chapter, drawer, color,
         highlight_time, koreader_device, koreader_device_id, auth_device_name, highlight_hash)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        ON CONFLICT (koreader_partial_md5, highlight_hash) DO NOTHING`

    args := []interface{}{
        h.DocumentID, h.Text, h.Note, h.Page, h.Chapter, h.Drawer, h.Color,
        time.Unix(h.Timestamp, 0), h.Device, h.DeviceID, h.AuthDeviceName, h.HighlightHash,
    }

    _, err := r.Pool.Exec(ctx, sql, args...)
    if err != nil {
        return fmt.Errorf("HighlightDatabaseRepo - Store - r.Pool.Exec: %w", err)
    }
    return nil
}

// GetByDocumentID retrieves all highlights for a document, ordered by time.
func (r *HighlightDatabaseRepo) GetByDocumentID(ctx context.Context, documentID string) ([]entity.Highlight, error) {
    sql := `SELECT id, koreader_partial_md5, text, note, page, chapter, drawer, color,
            highlight_time, koreader_device, koreader_device_id, auth_device_name, created_at
            FROM highlight_annotations
            WHERE koreader_partial_md5 = $1
            ORDER BY highlight_time ASC`

    rows, err := r.Pool.Query(ctx, sql, documentID)
    if err != nil {
        return nil, fmt.Errorf("HighlightDatabaseRepo - GetByDocumentID - r.Pool.Query: %w", err)
    }
    defer rows.Close()

    var highlights []entity.Highlight
    for rows.Next() {
        var h entity.Highlight
        var highlightTime time.Time
        err = rows.Scan(&h.ID, &h.DocumentID, &h.Text, &h.Note, &h.Page, &h.Chapter,
            &h.Drawer, &h.Color, &highlightTime, &h.Device, &h.DeviceID, &h.AuthDeviceName, &h.CreatedAt)
        if err != nil {
            return nil, fmt.Errorf("HighlightDatabaseRepo - GetByDocumentID - rows.Scan: %w", err)
        }
        h.Timestamp = highlightTime.Unix()
        highlights = append(highlights, h)
    }
    return highlights, nil
}
```

Create `internal/highlight/highlight_postgres_test.go` following pattern from internal/sync/progress_postgres_test.go using pgxmock:
- Test Store with valid highlight
- Test GetByDocumentID returns highlights
- Test GetByDocumentID with no results returns empty slice
  </action>
  <verify>
    <automated>go test -v ./internal/highlight/... -run TestHighlightRepo</automated>
  </verify>
  <acceptance_criteria>
    - File internal/highlight/highlight_postgres.go exists
    - File contains "type HighlightDatabaseRepo struct"
    - File contains "func NewHighlightDatabaseRepo(pg *postgres.Postgres)"
    - File contains "ON CONFLICT (koreader_partial_md5, highlight_hash) DO NOTHING"
    - Store uses 12 parameter placeholders ($1 through $12)
    - Store converts h.Timestamp to time.Unix(h.Timestamp, 0)
    - GetByDocumentID orders by "highlight_time ASC"
    - GetByDocumentID scans all 13 columns
    - File internal/highlight/highlight_postgres_test.go exists
    - Command `go test -v ./internal/highlight/... -run TestHighlightRepo` exits 0
  </acceptance_criteria>
  <done>HighlightDatabaseRepo implemented with UPSERT deduplication satisfying SYNC-01, SYNC-02</done>
</task>

<task type="auto">
  <name>Task 3: Generate mocks for testing</name>
  <files>internal/highlight/mocks_test.go</files>
  <read_first>
    - internal/highlight/interfaces.go (interfaces with mockgen directive)
    - internal/sync/mocks_test.go (example generated mock)
  </read_first>
  <action>
Generate mocks for the highlight package by running mockgen. The interfaces.go file already has the directive:

```go
//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=highlight_test
```

Run the generator:
```bash
go generate ./internal/highlight/...
```

This creates mocks_test.go with MockHighlightRepo and MockHighlight implementations that tests can use.
  </action>
  <verify>
    <automated>go test -v ./internal/highlight/... -run TestHighlightSync</automated>
  </verify>
  <acceptance_criteria>
    - File internal/highlight/mocks_test.go exists
    - File contains "type MockHighlightRepo struct"
    - File contains "type MockHighlight struct"
    - File contains "func (m *MockHighlightRepo) Store"
    - File contains "func (m *MockHighlightRepo) GetByDocumentID"
    - File contains "func (m *MockHighlight) Sync"
    - File contains "func (m *MockHighlight) Fetch"
    - Command `go test -v ./internal/highlight/...` exits 0
  </acceptance_criteria>
  <done>Mocks generated for HighlightRepo and Highlight interfaces</done>
</task>

</tasks>

<verification>
After completing all tasks:
1. All tests pass: `go test -v ./internal/highlight/...`
2. Mocks generated: `ls internal/highlight/mocks_test.go`
3. ON CONFLICT clause present for deduplication
</verification>

<success_criteria>
- HighlightSyncUseCase.Sync stores highlights with generated content hash (SYNC-01)
- HighlightDatabaseRepo.Store uses ON CONFLICT DO NOTHING for idempotency (SYNC-01)
- Orphan highlights stored without requiring book reference (SYNC-02)
- Unit tests pass with mocked repository
</success_criteria>

<output>
After completion, create `.planning/phases/01-api-storage/01-03-SUMMARY.md`
</output>
