local Provider = require("provider")
local KompanionTarget = require("target")

-- Register Kompanion as an exporter target
Provider:register("exporter", "kompanion", KompanionTarget)
