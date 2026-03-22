# Phase 5: Standalone KOReader Plugin with Native Highlight Extraction - Research

**Researched:** 2026-03-22
**Domain:** KOReader Lua plugin development, highlight extraction from DocSettings
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Implementation Decisions

### Plugin Architecture
- **D-01:** Standalone plugin under Tools menu (not Export highlights submenu)
- **D-02:** WidgetContainer-based plugin, NOT BaseExporter subclass
- **D-03:** Uses `clip.lua` patterns to read highlights directly from DocSettings

### Sync Behavior
- **D-04:** Manual trigger only — "Sync highlights" menu item
- **D-05:** Current book only — reads from current document's sidecar file
- **D-06:** Syncs to existing `/syncs/highlights` API endpoint

### Highlight Extraction
- **D-07:** Read `annotations` setting first (newer KOReader format)
- **D-08:** Fallback to `highlight` + `bookmarks` settings (legacy format)
- **D-09:** Get document hash from `partial_md5_checksum` in DocSettings
- **D-10:** Transform to Kompanion API format matching Phase 1 contract

### Configuration UI
- **D-11:** Three menu items: Setup, Sync Now, Help
- **D-12:** Setup dialog with URL, Device Name, Device Password fields
- **D-13:** Settings stored in `G_reader_settings:readSetting("kompanion")`

### Error Handling
- **D-14:** Show InfoMessage toast on success with synced count
- **D-15:** Show InfoMessage toast on failure with error message
- **D-16:** Log errors to KOReader logger for debugging

### Claude's Discretion
- Exact toast message text
- Plugin version number
- Help text content
- HTTP timeout values

### Deferred Ideas (OUT OF SCOPE)
- Auto-sync on book close — could be future enhancement
- Sync all books from history — out of scope for now
- Sync progress back to KOReader — out of scope (one-way sync per PROJECT.md)
- Sync status screen with history — keep it minimal

</user_constraints>

## Summary

Phase 5 creates a **standalone KOReader Lua plugin** that directly extracts highlights from document sidecar files (DocSettings) and syncs them to Kompanion. Unlike Phase 4's approach using the exporter.koplugin framework (which proved unreliable), this plugin is self-contained as a WidgetContainer-based plugin in the Tools menu.

The key insight from `clip.lua` is that highlights are stored in DocSettings under:
1. `annotations` (newer format, KOReader 2023+) — single array with page, text, note, chapter, drawer, color
2. `highlight` + `bookmarks` (legacy format) — page-indexed highlights with separate bookmarks for notes

**Primary recommendation:** Follow kosync.koplugin as the reference implementation for standalone plugin structure, HTTP requests, settings management, and menu integration. Use clip.lua patterns for highlight extraction.

---

## Standard Stack

### Core KOReader Components
| Component | Source | Purpose |
|-----------|--------|---------|
| WidgetContainer | `require("ui/widget/container/widgetcontainer")` | Base class for plugins with UI integration |
| InfoMessage | `require("ui/widget/infomessage")` | Toast notifications (success/error) |
| MultiInputDialog | `require("ui/widget/multiinputdialog")` | Setup dialog with multiple fields |
| UIManager | `require("ui/uimanager")` | Show/close widgets, schedule tasks |
| G_reader_settings | Global | Persistent settings storage |
| Device | `require("device")` | Device info (model name) |
| NetworkMgr | `require("ui/network/manager")` | Network connectivity handling |
| DocSettings | `require("docsettings")` | Access to sidecar file settings |

### HTTP/JSON Libraries (Bundled with KOReader)
| Library | Purpose |
|---------|---------|
| `socket.http` | HTTP requests |
| `ltn12` | Stream handling for HTTP body |
| `rapidjson` | JSON encoding/decoding |
| `socketutil` | Timeout management |
| `mime` | Base64 encoding for Basic Auth |
| `ffi/sha2` (md5) | MD5 hashing (already used in kosync) |

### Recommended Project Structure
```
koreader/kompanion.koplugin/
├── _meta.lua          # Plugin metadata (name, description)
└── main.lua           # Full plugin implementation
```

**Why this structure:**
- Single-file plugin (no need for separate client module like kosync)
- `_meta.lua` is required for plugin discovery by KOReader
- Cleaner than Phase 4's exporter-based approach

---

## Architecture Patterns

### Pattern 1: WidgetContainer Plugin Structure

**What:** KOReader plugins inherit from WidgetContainer and register with the menu system.

**When to use:** All standalone plugins that need menu integration and document context.

**Example (from kosync.koplugin/main.lua):**
```lua
local WidgetContainer = require("ui/widget/container/widgetcontainer")

local Kompanion = WidgetContainer:extend{
    name = "kompanion",
    is_doc_only = true,  -- Only active when document is open
}

function Kompanion:init()
    self.settings = G_reader_settings:readSetting("kompanion", {})
    self.ui.menu:registerToMainMenu(self)
end

function Kompanion:addToMainMenu(menu_items)
    menu_items.kompanion_sync = {
        text = _("Kompanion"),
        sub_item_table = {
            -- Menu items here
        }
    }
end

return Kompanion
```

**Source:** `/home/deploy/koreader/plugins/kosync.koplugin/main.lua` (lines 25-95)

### Pattern 2: Highlight Extraction from DocSettings

**What:** Read highlights directly from the document's sidecar file using DocSettings API.

**When to use:** When you need highlights without depending on exporter.koplugin.

**Example (adapted from clip.lua lines 362-385):**
```lua
local BookList = require("ui/widget/booklist")
local DocSettings = require("docsettings")

function Kompanion:getHighlightsFromCurrentDoc()
    -- Method 1: From in-memory annotations (if document is open)
    if self.ui and self.ui.annotation and self.ui.annotation.annotations then
        return self:parseAnnotationsFormat(self.ui.annotation.annotations)
    end

    -- Method 2: From sidecar file via DocSettings
    local doc_settings = self.ui.doc_settings
    local annotations = doc_settings:readSetting("annotations")

    if annotations then
        -- New format (KOReader 2023+)
        return self:parseAnnotationsFormat(annotations)
    else
        -- Legacy format
        local highlights = doc_settings:readSetting("highlight")
        local bookmarks = doc_settings:readSetting("bookmarks")
        return self:parseLegacyFormat(highlights, bookmarks)
    end
end

function Kompanion:parseAnnotationsFormat(annotations)
    local highlights = {}
    for _, item in ipairs(annotations) do
        if item.text and item.text ~= "" then
            table.insert(highlights, {
                text = item.text,
                note = item.note or "",
                page = item.pageref or item.pageno or "",
                chapter = item.chapter or "",
                time = self:parseTimestamp(item.datetime),
                drawer = item.drawer or "",
                color = item.color or "",
            })
        end
    end
    return highlights
end
```

**Source:** `/home/deploy/koreader/plugins/exporter.koplugin/clip.lua` (lines 260-277)

### Pattern 3: HTTP Request with Basic Auth

**What:** Make synchronous HTTP POST with JSON body and Basic Auth headers.

**When to use:** Syncing highlights to Kompanion API.

**Example (adapted from base.lua and kosync patterns):**
```lua
local http = require("socket.http")
local ltn12 = require("ltn12")
local rapidjson = require("rapidjson")
local socketutil = require("socketutil")
local mime = require("mime")

function Kompanion:makeJsonRequest(url, method, body, headers)
    local sink = {}
    local body_json = rapidjson.encode(body)
    local source = ltn12.source.string(body_json)

    socketutil:set_timeout(5, 15)  -- 5s connect, 15s total

    local request = {
        url = url,
        method = method,
        sink = ltn12.sink.table(sink),
        source = source,
        headers = {
            ["Content-Length"] = #body_json,
            ["Content-Type"] = "application/json",
        },
    }

    -- Merge extra headers (e.g., Authorization)
    for k, v in pairs(headers or {}) do
        request.headers[k] = v
    end

    local code, _, status = socket.skip(1, http.request(request))
    socketutil:reset_timeout()

    if code ~= 200 then
        return nil, status or code or "network unreachable"
    end

    local response = rapidjson.decode(table.concat(sink))
    return response
end

function Kompanion:syncHighlights()
    local url = self.settings.url .. "/syncs/highlights"
    local auth = mime.b64(self.settings.device_name .. ":" .. self.settings.device_password)

    local body = {
        document = self:getDocumentHash(),
        title = self:getDocumentTitle(),
        author = self:getDocumentAuthor(),
        highlights = self:getHighlightsFromCurrentDoc(),
    }

    local response, err = self:makeJsonRequest(url, "POST", body, {
        ["Authorization"] = "Basic " .. auth,
    })

    return response, err
end
```

**Source:** `/home/deploy/koreader/plugins/exporter.koplugin/base.lua` (lines 161-211)

### Pattern 4: Settings Management

**What:** Use G_reader_settings for persistent plugin configuration.

**When to use:** Storing URL, device name, password.

**Example:**
```lua
local Kompanion = WidgetContainer:extend{
    name = "kompanion",
    default_settings = {
        url = nil,
        device_name = nil,
        device_password = nil,
    },
}

function Kompanion:init()
    self.settings = G_reader_settings:readSetting("kompanion", self.default_settings)
    self.ui.menu:registerToMainMenu(self)
end

function Kompanion:saveSettings()
    G_reader_settings:saveSetting("kompanion", self.settings)
end
```

**Source:** `/home/deploy/koreader/plugins/kosync.koplugin/main.lua` (lines 57-67, 85-86)

### Anti-Patterns to Avoid

1. **Inheriting from BaseExporter:** Phase 4 proved this doesn't work reliably with the exporter Provider system. Use WidgetContainer directly.

2. **Blocking UI without scheduling:** Always wrap network calls in `UIManager:scheduleIn()` or use callbacks.

3. **Ignoring legacy highlight format:** Older KOReader versions use `highlight`+`bookmarks`, not `annotations`. Must support both.

4. **Not checking document is open:** Use `is_doc_only = true` and check `self.ui.document` before accessing.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| HTTP requests | Custom socket code | `socket.http` + `ltn12` pattern from base.lua | KOReader uses this everywhere |
| JSON encoding | Manual string building | `rapidjson.encode/decode` | Bundled, battle-tested |
| Settings persistence | File I/O | `G_reader_settings:readSetting/saveSetting` | KOReader standard |
| Toast notifications | Custom dialog | `UIManager:show(InfoMessage:new{...})` | Consistent UX |
| Timeout handling | socket.timeout | `socketutil:set_timeout/reset_timeout` | KOReader utility |
| Basic Auth | Manual header crafting | `mime.b64(user:pass)` | Standard encoding |
| Document hash | Calculate MD5 | `self.ui.doc_settings:readSetting("partial_md5_checksum")` | Already stored |

---

## Common Pitfalls

### Pitfall 1: Wrong Base Class
**What goes wrong:** Using BaseExporter requires exporter.koplugin to be loaded and Provider system to work.
**Why it happens:** Phase 4 attempted this approach.
**How to avoid:** Use WidgetContainer as base class, like kosync.koplugin.
**Warning signs:** Plugin not appearing in menu, export callbacks not firing.

### Pitfall 2: Missing Legacy Highlight Support
**What goes wrong:** Plugin only reads `annotations` setting, missing older KOReader highlights.
**Why it happens:** Newer KOReader uses `annotations`, older uses `highlight`+`bookmarks`.
**How to avoid:** Follow clip.lua pattern: check `annotations` first, fallback to `highlight`+`bookmarks`.
**Warning signs:** Users with older highlights report "0 highlights found".

### Pitfall 3: Not Waiting for Network
**What goes wrong:** HTTP request fails on devices with WiFi that needs to be enabled.
**Why it happens:** KOReader devices may have WiFi off to save battery.
**How to avoid:** Use `NetworkMgr:willRerunWhenOnline(callback)` pattern from kosync.
**Warning signs:** "network unreachable" errors on sync.

### Pitfall 4: Incorrect Timestamp Parsing
**What goes wrong:** KOReader stores timestamps as ISO strings, Kompanion expects Unix timestamps.
**Why it happens:** `item.datetime` is a string like "2024-01-15 10:30:00".
**How to avoid:** Use `os.time()` with parsed date, or leverage clip.lua's `getTime()` function.
**Warning signs:** Highlights showing epoch 0 or wrong dates in Kompanion.

### Pitfall 5: UI Blocking During Network Request
**What goes wrong:** KOReader freezes while waiting for HTTP response.
**Why it happens:** Socket.http is synchronous by default.
**How to avoid:** Use `UIManager:scheduleIn(0.5, callback)` to run sync in background.
**Warning signs:** Device becomes unresponsive during sync.

---

## Code Examples

### Complete Plugin Skeleton

```lua
-- koreader/kompanion.koplugin/main.lua
local Device = require("device")
local InfoMessage = require("ui/widget/infomessage")
local MultiInputDialog = require("ui/widget/multiinputdialog")
local NetworkMgr = require("ui/network/manager")
local UIManager = require("ui/uimanager")
local WidgetContainer = require("ui/widget/container/widgetcontainer")
local http = require("socket.http")
local ltn12 = require("ltn12")
local logger = require("logger")
local mime = require("mime")
local rapidjson = require("rapidjson")
local socketutil = require("socketutil")
local _ = require("gettext")

local Kompanion = WidgetContainer:extend{
    name = "kompanion",
    is_doc_only = true,
}

Kompanion.default_settings = {
    url = nil,
    device_name = nil,
    device_password = nil,
}

function Kompanion:init()
    self.settings = G_reader_settings:readSetting("kompanion", self.default_settings)
    self.ui.menu:registerToMainMenu(self)
end

function Kompanion:addToMainMenu(menu_items)
    menu_items.kompanion = {
        text = _("Kompanion Highlights"),
        sub_item_table = {
            {
                text = _("Setup"),
                keep_menu_open = true,
                callback = function() self:showSetupDialog() end,
            },
            {
                text = _("Sync highlights"),
                enabled_func = function() return self:isConfigured() end,
                callback = function() self:doSync() end,
            },
            {
                text = _("Help"),
                keep_menu_open = true,
                callback = function() self:showHelp() end,
            },
        }
    }
end

function Kompanion:isConfigured()
    return self.settings.url and self.settings.device_name and self.settings.device_password
end

function Kompanion:showSetupDialog()
    local dialog
    dialog = MultiInputDialog:new{
        title = _("Setup Kompanion"),
        fields = {
            { description = _("Server URL"), hint = "http://192.168.1.100:8080", text = self.settings.url },
            { description = _("Device Name"), text = self.settings.device_name },
            { description = _("Device Password"), text = self.settings.device_password, text_type = "password" },
        },
        buttons = {
            { { text = _("Cancel"), callback = function() UIManager:close(dialog) end },
              { text = _("Save"), callback = function()
                  local fields = dialog:getFields()
                  self.settings.url = fields[1]
                  self.settings.device_name = fields[2]
                  self.settings.device_password = fields[3]
                  G_reader_settings:saveSetting("kompanion", self.settings)
                  UIManager:close(dialog)
              end },
            },
        },
    }
    UIManager:show(dialog)
    dialog:onShowKeyboard()
end

function Kompanion:doSync()
    if not self:isConfigured() then
        UIManager:show(InfoMessage:new{ text = _("Please configure Kompanion first."), timeout = 3 })
        return
    end

    if NetworkMgr:willRerunWhenOnline(function() self:doSync() end) then
        return
    end

    UIManager:scheduleIn(0.5, function() self:performSync() end)
    UIManager:show(InfoMessage:new{ text = _("Syncing highlights..."), timeout = 1 })
end

function Kompanion:performSync()
    -- Extract highlights and send to server
    -- (implementation details in sync logic section)
end

function Kompanion:showHelp()
    UIManager:show(InfoMessage:new{
        text = _([[Sync highlights to your Kompanion server.

1. Configure URL, device name, and password
2. Open a book with highlights
3. Tap "Sync highlights"

Make sure your KOReader and Kompanion server are on the same network.]]),
    })
end

return Kompanion
```

### Highlight Extraction Logic

```lua
function Kompanion:getDocumentHash()
    return self.ui.doc_settings:readSetting("partial_md5_checksum")
end

function Kompanion:getDocumentTitle()
    local props = self.ui.doc_settings:readSetting("doc_props") or {}
    return props.title or self:getFilename()
end

function Kompanion:getDocumentAuthor()
    local props = self.ui.doc_settings:readSetting("doc_props") or {}
    return props.authors or ""
end

function Kompanion:getHighlights()
    local doc_settings = self.ui.doc_settings
    local annotations = doc_settings:readSetting("annotations")

    if annotations then
        return self:parseNewFormat(annotations)
    else
        local highlights = doc_settings:readSetting("highlight")
        local bookmarks = doc_settings:readSetting("bookmarks")
        return self:parseLegacyFormat(highlights, bookmarks)
    end
end

function Kompanion:parseNewFormat(annotations)
    local highlights = {}
    for _, item in ipairs(annotations) do
        if item.text and item.text ~= "" then
            table.insert(highlights, {
                text = item.text,
                note = item.note or "",
                page = tostring(item.pageref or item.pageno or ""),
                chapter = item.chapter or "",
                time = self:parseDateTime(item.datetime),
                drawer = item.drawer or "",
                color = item.color or "",
            })
        end
    end
    return highlights
end

function Kompanion:parseLegacyFormat(highlights, bookmarks)
    local result = {}
    if not highlights then return result end

    for page, items in pairs(highlights) do
        for _, item in ipairs(items) do
            if item.text and item.text ~= "" then
                local note = ""
                -- Look for matching bookmark for note
                if bookmarks then
                    for _, bm in ipairs(bookmarks) do
                        if bm.datetime == item.datetime and bm.text then
                            note = bm.text
                            break
                        end
                    end
                end
                table.insert(result, {
                    text = item.text,
                    note = note,
                    page = tostring(page),
                    chapter = item.chapter or "",
                    time = self:parseDateTime(item.datetime),
                    drawer = item.drawer or "",
                    color = item.color or "",
                })
            end
        end
    end
    return result
end

function Kompanion:parseDateTime(datetime_str)
    if not datetime_str then return 0 end
    -- Parse "2024-01-15 10:30:00" format
    local y, m, d, h, min, sec = datetime_str:match("(%d+)-(%d+)-(%d+) (%d+):(%d+):(%d+)")
    if y then
        return os.time{ year = y, month = m, day = d, hour = h, min = min, sec = sec }
    end
    return 0
end
```

### HTTP Sync Logic

```lua
function Kompanion:performSync()
    local body = {
        document = self:getDocumentHash() or "",
        title = self:getDocumentTitle() or "",
        author = self:getDocumentAuthor() or "",
        highlights = self:getHighlights(),
    }

    if #body.highlights == 0 then
        UIManager:show(InfoMessage:new{ text = _("No highlights found in this book."), timeout = 3 })
        return
    end

    local url = self.settings.url
    if not url:match("/$") then url = url .. "/" end
    url = url .. "syncs/highlights"

    local auth = mime.b64(self.settings.device_name .. ":" .. self.settings.device_password)
    local response, err = self:makeJsonRequest(url, "POST", body, {
        ["Authorization"] = "Basic " .. auth,
    })

    if response and response.synced then
        UIManager:show(InfoMessage:new{
            text = T(_("Synced %1 of %2 highlights."), response.synced, response.total),
            timeout = 3,
        })
    else
        UIManager:show(InfoMessage:new{
            text = T(_("Sync failed: %1"), err or "unknown error"),
            timeout = 3,
        })
        logger.warn("Kompanion sync error:", err)
    end
end
```

---

## Kompanion API Contract

### POST /syncs/highlights

**Request body:**
```json
{
    "document": "abc123def456",      // partial MD5 (required)
    "title": "Book Title",
    "author": "Author Name",
    "highlights": [
        {
            "text": "Highlighted text",
            "note": "User note",
            "page": "42",
            "chapter": "Chapter 1",
            "time": 1705312200,      // Unix timestamp
            "drawer": "lighten",      // highlight style
            "color": "yellow"
        }
    ]
}
```

**Response (200 OK):**
```json
{
    "synced": 5,
    "total": 5
}
```

**Source:** `/home/deploy/kompanion/internal/controller/http/v1/highlight.go`

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing with testify |
| Config file | None (standard Go test) |
| Quick run command | `go test ./internal/highlight/... -v -short` |
| Full suite command | `go test ./... -v` |

### Phase Requirements -> Test Map

This phase is a **KOReader Lua plugin** - testing is manual/integration:
- No automated unit tests for Lua code in Kompanion repo
- Testing requires installing plugin in KOReader device
- Integration test: sync highlights from KOReader, verify in Kompanion DB

### Manual Test Checklist
1. [ ] Plugin appears in KOReader Tools menu
2. [ ] Setup dialog saves URL, device name, password
3. [ ] Sync shows "No highlights" when book has none
4. [ ] Sync sends highlights to Kompanion `/syncs/highlights`
5. [ ] Success toast shows synced count
6. [ ] Error toast shows on network/auth failure
7. [ ] Settings persist across KOReader restarts

### Wave 0 Gaps
None - this is Lua plugin development. No Go test infrastructure changes needed.

---

## Sources

### Primary (HIGH confidence)
- `/home/deploy/koreader/plugins/kosync.koplugin/main.lua` - WidgetContainer pattern, menu integration, settings, network handling
- `/home/deploy/koreader/plugins/exporter.koplugin/clip.lua` (lines 260-385) - Highlight extraction from DocSettings
- `/home/deploy/koreader/plugins/exporter.koplugin/base.lua` (lines 161-211) - HTTP request pattern with JSON
- `/home/deploy/kompanion/internal/controller/http/v1/highlight.go` - API endpoint contract
- `/home/deploy/kompanion/internal/entity/highlight.go` - Highlight entity structure

### Secondary (MEDIUM confidence)
- `/home/deploy/koreader/frontend/ui/widget/booklist.lua` (line 335-341) - `BookList.getDocSettings()` function
- `/home/deploy/koreader/plugins/kosync.koplugin/KOSyncClient.lua` - HTTP client pattern with Spore (alternative approach)

### Tertiary (LOW confidence)
- Existing Phase 4 target.lua - Shows what NOT to do (BaseExporter approach)

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Direct source from kosync.koplugin and clip.lua
- Architecture: HIGH - WidgetContainer pattern is standard KOReader plugin approach
- Pitfalls: HIGH - Based on actual Phase 4 failure analysis
- API contract: HIGH - Verified from Go source code

**Research date:** 2026-03-22
**Valid until:** KOReader plugin API is stable - valid for 1+ year
