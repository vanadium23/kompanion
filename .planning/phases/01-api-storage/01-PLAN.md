---
phase: 01-api-storage
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/entity/highlight.go
  - migrations/20260321_highlights.up.sql
autonomous: true
requirements:
  - DATA-01
  - DATA-02
  - DATA-03
  - DATA-04
  - DATA-05
  - DATA-06
  - DATA-07
  - DATA-08
  - DATA-09
  - DATA-10

must_haves:
  truths:
    - "Highlight entity struct exists with all required fields"
    - "Database migration creates highlight_annotations table"
    - "Table has unique index on (koreader_partial_md5, highlight_hash)"
  artifacts:
    - path: "internal/entity/highlight.go"
      provides: "Highlight entity definition"
      min_lines: 20
    - path: "migrations/20260321_highlights.up.sql"
      provides: "Database schema for highlights"
      contains: "CREATE TABLE highlight_annotations"
  key_links:
    - from: "internal/entity/highlight.go"
      to: "migrations/20260321_highlights.up.sql"
      via: "field names match column names"
      pattern: "Text.*string.*json.*text"
---

<objective>
Create the Highlight entity and database migration schema for storing book highlights synced from KOReader devices.

Purpose: Establish the data model and persistence layer for highlight storage.
Output: Entity struct with all required fields, database table with indexes and constraints.
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

From internal/entity/progress.go (pattern to follow):
```go
type Progress struct {
    Document       string  `json:"document"`
    Percentage     float64 `json:"percentage"`
    Progress       string  `json:"progress"`
    Device         string  `json:"device"`
    DeviceID       string  `json:"device_id"`
    Timestamp      int64   `json:"timestamp"`
    AuthDeviceName string
}
```

From migrations/20250211190954_sync.up.sql (pattern to follow):
```sql
CREATE TABLE sync_progress (
    id BIGSERIAL PRIMARY KEY,
    koreader_partial_md5 TEXT NOT NULL,
    percentage REAL NOT NULL,
    progress TEXT,
    koreader_device TEXT NOT NULL,
    koreader_device_id TEXT NOT NULL,
    auth_device_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE sync_progress IS 'This table stores the current progress on book to sync between devices';
COMMENT ON COLUMN sync_progress.koreader_device IS 'Device name from KOReader';
COMMENT ON COLUMN sync_progress.auth_device_name IS 'Device name from KOmpanion';
```

</interfaces>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Create Highlight entity struct</name>
  <files>internal/entity/highlight.go</files>
  <read_first>
    - internal/entity/progress.go (entity pattern to follow)
    - .planning/phases/01-api-storage/01-RESEARCH.md (research with entity definition)
  </read_first>
  <behavior>
    - Test 1: Highlight struct exists with Text field (required)
    - Test 2: Highlight struct has Note field (optional)
    - Test 3: Highlight struct has Page field (required)
    - Test 4: Highlight struct has Chapter field (optional)
    - Test 5: Highlight struct has Timestamp field (int64)
    - Test 6: Highlight struct has Drawer and Color fields
    - Test 7: Highlight struct has DocumentID, Device, DeviceID, AuthDeviceName, HighlightHash, CreatedAt fields
    - Test 8: JSON tags match KOReader field names (text, note, page, chapter, time, drawer, color, device, device_id)
  </behavior>
  <action>
Create `internal/entity/highlight.go` with Highlight struct matching the following specification:

```go
package entity

import "time"

type Highlight struct {
    ID             string    `json:"id"`
    DocumentID     string    `json:"document"`          // MD5 hash (koreader_partial_md5)
    Text           string    `json:"text" binding:"required"`
    Note           string    `json:"note"`
    Page           string    `json:"page"`
    Chapter        string    `json:"chapter"`
    Timestamp      int64     `json:"time"`
    Drawer         string    `json:"drawer"`            // highlight style
    Color          string    `json:"color"`             // highlight color
    Device         string    `json:"device"`
    DeviceID       string    `json:"device_id"`
    AuthDeviceName string    `json:"-"`                 // set from middleware, not from KOReader
    HighlightHash  string    `json:"-"`                 // generated for deduplication
    CreatedAt      time.Time `json:"created_at"`
}
```

Key points:
- Text field has `binding:"required"` for Gin validation
- Timestamp uses `json:"time"` to match KOReader's field name
- AuthDeviceName and HighlightHash use `json:"-"` to exclude from JSON serialization
- Follow exact naming pattern from progress.go (Device, DeviceID, AuthDeviceName)
  </action>
  <verify>
    <automated>go build ./internal/entity/...</automated>
  </verify>
  <acceptance_criteria>
    - File internal/entity/highlight.go exists
    - File contains "type Highlight struct"
    - File contains `Text string \`json:"text" binding:"required"\``
    - File contains `Note string \`json:"note"\``
    - File contains `Page string \`json:"page"\``
    - File contains `Chapter string \`json:"chapter"\``
    - File contains `Timestamp int64 \`json:"time"\``
    - File contains `Drawer string \`json:"drawer"\``
    - File contains `Color string \`json:"color"\``
    - File contains `Device string \`json:"device"\``
    - File contains `DeviceID string \`json:"device_id"\``
    - File contains `AuthDeviceName string \`json:"-"\``
    - File contains `HighlightHash string \`json:"-"\``
    - File contains `CreatedAt time.Time \`json:"created_at"\``
    - Command `go build ./internal/entity/...` exits 0
  </acceptance_criteria>
  <done>Highlight entity struct compiles with all required fields for DATA-01 through DATA-10</done>
</task>

<task type="auto">
  <name>Task 2: Create database migration for highlight_annotations table</name>
  <files>migrations/20260321_highlights.up.sql</files>
  <read_first>
    - migrations/20250211190954_sync.up.sql (migration pattern to follow)
    - .planning/phases/01-api-storage/01-RESEARCH.md (research with schema definition)
  </read_first>
  <action>
Create `migrations/20260321_highlights.up.sql` with the following schema:

```sql
CREATE TABLE highlight_annotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    koreader_partial_md5 TEXT NOT NULL,

    -- Highlight content
    text TEXT NOT NULL,
    note TEXT,
    page TEXT NOT NULL,
    chapter TEXT,

    -- Highlight metadata
    drawer TEXT,
    color TEXT,

    -- Timestamps
    highlight_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Device info
    koreader_device TEXT NOT NULL,
    koreader_device_id TEXT,
    auth_device_name TEXT NOT NULL,

    -- Deduplication
    highlight_hash TEXT NOT NULL
);

CREATE INDEX idx_highlights_document ON highlight_annotations(koreader_partial_md5);
CREATE INDEX idx_highlights_time ON highlight_annotations(highlight_time DESC);
CREATE UNIQUE INDEX idx_highlights_unique ON highlight_annotations(koreader_partial_md5, highlight_hash);

COMMENT ON TABLE highlight_annotations IS 'Stores book highlights synced from KOReader devices';
COMMENT ON COLUMN highlight_annotations.highlight_hash IS 'MD5 hash of (text:page:timestamp) for deduplication';
COMMENT ON COLUMN highlight_annotations.koreader_partial_md5 IS 'MD5 hash of book document for matching with library';
```

Key points:
- Use UUID with gen_random_uuid() for primary key (not BIGSERIAL like progress)
- Unique index on (koreader_partial_md5, highlight_hash) prevents duplicates
- koreader_device_id is nullable (consistency with progress sync)
- Note and chapter are nullable (optional fields)
  </action>
  <verify>
    <automated>ls -la migrations/20260321_highlights.up.sql</automated>
  </verify>
  <acceptance_criteria>
    - File migrations/20260321_highlights.up.sql exists
    - File contains "CREATE TABLE highlight_annotations"
    - File contains "id UUID PRIMARY KEY DEFAULT gen_random_uuid()"
    - File contains "text TEXT NOT NULL"
    - File contains "note TEXT"
    - File contains "page TEXT NOT NULL"
    - File contains "chapter TEXT"
    - File contains "drawer TEXT"
    - File contains "color TEXT"
    - File contains "highlight_time TIMESTAMPTZ NOT NULL"
    - File contains "created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()"
    - File contains "koreader_device TEXT NOT NULL"
    - File contains "koreader_device_id TEXT"
    - File contains "auth_device_name TEXT NOT NULL"
    - File contains "highlight_hash TEXT NOT NULL"
    - File contains "CREATE UNIQUE INDEX idx_highlights_unique ON highlight_annotations(koreader_partial_md5, highlight_hash)"
    - File contains "CREATE INDEX idx_highlights_document"
    - File contains "CREATE INDEX idx_highlights_time"
  </acceptance_criteria>
  <done>Migration file created with complete schema satisfying DATA-01 through DATA-10</done>
</task>

</tasks>

<verification>
After completing all tasks:
1. Entity compiles: `go build ./internal/entity/...`
2. Migration file exists: `ls -la migrations/20260321_highlights.up.sql`
3. Schema matches RESEARCH.md specification
</verification>

<success_criteria>
- Highlight entity struct exists with all fields (DATA-01 through DATA-10)
- Database migration creates highlight_annotations table with correct schema
- Unique index on (koreader_partial_md5, highlight_hash) for deduplication (SYNC-01)
</success_criteria>

<output>
After completion, create `.planning/phases/01-api-storage/01-01-SUMMARY.md`
</output>
