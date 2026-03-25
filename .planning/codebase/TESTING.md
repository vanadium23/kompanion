# Testing Patterns

**Analysis Date:** 2026-03-21

## Test Framework

**Runner:**
- Go standard testing package (`testing`)
- Testify for assertions (`github.com/stretchr/testify`)

**Assertion Libraries:**
- `github.com/stretchr/testify/require` - Fatal assertions
- `github.com/stretchr/testify/assert` - Non-fatal assertions

**Run Commands:**
```bash
make test                           # Run unit tests with coverage
go test -v -cover ./internal/...    # Direct unit test invocation
make integration-test               # Run integration tests
go test -v ./integration-test/...   # Direct integration test invocation
```

## Test File Organization

**Location:**
- Co-located with source files (same directory)
- Test files named `*_test.go`

**Naming:**
- Test files: `[source]_test.go` (e.g., `auth_test.go`, `progress_test.go`)
- Mock files: `mocks_test.go` (generated)

**Structure:**
```
internal/
├── auth/
│   ├── auth.go
│   ├── auth_test.go
│   ├── interface.go
│   ├── repo_memory.go
│   ├── repo_memory_test.go
│   └── repo_postgres.go
├── sync/
│   ├── progress.go
│   ├── progress_test.go
│   ├── mocks_test.go         # Generated mocks
│   └── interfaces.go
├── storage/
│   ├── postgres.go
│   ├── postgres_test.go
│   ├── memory_test.go
│   └── filesystem_test.go
pkg/metadata/
├── metadata.go
├── metadata_test.go
├── series_test.go
```

## Test Structure

**Suite Organization:**
```go
func TestAuthServiceUserLogin(t *testing.T) {
    ctx := context.Background()

    memory_repo := auth.NewMemoryUserRepo()
    authService := auth.InitAuthService(memory_repo, "user", "password")

    sessionKey, err := authService.Login(ctx, "user", "password", "user-agent", nil)
    if err != nil {
        t.Error("Login failed")
    }

    if !authService.IsAuthenticated(ctx, sessionKey) {
        t.Error("IsAuthenticated failed")
    }
}
```

**Table-Driven Tests:**
```go
func TestProgressFetch(t *testing.T) {
    t.Parallel()

    bookID := "bookID"
    errInternalServErr := errors.New("internal server error")

    tests := []struct {
        name string
        mock func(*MockProgressRepo)
        res  entity.Progress
        err  error
    }{
        {
            name: "empty result",
            mock: func(repo *MockProgressRepo) {
                repo.EXPECT().GetBookHistory(context.Background(), bookID, 1).Return(nil, nil)
            },
            res: entity.Progress{},
            err: nil,
        },
        // ... more cases
    }

    for _, tc := range tests {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            progressSync, repo := mockedProgress(t)
            tc.mock(repo)

            res, err := progressSync.Fetch(context.Background(), bookID)

            require.Equal(t, tc.res, res)
            require.ErrorIs(t, err, tc.err)
        })
    }
}
```

**Setup Helper Pattern:**
```go
func mockedProgress(t *testing.T) (*sync.ProgressSyncUseCase, *MockProgressRepo) {
    t.Helper()

    mockCtl := gomock.NewController(t)
    repo := NewMockProgressRepo(mockCtl)
    progress := sync.NewProgressSync(repo)

    return progress, repo
}

func setupTestBookDatabaseRepo() (pgxmock.PgxPoolIface, *library.BookDatabaseRepo) {
    mock, err := pgxmock.NewPool()
    if err != nil {
        panic(err)
    }
    pg := postgres.Mock(mock)
    bdr := library.NewBookDatabaseRepo(pg)
    return mock, bdr
}
```

## Mocking

**Framework:** gomock (`github.com/golang/mock`)

**Mock Generation:**
```bash
# Generate mocks from interface file
//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=sync_test
```

**Mock Usage Pattern:**
```go
func TestProgressFetch(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name string
        mock func(*MockProgressRepo)
        // ...
    }{
        {
            name: "first result",
            mock: func(repo *MockProgressRepo) {
                repo.EXPECT().GetBookHistory(context.Background(), bookID, 1).Return(
                    []entity.Progress{{Document: bookID}}, nil)
            },
        },
    }
    // ...
}
```

**Database Mocking with pgxmock:**
```go
func TestPostgresStorage(t *testing.T) {
    t.Run("write and read file", func(t *testing.T) {
        mock, err := pgxmock.NewPool()
        require.NoError(t, err)
        defer mock.Close()

        pg := postgres.Mock(mock)
        store := storage.NewPostgresStorage(pg)

        // Expect Write query
        mock.ExpectExec("INSERT INTO storage_blob").
            WithArgs("test.txt", pgxmock.AnyArg(), pgxmock.AnyArg()).
            WillReturnResult(pgxmock.NewResult("INSERT", 1))

        // Expect Read query
        mock.ExpectQuery("SELECT file_data FROM storage_blob").
            WithArgs("test.txt").
            WillReturnRows(mock.NewRows([]string{"file_data"}).AddRow(content))

        // Verify expectations
        err = mock.ExpectationsWereMet()
        require.NoError(t, err)
    })
}
```

**What to Mock:**
- Database connections (using pgxmock)
- Repository interfaces (using gomock)
- External dependencies

**What NOT to Mock:**
- Value objects and entities
- Pure functions
- Internal business logic (test it directly)

## Fixtures and Factories

**Test Data:**
- Located in `test/test_data/`
- Subdirectories: `books/`, `covers/`, `koreader/`

**Test Data Access Pattern:**
```go
const pathToTestDataFolder = "../../../test/test_data/books/"

func TestExtractBookMetadata(t *testing.T) {
    tests := []struct {
        name     string
        fileName string
        want     metadata.Metadata
    }{
        {
            name:     "PDF",
            fileName: "PrincessOfMars-PDF.pdf",
            want: metadata.Metadata{
                Title:  "A Princess of Mars",
                Author: "Edgar Rice Burroughs",
                Format: "pdf",
            },
        },
    }
    // ...
}
```

**Helper Functions for Test Data:**
```go
func readAll(path string) []byte {
    file, err := os.Open(path)
    if err != nil {
        return nil
    }
    defer file.Close()
    b, _ := io.ReadAll(file)
    return b
}
```

**Integration Test Fixtures:**
- `integration-test/book.epub` - Sample book for upload tests
- `test/test_data/koreader/koreader_statistics_example.sqlite3` - KOReader stats database

## Coverage

**Requirements:** No explicit coverage target enforced

**View Coverage:**
```bash
make test  # Shows coverage summary
go test -v -cover ./internal/...
```

## Test Types

**Unit Tests:**
- Test individual functions and methods in isolation
- Use mocks for external dependencies
- Located alongside source code

**Integration Tests:**
- Located in `integration-test/` directory
- Test full HTTP workflows against running server
- Use Docker Compose for test environment

**Repository Tests:**
- Test database interactions with pgxmock
- Pattern: `[Type]DatabaseRepo[Test]` (e.g., `TestBookDatabaseRepoCreate`)

## Integration Testing

**Framework:** go-hit (`github.com/Eun/go-hit`)

**Test Structure:**
```go
func TestHTTPKoreaderSyncProgress(t *testing.T) {
    // Define request body
    doc := document{
        Document:   "test",
        Percentage: 1.0,
        Progress:   "test",
        Device:     "test",
        DeviceID:   "test",
    }

    client, loginSteps := webAuthSteps()
    Test(t, Description("Login for Device"), loginSteps)
    deviceName := generateDeviceName()
    deviceSteps := setupDeviceSteps(client, deviceName)
    Test(t, Description("Device Register"), deviceSteps)

    Test(t,
        Description("Koreader Put Document Progress"),
        Put(basePath+"/syncs/progress"),
        Send().Headers("Content-Type").Add("application/json"),
        Send().Body().JSON(doc),
        Send().Headers("x-auth-user").Add(deviceName),
        Send().Headers("x-auth-key").Add(hashSyncPassword("password")),
        Expect().Status().Equal(http.StatusOK),
    )
}
```

**Helper Patterns for Integration Tests:**
```go
// Reusable auth steps
func webAuthSteps() (*http.Client, hit.IStep) {
    username, password := grabTestUser()
    client := &http.Client{
        Jar: jar,
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
    }
    template := CombineSteps(
        HTTPClient(client),
        Post(basePath+"/auth/login"),
        // ... auth steps
    )
    return client, template
}

// Random device name generator
func generateDeviceName() string {
    return petname.Generate(2, "-")
}
```

**Integration Test Setup:**
```go
func TestMain(m *testing.M) {
    err := healthCheck(attempts)
    if err != nil {
        log.Fatalf("Integration tests: host %s is not available: %s", host, err)
    }
    code := m.Run()
    os.Exit(code)
}
```

**Docker Compose for Integration Tests:**
```bash
make compose-up-integration-test
```

## Common Patterns

**Parallel Testing:**
```go
func TestProgressFetch(t *testing.T) {
    t.Parallel()

    for _, tc := range tests {
        tc := tc  // Capture range variable
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            // test code
        })
    }
}
```

**Error Testing:**
```go
// Using testify
require.ErrorIs(t, err, tc.err)
require.NoError(t, err)
assert.Error(t, err)

// Standard library
if err == nil {
    t.Error("Expected error, got nil")
}
```

**Context Usage:**
```go
ctx := context.Background()
// Pass context to all operations that need it
result, err := repo.GetById(ctx, bookID)
```

**Cleanup Pattern:**
```go
func TestSyncer(t *testing.T) {
    // Create temp file
    fp, err := os.CreateTemp("", "")
    require.NoError(t, err)

    // Copy test data to temp file
    src, err := os.Open(koreaderTestSQlite)
    io.Copy(fp, src)
    src.Close()

    // Use pgxmock
    pgmock, err := pgxmock.NewPool()
    require.NoError(t, err)
    defer pgmock.Close()

    // ... test code
}
```

---

*Testing analysis: 2026-03-21*
