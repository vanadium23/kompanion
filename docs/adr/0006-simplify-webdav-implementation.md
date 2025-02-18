# 6. Simplify WebDAV Implementation

Date: 2024-09-25

## Status

Accepted

## Context

In the previous ADR, a choice was made to use WebDAV. A vanilla implementation in Golang was selected for this purpose.
Two issues were discovered:
1. In Gin, it was necessary to integrate through Any, Handle, WrapH.
2. KOReader only selects the directory and performs GET/PUT operations.

Logs from production:
```
[GIN] 2024/09/25 - 05:52:09 | 207 |  1.483733941s |  172.71.172.229 | PROPFIND  "/webdav/"
[GIN] 2024/09/25 - 05:52:09 | 207 |  1.483966387s |  172.71.172.229 | PROPFIND  "/webdav/"
[GIN] 2024/09/25 - 05:52:09 | 207 |   1.48404868s |  172.71.172.229 | PROPFIND  "/webdav/"
[GIN] 2024/09/25 - 05:52:17 | 207 |  1.462503894s |  162.158.111.82 | PROPFIND  "/webdav/"
[GIN] 2024/09/25 - 05:52:17 | 207 |  1.462736363s |  162.158.111.82 | PROPFIND  "/webdav/"
[GIN] 2024/09/25 - 05:52:17 | 207 |  1.462821722s |  162.158.111.82 | PROPFIND  "/webdav/"
[GIN] 2024/09/25 - 05:52:36 | 200 |  1.487182474s |  172.69.151.186 | GET      "/webdav/statistics.sqlite3"
[GIN] 2024/09/25 - 05:52:36 | 200 |  1.488056255s |  172.69.151.186 | GET      "/webdav/statistics.sqlite3"
[GIN] 2024/09/25 - 05:52:36 | 200 |  1.488237468s |  172.69.151.186 | GET      "/webdav/statistics.sqlite3"
[GIN] 2024/09/25 - 05:52:39 | 201 |  1.481713443s |  172.70.243.149 | PUT      "/webdav/statistics.sqlite3"
[GIN] 2024/09/25 - 05:52:39 | 201 |   1.48192425s |  172.70.243.149 | PUT      "/webdav/statistics.sqlite3"
[GIN] 2024/09/25 - 05:52:39 | 201 |  1.482005659s |  172.70.243.149 | PUT      "/webdav/statistics.sqlite3"
```

## Decision

The implementation will be simplified using custom code.
The API GET/PUT has a static path since the uploaded file is `statistics.sqlite3`.
The PROPFIND API is part of the WebDAV RFC: https://datatracker.ietf.org/doc/html/rfc4918#section-9.1

Overall, the implementation for responses can be "copied" from the vanilla implementation: https://cs.opensource.google/go/x/net/+/master:webdav/webdav.go;drc=4542a42604cd159f1adb93c58368079ae37b3bf6;l=582 

## Consequences

- (+) It will be possible to write a use case with a ready filesystem.
- (-) It will not be possible to change the directory.
- (-) There is a risk that the vanilla implementation will need to be restored.
