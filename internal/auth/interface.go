package auth

import (
	"context"
	"errors"
	"net"
)

type User struct {
	Username       string
	HashedPassword string
}

type Device struct {
	Name           string
	HashedPassword string
}

// TODO: move session key to separate type
// type SessionKey string

type AuthInterface interface {
	CheckPassword(ctx context.Context, username string, password string) bool
	Login(ctx context.Context, username string, password string, userAgent string, clientIP net.IP) (string, error)
	IsAuthenticated(ctx context.Context, sessionKey string) bool
	Logout(ctx context.Context, sessionKey string) error
	RegisterUser(ctx context.Context, username, password string) error

	AddUserDevice(ctx context.Context, device_name, password string) error
	DeactivateUserDevice(ctx context.Context, device_name string) error
	CheckDevicePassword(ctx context.Context, device_name, password string, plain bool) bool
	ListDevices(ctx context.Context) ([]Device, error)
}

var ErrAuth = errors.New("auth error")
var IncorrectPassword = errors.New("incorrect password")

type UserRepo interface {
	CreateUser(ctx context.Context, user User) error
	GetUserByUsername(ctx context.Context, username string) (User, error)
	GetUserBySession(ctx context.Context, sessionKey string) (User, error)

	StoreSession(ctx context.Context, username string, sessionKey string, userAgent string, clientIP net.IP) error
	DeleteSession(ctx context.Context, sessionKey string) error

	CreateDevice(ctx context.Context, device Device) error
	GetDeviceByName(ctx context.Context, device_name string) (Device, error)
	DeleteDevice(ctx context.Context, device_name string) error
	ListDevices(ctx context.Context) ([]Device, error)
}

var UserAlreadyCreated = errors.New("user already created")
var UserNotFound = errors.New("user not found")
var SessionNotFound = errors.New("session not found")
var DeviceAlreadyCreated = errors.New("device already created")
var DeviceNotFound = errors.New("device not found")
