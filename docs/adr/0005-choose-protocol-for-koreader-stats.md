# 5. Choose Protocol for KOReader Statistics

Date: 2024-09-24

## Status

Accepted

## Context

Inside KOReader, there is cloud sync for obtaining reading statistics. Essentially, this is the main purpose of its implementation. For synchronization, they use an SQLite database, which can be placed on:
- FTP
- Dropbox
- WebDAV

Additionally, according to their expectations, they can merge data between different devices. The logic for this is written [here](https://github.com/koreader/koreader/blob/master/frontend/apps/cloudstorage/syncservice.lua#L104-L123).

## Decision

For the synchronization protocol, WebDAV was chosen.
The main advantages of this solution are:
1. KOReader supports syncing between devices for this protocol.
2. Golang has built-in support for WebDAV (see cmd/webdav/main.go).
3. WebDAV uses basic authentication, which we can reuse with OPDS.

Alternatives:
1. The Dropbox protocol does not make sense to implement because it is proprietary.
2. The FTP protocol has exactly the opposite issues compared to the advantages of WebDAV.

## Consequences

- (Pros) All complexity has already been implemented in KOReader.
- (Cons) We need to think about how to synchronize the state between KOReader's SQLite and KOmpanion's PostgreSQL.
- (Cons) There is a tight coupling with the DDL schema of KOReader.
 