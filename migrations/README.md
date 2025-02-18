# Schema convention

- `_at` - postfix for timestamp fields
- `is_` - prefix for boolean fields
- `koreader_partial_md5` - text field for document matching method
    - implemented in [utils](/pkg/utils/koreader.go)
- foreign keys only inside one package, but not between
    - example: koreader has more statistics, that books uploaded to library
- prefer to add comment on schema: https://www.postgresql.org/docs/current/sql-comment.html
