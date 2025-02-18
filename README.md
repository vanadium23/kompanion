# KOmpanion - bookshelf companion for KOreader

[![Go Report Card](https://goreportcard.com/badge/github.com/vanadium23/kompanion/)](https://goreportcard.com/report/github.com/vanadium23/kompanion/)
[![License](https://img.shields.io/github/license/vanadium23/kompanion.svg)](https://github.com/vanadium23/kompanion/blob/master/LICENSE)
[![Release](https://img.shields.io/github/v/release/vanadium23/kompanion.svg)](https://github.com/vanadium23/kompanion//releases/)

KOmpanion is a minimalistic library web application, that tightly coupled to KOReader features.
Main features are:
- upload and view your bookshelf
- OPDS to download books
- KOReader sync progress API 
- KOReader book stats via WebDAV

What KOmpanion is NOT about:
- web interface for book reading (just install KOReader)
- converter between formats (I don't want to do another calibre)

## Why KOReader for all?

KOReader is the best available reader on the market (personal opinion).
Features, that can buy you in:
- sync progress between tablet, phone and ebook
- extensive stats for book reading: total time, time per page, estimates

## Installation

### Docker (preferred)

1. you need a postgresql instance
2. run `docker run -e KOMPANION_PG_URL=postgres://... -e KOMPANION_AUTH_PASSWORD=password -e KOMPANION_AUTH_USERNAME=username kompanion` , where you pass pg url and admin username and password to init

### Configuration

- KOMPANION_AUTH_USERNAME - required for setup
- KOMPANION_AUTH_PASSWORD - required for setup
- KOMPANION_AUTH_STORAGE - postgres or memory (default: postgres)
- KOMPANION_HTTP_PORT - port for service (default: postgres)
- KOMPANION_LOG_LEVEL - debug, info, error (default: info)
- KOMPANION_PG_POOL_MAX - integer number for pooling connections (default: 2)
- KOMPANION_PG_URL - postgresql link
- KOMPANION_BSTORAGE_TYPE - type of storage for books: postgres, memory, filesystem (default: postgres)
- KOMPANION_BSTORAGE_PATH - path in case of filesystem
- KOMPANION_STATS_TYPE - type of temporary storage for uploaded sqlite3 stats files: postgres, memory, filesystem (default: memory)
- KOMPANION_STATS_PATH - path in case of filesystem

## Usage

### Web interface

First of all, you need to add your devices:
1. Go to service
2. Login
3. Click devices
4. Add device name and password

**Warning:** password for device stored as md5 hash without salt to be compatible with [kosync plugin](https://github.com/koreader/koreader/blob/master/plugins/kosync.koplugin/main.lua#L544).

### KOReader

Go to following plugins:
1. Cloud storage
    1. Add new WebDAV: URL - `https://your-kompanion.org/webdav/`, username - device name, password - password
2. Statistics - Settings - Cloud sync
    1. It's OKAY to have empty list, just press on **Long press to choose current folder**.
3. Open book - tools - Progress sync
    1. Custom sync server: `https://your-kompanion.org/`
    1. Login: username - device name, password - password

## Development

Project was started with [go-clean-template](https://github.com/evrone/go-clean-template), but then heavily modified.

Local development:
```sh
# Postgres
$ make compose-up
# Run app with migrations
$ make run
```

Integration tests (can be run in CI):
```sh
# DB, app + migrations, integration tests
$ make compose-up-integration-test
```
