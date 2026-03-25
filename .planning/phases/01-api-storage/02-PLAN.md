---
phase: 01-api-storage
plan: 02
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/highlight/interfaces.go
autonomous: true
requirements:
  - API-01
  - API-02
  - API-04

must_haves:
  truths:
    - "HighlightRepo interface defines Store and GetByDocumentID methods"
    - "Highlight interface defines Sync and Fetch methods"
    - "mockgen directive present for automatic mock generation"
  artifacts:
    - path: "internal/highlight/interfaces.go"
      provides: "Interface definitions for highlight package"
      min_lines: 25
      contains: "type HighlightRepo interface"
  key_links:
    - from: "internal/highlight/interfaces.go"
      to: "internal/entity/highlight.go"
      via: "import entity package"
      pattern: "entity.Highlight"
---

<objective>
Create the interface definitions for the highlight package, establishing contracts for repository and use case implementations.

Purpose: Define clean architecture boundaries enabling testability via mocks.
Output: Interface definitions with mockgen directive for automatic mock generation.
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

From internal/sync/interfaces.go (pattern to follow):
```go
package sync

import (
    "context"
    "github.com/vanadium23/kompanion/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=sync_test

type ProgressRepo interface {
    Store(ctx context.Context, t entity.Progress) error
    GetBookHistory(ctx context.Context, bookID string, limit int) ([]entity.Progress, error)
}

// Progress -.
type Progress interface {
    Sync(context.Context, entity.Progress) (entity.Progress, error)
    Fetch(ctx context.Context, bookID string) (entity.Progress, error)
}
```

</interfaces>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Create highlight interfaces</name>
  <files>internal/highlight/interfaces.go</files>
  <read_first>
    - internal/sync/interfaces.go (interface pattern to follow)
    - internal/entity/highlight.go (entity to import - created in plan 01)
    - .planning/phases/01-api-storage/01-RESEARCH.md (research with interface definitions)
  </read_first>
  <behavior>
    - Test 1: HighlightRepo interface exists with Store method
    - Test 2: HighlightRepo interface has GetByDocumentID method
    - Test 3: Highlight interface exists with Sync method accepting array of highlights
    - Test 4: Highlight interface has Fetch method returning array of highlights
    - Test 5: Sync method returns (syncedCount int, error)
    - Test 6: Fetch method returns ([]entity.Highlight, error)
  </behavior>
  <action>
Create `internal/highlight/interfaces.go` with interface definitions matching the following specification:

```go
package highlight

import (
    "context"

    "github.com/vanadium23/kompanion/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=highlight_test

// HighlightRepo defines the repository interface for highlight persistence.
type HighlightRepo interface {
    Store(ctx context.Context, h entity.Highlight) error
    GetByDocumentID(ctx context.Context, documentID string) ([]entity.Highlight, error)
}

// Highlight defines the use case interface for highlight synchronization.
type Highlight interface {
    Sync(ctx context.Context, documentID string, highlights []entity.Highlight, deviceName string) (int, error)
    Fetch(ctx context.Context, documentID string) ([]entity.Highlight, error)
}
```

Key points:
- Sync takes array of highlights (API-02: batch support)
- Sync returns count of synced highlights (API-04: return synced count)
- GetByDocumentID fetches all highlights for a book
- Include mockgen directive for automatic mock generation
- Follow exact pattern from internal/sync/interfaces.go
  </action>
  <verify>
    <automated>go build ./internal/highlight/...</automated>
  </verify>
  <acceptance_criteria>
    - File internal/highlight/interfaces.go exists
    - File contains "package highlight"
    - File contains `//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=highlight_test`
    - File contains "type HighlightRepo interface"
    - File contains "Store(ctx context.Context, h entity.Highlight) error"
    - File contains "GetByDocumentID(ctx context.Context, documentID string) ([]entity.Highlight, error)"
    - File contains "type Highlight interface"
    - File contains "Sync(ctx context.Context, documentID string, highlights []entity.Highlight, deviceName string) (int, error)"
    - File contains "Fetch(ctx context.Context, documentID string) ([]entity.Highlight, error)"
    - Command `go build ./internal/highlight/...` exits 0
  </acceptance_criteria>
  <done>Interface definitions created with mockgen directive for API-01, API-02, API-04</done>
</task>

</tasks>

<verification>
After completing:
1. Interfaces compile: `go build ./internal/highlight/...`
2. mockgen directive present for mock generation
3. Interface signatures match RESEARCH.md specification
</verification>

<success_criteria>
- HighlightRepo interface defines Store and GetByDocumentID methods
- Highlight interface defines Sync (with array input, count output) and Fetch methods
- mockgen directive present for automatic mock generation
</success_criteria>

<output>
After completion, create `.planning/phases/01-api-storage/01-02-SUMMARY.md`
</output>
