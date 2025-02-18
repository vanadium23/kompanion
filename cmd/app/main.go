package main

import (
	"log"

	"gitea.chrnv.ru/vanadium23/kompanion/config"
	"gitea.chrnv.ru/vanadium23/kompanion/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
