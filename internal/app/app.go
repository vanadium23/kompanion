// Package app configures and runs application.
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/vanadium23/kompanion/config"
	"github.com/vanadium23/kompanion/internal/auth"
	"github.com/vanadium23/kompanion/internal/controller/http/opds"
	v1 "github.com/vanadium23/kompanion/internal/controller/http/v1"
	"github.com/vanadium23/kompanion/internal/controller/http/web"
	"github.com/vanadium23/kompanion/internal/controller/http/webdav"
	"github.com/vanadium23/kompanion/internal/library"
	"github.com/vanadium23/kompanion/internal/stats"
	"github.com/vanadium23/kompanion/internal/storage"
	"github.com/vanadium23/kompanion/internal/sync"
	"github.com/vanadium23/kompanion/pkg/httpserver"
	"github.com/vanadium23/kompanion/pkg/logger"
	"github.com/vanadium23/kompanion/pkg/postgres"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	bookStorage, err := storage.NewStorage(cfg.BookStorage.Type, cfg.BookStorage.Path, pg)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - storage.NewStorage: %w", err))
	}
	statsStorage, err := storage.NewStorage(cfg.StatsStorage.Type, cfg.StatsStorage.Path, pg)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - storage.NewStorage: %w", err))
	}

	// Use case
	var repo auth.UserRepo
	switch cfg.Auth.Storage {
	case "memory":
		repo = auth.NewMemoryUserRepo()
	case "postgres":
		repo = auth.NewUserDatabaseRepo(pg)
	default:
		l.Fatal(fmt.Errorf("app - Run - unknown storage: %s", cfg.Auth.Storage))
	}
	authService := auth.InitAuthService(
		repo,
		cfg.Auth.Username,
		cfg.Auth.Password,
	)
	progress := sync.NewProgressSync(sync.NewProgressDatabaseRepo(pg))
	shelf := library.NewBookShelf(bookStorage, library.NewBookDatabaseRepo(pg), l)
	rs := stats.NewKOReaderStats(statsStorage, pg)

	// HTTP Server
	handler := gin.New()
	web.NewRouter(handler, l, authService, progress, shelf, rs, cfg.Version)
	v1.NewRouter(handler, l, authService, progress, shelf)
	opds.NewRouter(handler, l, authService, progress, shelf)
	webdav.NewRouter(handler, authService, l, rs, cfg.StatsStorage.Path)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
