# 9. Store book files in DB

Date: 2025-01-25

## Status

Accepted

## Context

First decision to store files binary on filesystem was good. But it's a bit harder to maintain consistency in full backups: you need to backup DB and backup your local filestorage. Same logic applies to S3. If we take a look at calibre, all information is stored in filesystem. Let's make 

## Decision

1. Move book files to postgresql instead of filesystem by default.
2. We will use bytea instead of large object, because book is like tenth of megabytes.
3. Table will be only append only.

## Consequences

1. (+) We can check that files are present by join (with index only data).
2. (+) We can perform a single database backup to have all data.
3. (-) Database become larger.
