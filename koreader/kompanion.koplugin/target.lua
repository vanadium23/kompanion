local DocSettings = require("docsettings")
local InfoMessage = require("ui/widget/infomessage")
local MultiInputDialog = require("ui/widget/multiinputdialog")
local UIManager = require("ui/uimanager")
local mime = require("mime")
local logger = require("logger")
local T = require("ffi/util").template
local _ = require("gettext")

-- Kompanion exporter for highlights sync
local KompanionExporter = require("base"):new {
    name = "kompanion",
    is_remote = true,
}

function KompanionExporter:isReadyToExport()
    -- Need url, device_name and device_password configured
    return self.settings.url and self.settings.device_name and self.settings.device_password
end

function KompanionExporter:getMenuTable()
    local dialog_title = _("Setup Kompanion plugin")
    return {
        text = _("Kompanion"),
        checked_func = function() return self:isEnabled() end,
        sub_item_table = {
            {
                text = dialog_title,
                keep_menu_open = true,
                callback = function()
                    self:showSetupDialog()
                end
            },
            {
                text = _("Export to Kompanion"),
                checked_func = function() return self:isEnabled() end,
                callback = function() self:toggleEnabled() end,
            },
            {
                text = _("Help"),
                keep_menu_open = true,
                callback = function()
                    UIManager:show(InfoMessage:new {
                        text = T(_([[Export highlights to your Kompanion server.

1. Configure your Kompanion server URL (e.g., http://192.168.1.100:8080)
2. Enter device name (must match name in Kompanion's Devices page)
3. Enter device password from Kompanion's Devices page

Make sure your KOReader and Kompanion server are on the same network.]])
                        )
                    })
                end
            }
        }
    }
end

function KompanionExporter:showSetupDialog()
    local dialog
    dialog = MultiInputDialog:new {
        title = _("Setup Kompanion"),
        fields = {
            {
                description = _("Server URL"),
                hint = "http://192.168.1.100:8080",
                text = self.settings.url,
                input_type = "string"
            },
            {
                description = _("Device Name"),
                hint = _("Name from Kompanion Devices page"),
                text = self.settings.device_name,
                input_type = "string"
            },
            {
                description = _("Device Password"),
                hint = _("Password from Kompanion Devices page"),
                text = self.settings.device_password,
                text_type = "password",
                input_type = "string"
            }
        },
        buttons = {
            {
                {
                    text = _("Cancel"),
                    callback = function()
                        UIManager:close(dialog)
                    end
                },
                {
                    text = _("OK"),
                    callback = function()
                        local fields = dialog:getFields()
                        local url = fields[1]
                        local device_name = fields[2]
                        local device_password = fields[3]
                        if url ~= "" then
                            self.settings.url = url
                            self:saveSettings()
                        end
                        if device_name ~= "" then
                            self.settings.device_name = device_name
                            self:saveSettings()
                        end
                        if device_password ~= "" then
                            self.settings.device_password = device_password
                            self:saveSettings()
                        end
                        UIManager:close(dialog)
                    end
                }
            }
        }
    }
    UIManager:show(dialog)
    dialog:onShowKeyboard()
end

function KompanionExporter:createRequestBody(booknotes)
    -- Get document hash from DocSettings
    local doc_settings = DocSettings:open(booknotes.file)
    local partial_md5 = doc_settings:readSetting("partial_md5_checksum")

    -- Build request body matching Kompanion API format
    local request_body = {
        document = partial_md5 or "",
        title = booknotes.title or "",
        author = booknotes.author or "",
        highlights = {}
    }

    -- Transform booknotes chapters to highlights array
    for _, chapter in ipairs(booknotes) do
        for _, clipping in ipairs(chapter) do
            local highlight = {
                text = clipping.text or "",
                note = clipping.note or "",
                page = clipping.page or "",
                chapter = clipping.chapter or "",
                time = clipping.time or 0,
                drawer = clipping.drawer or "",
                color = clipping.color or ""
            }
            table.insert(request_body.highlights, highlight)
        end
    end

    return request_body
end

function KompanionExporter:export(t)
    if not self:isReadyToExport() then
        logger.warn("KompanionExporter: not ready to export")
        return false
    end

    -- Build Basic Auth header: device_name:device_password
    local auth = mime.b64(self.settings.device_name .. ":" .. self.settings.device_password)
    local headers = {
        ["Authorization"] = "Basic " .. auth,
    }

    local total_synced = 0
    local total_highlights = 0

    -- Export each book's highlights
    for _, booknotes in ipairs(t) do
        if booknotes.file then
            local body = self:createRequestBody(booknotes)
            total_highlights = total_highlights + #body.highlights

            -- Build URL with /syncs/highlights endpoint
            local url = self.settings.url
            if not url:match("/$") then
                url = url .. "/"
            end
            url = url .. "syncs/highlights"

            local response, err = self:makeJsonRequest(url, "POST", body, headers)
            if not response then
                logger.warn("KompanionExporter: error syncing highlights", err)
                UIManager:show(InfoMessage:new {
                    text = T(_("Failed to sync highlights: %1"), err or "unknown error"),
                    timeout = 3,
                })
                return false
            end

            -- D-13: Extract synced count from response
            if response.synced then
                total_synced = total_synced + response.synced
            end

            logger.dbg("KompanionExporter: synced", response.synced, "of", response.total, "highlights")
        end
    end

    -- D-13: Show success toast with synced count
    if total_synced > 0 then
        UIManager:show(InfoMessage:new {
            text = T(_("Synced %1 highlights to Kompanion"), total_synced),
            timeout = 3,
        })
    end

    return true
end

return KompanionExporter
