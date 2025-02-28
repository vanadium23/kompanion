package main

import (
	"log"

	"github.com/vanadium23/kompanion/config"
	"github.com/vanadium23/kompanion/internal/app"
)

var Version = "dev"

func main() {
	// Configuration
	cfg, err := config.NewConfig(Version)
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
