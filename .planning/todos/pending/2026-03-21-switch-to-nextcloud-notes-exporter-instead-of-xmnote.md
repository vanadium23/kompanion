---
created: 2026-03-21T20:21:51.776Z
title: Switch to Nextcloud Notes exporter instead of XMNote
area: api
files:
  - /home/deploy/koreader/plugins/exporter.koplugin/target/xmnote.lua
  - /home/deploy/koreader/plugins/exporter.koplugin/target/nextcloud.lua
  - internal/controller/http/v1/highlight.go
---

## Problem

XMNote exporter protocol has no authentication mechanism. This creates security concerns for a self-hosted Kompanion server exposed to the network. Anyone who knows the server IP could send highlights.

Phase 3 was originally planned for XMNote API endpoint (`POST /send`), but the lack of auth makes it unsuitable for production use.

## Solution

Switch to Nextcloud Notes exporter protocol instead:
- Nextcloud exporter supports authentication (username/password)
- KOReader already has this exporter built-in
- Same exporter plugin infrastructure, just different target

**Implementation approach:**
1. Research Nextcloud Notes API format in `/home/deploy/koreader/plugins/exporter.koplugin/target/nextcloud.lua`
2. Create endpoint compatible with Nextcloud Notes protocol
3. Use existing device authentication (MD5 hash) or basic auth
4. Map Nextcloud format to highlight_annotations table

**Key decision:** Replace Phase 3 goal from "XMNote API endpoint" to "Nextcloud Notes API endpoint"
