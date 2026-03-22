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
local T = require("ffi/util").template
local _ = require("gettext")

local Kompanion = WidgetContainer:extend{
    name = "kompanion",
    is_doc_only = true,  -- Only active when document is open
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
    menu_items.kompanion_highlights = {
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
    return self.settings.url and self.settings.url ~= ""
        and self.settings.device_name and self.settings.device_name ~= ""
        and self.settings.device_password and self.settings.device_password ~= ""
end

function Kompanion:showSetupDialog()
    local dialog
    dialog = MultiInputDialog:new{
        title = _("Setup Kompanion"),
        fields = {
            {
                description = _("Server URL"),
                hint = "http://192.168.1.100:8080",
                text = self.settings.url or "",
            },
            {
                description = _("Device Name"),
                hint = _("Name from Kompanion Devices page"),
                text = self.settings.device_name or "",
            },
            {
                description = _("Device password"),
                hint = _("Password from Kompanion Devices page"),
                text = self.settings.device_password or "",
                text_type = "password",
            },
        },
        buttons = {
            {
                {
                    text = _("Cancel"),
                    id = "close",
                    callback = function()
                        UIManager:close(dialog)
                    end,
                },
                {
                    text = _("Save"),
                    is_enter_default = true,
                    callback = function()
                        local fields = dialog:getFields()
                        self.settings.url = fields[1] ~= "" and fields[1] or nil
                        self.settings.device_name = fields[2] ~= "" and fields[2] or nil
                        self.settings.device_password = fields[3] ~= "" and fields[3] or nil
                        G_reader_settings:saveSetting("kompanion", self.settings)
                        UIManager:close(dialog)
                    end,
                },
            },
        },
    }
    UIManager:show(dialog)
    dialog:onShowKeyboard()
end

function Kompanion:doSync()
    if not self:isConfigured() then
        UIManager:show(InfoMessage:new{
            text = _("Please configure Kompanion first using Setup."),
            timeout = 3,
        })
        return
    end

    -- Wait for network if not online
    if NetworkMgr:willRerunWhenOnline(function() self:doSync() end) then
        return
    end

    -- Schedule sync to avoid blocking UI
    UIManager:show(InfoMessage:new{
        text = _("Syncing highlights..."),
        timeout = 1,
    })
    UIManager:scheduleIn(0.5, function() self:performSync() end)
end

function Kompanion:performSync()
    local highlights = self:getHighlights()

    if #highlights == 0 then
        UIManager:show(InfoMessage:new{
            text = _("No highlights found in this book."),
            timeout = 3,
        })
        return
    end

    local body = {
        document = self:getDocumentHash() or "",
        title = self:getDocumentTitle() or "",
        author = self:getDocumentAuthor() or "",
        highlights = highlights,
    }

    local url = self.settings.url
    if not url:match("/$") then url = url .. "/" end
    url = url .. "syncs/highlights"

    local auth = mime.b64(self.settings.device_name .. ":" .. self.settings.device_password)
    local response, err = self:makeJsonRequest(url, "POST", body, {
        ["Authorization"] = "Basic " .. auth,
    })

    if response and response.synced then
        -- Show success toast with synced count
        UIManager:show(InfoMessage:new{
            text = T(_("Synced %1 of %2 highlights."), response.synced, response.total),
            timeout = 3,
        })
        logger.dbg("Kompanion: synced", response.synced, "of", response.total, "highlights")
    else
        -- Show error toast
        UIManager:show(InfoMessage:new{
            text = T(_("Sync failed: %1"), err or "unknown error"),
            timeout = 3,
        })
        -- Log error for debugging
        logger.warn("Kompanion: sync error:", err)
    end
end

function Kompanion:getDocumentHash()
    return self.ui.doc_settings:readSetting("partial_md5_checksum")
end

function Kompanion:getDocumentTitle()
    local props = self.ui.doc_settings:readSetting("doc_props") or {}
    if props.title and props.title ~= "" then
        return props.title
    end
    -- Fallback to filename
    local file = self.ui.document.file
    if file then
        local _, name = file:match("(.*/)(.*)")
        return name or file
    end
    return "Unknown"
end

function Kompanion:getDocumentAuthor()
    local props = self.ui.doc_settings:readSetting("doc_props") or {}
    return props.authors or ""
end

function Kompanion:getHighlights()
    local doc_settings = self.ui.doc_settings
    local annotations = doc_settings:readSetting("annotations")

    if annotations then
        -- New format (KOReader 2023+)
        return self:parseNewFormat(annotations)
    else
        -- Legacy format
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
        return os.time({
            year = tonumber(y),
            month = tonumber(m),
            day = tonumber(d),
            hour = tonumber(h),
            min = tonumber(min),
            sec = tonumber(sec)
        })
    end
    return 0
end

function Kompanion:makeJsonRequest(url, method, body, headers)
    local sink = {}
    local body_json, err = rapidjson.encode(body)
    if not body_json then
        return nil, "cannot encode request body: " .. (err or "unknown error")
    end

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
        return nil, status or tostring(code) or "network unreachable"
    end

    if not sink[1] then
        return nil, "no response from server"
    end

    local response
    response, err = rapidjson.decode(table.concat(sink))
    if not response then
        return nil, "cannot decode response: " .. (err or "unknown error")
    end

    return response
end

function Kompanion:showHelp()
    UIManager:show(InfoMessage:new{
        text = _([[Sync highlights from current book to your Kompanion server.

1. Configure URL, device name, and password via Setup
2. Open a book with highlights
3. Tap "Sync highlights" from Tools menu

Make sure your device and Kompanion server are on the same network.]]),
        timeout = 5,
    })
end

return Kompanion
