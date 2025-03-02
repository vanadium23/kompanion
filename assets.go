package kompanion

import (
	"embed"
)

//go:embed migrations/*.sql
var Migrations embed.FS

//go:embed web/*
var WebAssets embed.FS
