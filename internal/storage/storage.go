package storage

import (
	"errors"

	"gitea.chrnv.ru/vanadium23/kompanion/pkg/postgres"
)

func NewStorage(storage_type, dir string, pg *postgres.Postgres) (Storage, error) {
	switch storage_type {
	case "memory":
		return NewMemoryStorage(), nil
	case "filesystem":
		st, err := NewFilesystemStorage(dir)
		return st, err
	case "postgres":
		st := NewPostgresStorage(pg)
		return st, nil
	}
	return nil, errors.New("unknown storage type")
}
