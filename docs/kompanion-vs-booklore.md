# KOmpanion vs BookLore: Comprehensive Comparison

> Generated: 2026-02-28

## Executive Summary

| Aspect | KOmpanion | BookLore |
|--------|-----------|----------|
| **Philosophy** | Minimalist KOReader companion | Full-featured digital library |
| **Tagline** | "KOReader-first" simplicity | "Your books deserve a home" |
| **License** | MIT (permissive) | AGPL v3.0 (copyleft) |
| **Tech Stack** | Go + PostgreSQL | Java/Spring Boot + MariaDB |
| **Deployment** | Single binary (~25MB) | Docker Compose (~500MB+) |

---

## Architecture

### KOmpanion

```
cmd/app/main.go
    │
    ▼
internal/                          pkg/
├── controller/http/               ├── postgres/
│   ├── v1/ (REST API)            ├── metadata/
│   ├── web/ (HTML UI)            ├── httpserver/
│   ├── opds/                     └── utils/
│   └── webdav/
├── entity/                        web/
├── usecase/                       ├── static/
├── library/                       └── templates/
├── sync/
├── stats/
├── auth/
└── storage/
```

**Key Design Decisions:**
- Clean Architecture with layered separation
- Single binary with `go:embed` for assets and migrations
- Repository pattern for data access
- Multiple storage backends (PostgreSQL, filesystem, memory)

### BookLore

```
Docker Container
    │
    ▼
Spring Boot Application
├── REST Controllers
├── Service Layer
├── JPA/Hibernate
├── BookDrop Watcher
└── External API Clients
    ├── Google Books
    ├── Open Library
    └── Amazon
```

**Key Design Decisions:**
- Standard Spring MVC architecture
- Docker-first deployment
- External service integrations
- Background file watching

---

## Feature Comparison Matrix

### File Format Support

| Format | KOmpanion | BookLore |
|--------|:---------:|:--------:|
| EPUB | ✅ | ✅ |
| PDF | ✅ | ✅ |
| FB2 | ✅ | ❌ |
| MOBI | ⚠️ (basic) | ❌ |
| CBZ/CBR/CB7 (comics) | ❌ | ✅ |
| AZW3 | ❌ | ❌ |

### Library Management

| Feature | KOmpanion | BookLore |
|---------|:---------:|:--------:|
| Upload books | ✅ | ✅ |
| Single file upload | ✅ | ✅ |
| Batch upload | ❌ | ✅ |
| BookDrop (auto-import) | ❌ | ✅ |
| Book deletion | ❌ | ✅ |
| Metadata editing | ⚠️ (partial) | ✅ |
| Series support | ✅ | ✅ |
| Collections/Tags | ❌ | ✅ |
| Smart/Magic Shelves | ❌ | ✅ |
| Full-text search | ❌ | ✅ |
| Sort/filter | ❌ | ✅ |

### Metadata

| Feature | KOmpanion | BookLore |
|---------|:---------:|:--------:|
| Extract from file | ✅ | ✅ |
| Title/Author | ✅ | ✅ |
| Series/SeriesIndex | ✅ | ✅ |
| Publisher/Year | ✅ | ✅ |
| ISBN | ✅ | ✅ |
| Description/Synopsis | ❌ | ✅ |
| Cover extraction | ✅ | ✅ |
| Auto-fetch (Google Books) | ❌ | ✅ |
| Auto-fetch (Open Library) | ❌ | ✅ |
| Auto-fetch (Amazon) | ❌ | ✅ |
| Ratings/Reviews | ❌ | ✅ |

### Reading Experience

| Feature | KOmpanion | BookLore |
|---------|:---------:|:--------:|
| Built-in web reader | ❌ | ✅ |
| Annotations | ❌ | ✅ |
| Highlights | ❌ | ✅ |
| Bookmarks | ❌ | ✅ |
| Reading progress | ✅ | ✅ |
| Progress bars | ✅ | ✅ |

### Device Integration

| Feature | KOmpanion | BookLore |
|---------|:---------:|:--------:|
| KOReader progress sync | ✅ | ✅ |
| KOReader MD5 passwords | ✅ | ✅ |
| WebDAV statistics | ✅ | ❌ |
| OPDS catalog | ✅ | ✅ |
| Kobo sync | ❌ | ✅ |
| Kindle sharing | ❌ | ✅ |

### User Management

| Feature | KOmpanion | BookLore |
|---------|:---------:|:--------:|
| Multi-user | ✅ | ✅ |
| Device management | ✅ | ✅ |
| Session-based auth | ✅ | ✅ |
| Basic auth | ✅ | ✅ |
| OIDC/OAuth | ❌ | ✅ |
| User permissions | ⚠️ (basic) | ✅ |

### APIs & Protocols

| Feature | KOmpanion | BookLore |
|---------|:---------:|:--------:|
| REST API | ✅ | ✅ |
| OPDS feed | ✅ | ✅ |
| WebDAV | ✅ | ❌ |
| Health check | ✅ | ✅ |
| Prometheus metrics | ✅ | ❌ |

---

## Technical Comparison

### Performance & Resources

| Metric | KOmpanion | BookLore |
|--------|-----------|----------|
| Binary/Image size | ~25MB | ~200MB+ |
| Memory usage | ~50-100MB | ~512MB+ |
| Startup time | <1s | ~10-30s |
| Database | PostgreSQL | MariaDB |
| Connection pooling | pgx | HikariCP |

### Deployment

| Aspect | KOmpanion | BookLore |
|--------|-----------|----------|
| **Prerequisites** | PostgreSQL only | Docker + Docker Compose |
| **Installation** | Download binary | `docker-compose up` |
| **Configuration** | Environment vars | .env + docker-compose.yml |
| **Updates** | Replace binary | Pull new image |
| **Portability** | Single file | Multiple containers |

### Development

| Aspect | KOmpanion | BookLore |
|--------|-----------|----------|
| Language | Go 1.22+ | Java (Spring Boot) |
| Framework | Gin | Spring MVC |
| ORM | Raw SQL/pgx | JPA/Hibernate |
| Migrations | golang-migrate | Flyway/Liquibase |
| Templates | Goview | Thymeleaf/React |
| Testing | testify | JUnit/Mockito |

---

## Design Philosophy

### KOmpanion Principles

1. **KOReader-First**: Built specifically for KOReader workflows
2. **Minimalist**: Only essential features, no bloat
3. **Single Binary**: Easy deployment, embedded assets
4. **Self-Contained**: Files can live in PostgreSQL itself
5. **No Competition**: Doesn't try to be a reader itself

> "Avoids being another Calibre-like complex solution"

### BookLore Principles

1. **All-In-One**: Complete library management experience
2. **Automation-First**: BookDrop, auto-metadata, smart shelves
3. **Social**: Sharing, ratings, reviews, multi-user
4. **Feature-Rich**: Built-in reader, annotations, highlights

> "The ebook world's Jellyfin"

---

## Code Structure Comparison

### KOmpanion: Clean Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Web HTML   │  │  REST API   │  │  OPDS / WebDAV      │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    Application Layer                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Library   │  │    Sync     │  │       Stats         │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                       Domain Layer                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Entity    │  │   UseCase   │  │    Interfaces       │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                   Infrastructure Layer                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  PostgreSQL │  │  Filesystem │  │    Metadata Pkg     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### BookLore: Spring MVC Style

```
┌─────────────────────────────────────────────────────────────┐
│                      Controllers                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  BookCtrl   │  │  UserCtrl   │  │   MetadataCtrl      │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                       Services                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  BookSvc    │  │  SyncSvc    │  │   MetadataSvc       │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    Repositories (JPA)                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  BookRepo   │  │  UserRepo   │  │    ProgressRepo     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    External Integrations                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ GoogleBooks │  │ OpenLibrary │  │     BookDrop        │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## Use Case Scenarios

### Choose KOmpanion When:

- You primarily use KOReader across multiple devices
- You want simple progress sync without complexity
- You have a low-resource VPS or prefer minimal footprint
- You value single-binary deployment simplicity
- You self-host PostgreSQL already
- You don't need a web-based reader
- You prefer MIT license for modifications

### Choose BookLore When:

- You want a complete library management system
- You need automatic metadata enrichment
- You read comics/manga (CBZ/CBR support)
- You want a built-in web reader
- You use Kobo devices (not just KOReader)
- You want smart shelves and automation
- You prefer Docker-based deployment
- OIDC/SAML authentication is required

---

## Integration Points

### KOReader Sync Protocol

Both implement the Kosync protocol:

```
POST /syncs/progress
{
  "document": "md5hash",
  "progress": "0.5",
  "percentage": 50,
  "device": "device-id",
  "device_id": "device-id"
}

GET /syncs/progress/{document}
→ Returns last sync state
```

### OPDS Catalog

Both provide OPDS feeds for discovery:

```xml
<feed xmlns="http://www.w3.org/2005/Atom">
  <entry>
    <title>Book Title</title>
    <author><name>Author Name</name></author>
    <link href="/books/{id}/download" type="application/epub+zip"/>
  </entry>
</feed>
```

### Unique to KOmpanion: WebDAV Statistics

KOmpanion accepts reading statistics via WebDAV:

```
PUT /webdav/stats/{device}/{book}.json
{
  "total_time": 3600,
  "pages_read": 50,
  "percentage": 25
}
```

---

## Feature Roadmap Suggestions

### For KOmpanion (Maintaining Minimalism)

1. **Search/Filter** - Essential for large libraries
2. **Metadata Editing UI** - Fix incorrect extracted data
3. **Book Deletion** - Basic library management
4. **Collections/Tags** - Custom organization
5. **Description Field** - Enhanced metadata

### What KOmpanion Should NOT Add

To maintain its identity:

| Skip | Reason |
|------|--------|
| Built-in reader | KOReader is the reader |
| CBZ/CBR support | Out of ebook scope |
| Magic Shelves | Too complex |
| OIDC | Basic auth suffices |
| BookDrop watcher | Manual upload is simpler |

---

## Conclusion

| | KOmpanion | BookLore |
|--|-----------|----------|
| **Best for** | KOReader power users | General ebook enthusiasts |
| **Complexity** | Low | Medium-High |
| **Resources** | Minimal | Moderate |
| **Features** | Focused | Comprehensive |
| **Setup** | Simple | Docker required |

Both are excellent self-hosted solutions targeting different audiences. KOmpanion excels as a lightweight sync companion for dedicated KOReader users, while BookLore serves users wanting a full-featured library platform with all the bells and whistles.

---

## References

- KOmpanion: https://github.com/psolar/kompanion (or current repo)
- BookLore: https://github.com/booklore-app/booklore
- BookLore Demo: https://demo.booklore.org
- KOReader: https://github.com/koreader/koreader
