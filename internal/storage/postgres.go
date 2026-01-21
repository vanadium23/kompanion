package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/vanadium23/kompanion/pkg/postgres"
	"github.com/vanadium23/kompanion/pkg/utils"
)

type PostgresStorage struct {
	*postgres.Postgres
}

func NewPostgresStorage(pg *postgres.Postgres) *PostgresStorage {
	return &PostgresStorage{pg}
}

func (ps *PostgresStorage) Write(
	ctx context.Context,
	source string,
	filepath string,
) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	md5Hash, err := utils.PartialMD5(source)
	if err != nil {
		return err
	}

	sql := `
		INSERT INTO storage_blob (file_path, koreader_partial_md5, file_data)
		VALUES ($1, $2, $3)
	`
	args := []interface{}{filepath, md5Hash, data}

	_, err = ps.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("PostgresStorage - Write - r.Pool.Exec: %w", err)
	}

	return nil
}

func (ps *PostgresStorage) Read(ctx context.Context, filepath string) (*os.File, error) {
	sql := `
		SELECT file_data
		FROM storage_blob
		WHERE file_path = $1
	`
	args := []interface{}{filepath}

	var data []byte
	err := ps.Pool.QueryRow(ctx, sql, args...).Scan(&data)
	if err != nil {
		return nil, fmt.Errorf("PostgresStorage - Read - r.Pool.QueryRow: %w", err)
	}

	// make by temp files
	tempFile, err := os.CreateTemp("", "")
	if err != nil {
		return nil, err
	}
	defer tempFile.Close()
	_, err = tempFile.Write(data)
	if err != nil {
		return nil, err
	}
	return tempFile, nil
}

func (ps *PostgresStorage) Delete(ctx context.Context, filepath string) error {
	sql := `
		DELETE FROM storage_blob
		WHERE file_path = $1
	`
	args := []interface{}{filepath}
	_, err := ps.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("PostgresStorage - Read - r.Pool.QueryRow: %w", err)
	}

	return nil
}