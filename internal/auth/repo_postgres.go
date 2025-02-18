package auth

import (
	"context"
	"fmt"
	"net"

	"gitea.chrnv.ru/vanadium23/kompanion/pkg/postgres"
)

type UserDatabaseRepo struct {
	*postgres.Postgres
}

func NewUserDatabaseRepo(pg *postgres.Postgres) *UserDatabaseRepo {
	return &UserDatabaseRepo{pg}
}

func (r *UserDatabaseRepo) GetUserByUsername(ctx context.Context, username string) (User, error) {
	sql := `
		SELECT username, hashed_password
		FROM auth_user
		WHERE username = $1
	`
	args := []interface{}{username}

	row := r.Pool.QueryRow(ctx, sql, args...)
	var user User
	err := row.Scan(&user.Username, &user.HashedPassword)
	if err != nil {
		return User{}, fmt.Errorf("UserDatabaseRepo - GetUser - row.Scan: %w", err)
	}

	return user, nil
}

func (r *UserDatabaseRepo) CreateUser(ctx context.Context, user User) error {
	sql := `
		INSERT INTO auth_user (username, hashed_password)
		VALUES ($1, $2)
	`
	args := []interface{}{user.Username, user.HashedPassword}

	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserDatabaseRepo - CreateUser - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *UserDatabaseRepo) StoreSession(
	ctx context.Context,
	username string,
	sessionKey string,
	userAgent string,
	clientIP net.IP,
) error {
	sql := `
		INSERT INTO auth_session (username, session_key, user_agent, ip_address)
		VALUES ($1, $2, $3, $4)
	`
	args := []interface{}{username, sessionKey, userAgent, clientIP}

	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserDatabaseRepo - StoreSession - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *UserDatabaseRepo) GetUserBySession(ctx context.Context, sessionKey string) (User, error) {
	sql := `
		SELECT auth_user.username, auth_user.hashed_password
		FROM auth_user
		JOIN auth_session ON auth_user.username = auth_session.username
		WHERE session_key = $1 AND auth_session.is_active
	`
	args := []interface{}{sessionKey}

	row := r.Pool.QueryRow(ctx, sql, args...)
	var user User
	err := row.Scan(&user.Username, &user.HashedPassword)
	if err != nil {
		return User{}, fmt.Errorf("UserDatabaseRepo - GetUserBySession - row.Scan: %w", err)
	}

	return user, nil
}

func (r *UserDatabaseRepo) DeleteSession(ctx context.Context, sessionKey string) error {
	sql := `UPDATE auth_session SET is_active = false WHERE session_key = $1`
	args := []interface{}{sessionKey}

	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserDatabaseRepo - DeleteSession - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *UserDatabaseRepo) CreateDevice(ctx context.Context, device Device) error {
	sql := `
		INSERT INTO auth_device (device_name, hashed_password)
		VALUES ($1, $2)
	`
	args := []interface{}{device.Name, device.HashedPassword}

	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserDatabaseRepo - CreateDevice - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *UserDatabaseRepo) GetDeviceByName(ctx context.Context, deviceName string) (Device, error) {
	sql := `
		SELECT device_name, hashed_password
		FROM auth_device
		WHERE device_name = $1
	`
	args := []interface{}{deviceName}

	row := r.Pool.QueryRow(ctx, sql, args...)
	var device Device
	err := row.Scan(&device.Name, &device.HashedPassword)
	if err != nil {
		return Device{}, fmt.Errorf("UserDatabaseRepo - GetDeviceByName - row.Scan: %w", err)
	}

	return device, nil
}

func (r *UserDatabaseRepo) DeleteDevice(ctx context.Context, deviceName string) error {
	sql := `
		UPDATE auth_device
		SET is_active = false,
			deactivated_at = NOW()
		WHERE device_name = $1
	`
	args := []interface{}{deviceName}

	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserDatabaseRepo - DeleteDevice - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *UserDatabaseRepo) ListDevices(ctx context.Context) ([]Device, error) {
	sql := `
		SELECT device_name, hashed_password
		FROM auth_device
	`

	rows, err := r.Pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("UserDatabaseRepo - ListDevices - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	var devices []Device
	for rows.Next() {
		var device Device
		err = rows.Scan(&device.Name, &device.HashedPassword)
		if err != nil {
			return nil, fmt.Errorf("UserDatabaseRepo - ListDevices - rows.Scan: %w", err)
		}
		devices = append(devices, device)
	}

	return devices, nil
}
