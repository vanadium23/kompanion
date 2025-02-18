# 8. Introduce device concept

Date: 2025-01-31

## Status

Accepted

## Context

In terms of authenication, we want to divide devices used with KOmpanion.

## Decision

1. Introduce new concept Device.
2. Device has a slug and a password.
3. You can add any number of device through web interface.
4. Device slug and password only used in authenication for progress, opds and webdav.
5. Device slug and password can not be used in authenication for web interface.

## Consequences

- positive: We can divide stats by device.
- positive: We can mark which device sent progress.
- negative: KOmpanion will be used only by single user.
