package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type (
	// Config -.
	Config struct {
		App
		Auth
		HTTP
		Log
		PG
		BookStorage
		StatsStorage
	}

	// App -.
	App struct {
		Name    string
		Version string
	}

	// Auth -.
	Auth struct {
		Username string
		Password string
		Storage  string
	}

	// HTTP -.
	HTTP struct {
		Port string
	}

	// Log -.
	Log struct {
		Level string
	}

	// PG -.
	PG struct {
		PoolMax int
		URL     string
	}

	BookStorage struct {
		Type string
		Path string
	}

	// StatsStorage -.
	StatsStorage struct {
		Type string
		Path string
	}
)

// NewConfig - reads from env, validates and returns the config.
func NewConfig() (*Config, error) {
	app, err := readAppConfig()
	if err != nil {
		return nil, err
	}

	auth, err := readAuthConfig()
	if err != nil {
		return nil, err
	}

	http, err := readHTTPConfig()
	if err != nil {
		return nil, err
	}

	log, err := readLogConfig()
	if err != nil {
		return nil, err
	}

	postgres, err := readPostgresConfig()
	if err != nil {
		return nil, err
	}

	bookStorage, err := readBookStorageConfig()
	if err != nil {
		return nil, err
	}

	statsStorage, err := readStatsStorageConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		App:          app,
		Auth:         auth,
		HTTP:         http,
		Log:          log,
		PG:           postgres,
		BookStorage:  bookStorage,
		StatsStorage: statsStorage,
	}, nil
}

func readAppConfig() (App, error) {
	return App{
		Name:    "kompanion",
		Version: "0.0.1",
	}, nil
}

func readAuthConfig() (Auth, error) {
	username := readPrefixedEnv("AUTH_USERNAME")
	password := readPrefixedEnv("AUTH_PASSWORD")
	if username == "" || password == "" {
		return Auth{}, fmt.Errorf("username or password is empty")
	}

	storage := readPrefixedEnv("AUTH_STORAGE")
	if storage == "" {
		storage = "postgres"
	}

	return Auth{
		Username: username,
		Password: password,
		Storage:  storage,
	}, nil
}

func readHTTPConfig() (HTTP, error) {
	port := readPrefixedEnv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	return HTTP{
		Port: port,
	}, nil
}

func readLogConfig() (Log, error) {
	level := readPrefixedEnv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	return Log{
		Level: level,
	}, nil
}

func readPostgresConfig() (PG, error) {
	var poolMax int
	poolMaxEnv := readPrefixedEnv("PG_POOL_MAX")

	if poolMaxEnv == "" {
		poolMax = 2
	} else {
		poolMaxEnvInt, err := strconv.Atoi(poolMaxEnv)
		if err != nil {
			return PG{}, fmt.Errorf("pool max is not a number")
		}
		poolMax = poolMaxEnvInt
	}

	url := readPrefixedEnv("PG_URL")
	if url == "" {
		return PG{}, fmt.Errorf("postgres url is empty")
	}

	return PG{
		PoolMax: poolMax,
		URL:     url,
	}, nil
}

func readBookStorageConfig() (BookStorage, error) {
	bstorage_type := readPrefixedEnv("BSTORAGE_TYPE")
	if bstorage_type == "" {
		bstorage_type = "postgres"
	}
	bstorage_path := readPrefixedEnv("BSTORAGE_PATH")
	return BookStorage{
		Type: bstorage_type,
		Path: bstorage_path,
	}, nil
}

func readStatsStorageConfig() (StatsStorage, error) {
	stats_type := readPrefixedEnv("STATS_TYPE")
	if stats_type == "" {
		stats_type = "postgres"
	}
	stats_path := readPrefixedEnv("STATS_PATH")
	return StatsStorage{
		Type: stats_type,
		Path: stats_path,
	}, nil
}

func readPrefixedEnv(key string) string {
	envKey := fmt.Sprintf("KOMPANION_%s", strings.ToUpper(key))
	return os.Getenv(envKey)
}
