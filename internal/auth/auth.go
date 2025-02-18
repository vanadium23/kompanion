package auth

import (
	"context"
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"net"

	"github.com/moroz/uuidv7-go"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo UserRepo
}

func InitAuthService(repo UserRepo, username, password string) *AuthService {
	auth := &AuthService{repo: repo}
	auth.RegisterUser(context.Background(), username, password)
	return auth
}

func (a *AuthService) RegisterUser(ctx context.Context, username, password string) error {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	newUser := User{
		Username:       username,
		HashedPassword: hashedPassword,
	}
	return a.repo.CreateUser(ctx, newUser)
}

func (a *AuthService) CheckPassword(ctx context.Context, username string, password string) bool {
	user, err := a.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return false
	}
	return comparePasswords(user.HashedPassword, password)
}

func (a *AuthService) Login(ctx context.Context, username string, password string, userAgent string, clientIP net.IP) (string, error) {
	user, err := a.repo.GetUserByUsername(ctx, username)
	if err != nil {
		// we don't want to leak information about user existence
		return "", IncorrectPassword
	}

	if !comparePasswords(user.HashedPassword, password) {
		return "", IncorrectPassword
	}

	sessionKey := uuidv7.Generate().String()
	err = a.repo.StoreSession(ctx, username, sessionKey, userAgent, clientIP)
	if err != nil {
		return "", err
	}
	return sessionKey, nil
}

func (a *AuthService) Logout(ctx context.Context, sessionKey string) error {
	return a.repo.DeleteSession(ctx, sessionKey)
}

func (a *AuthService) IsAuthenticated(ctx context.Context, sessionKey string) bool {
	_, err := a.repo.GetUserBySession(ctx, sessionKey)
	return err == nil
}

func (a *AuthService) AddUserDevice(ctx context.Context, device_name, password string) error {
	hashedPassword := hashSyncPassword(password)

	newDevice := Device{
		Name:           device_name,
		HashedPassword: hashedPassword,
	}
	return a.repo.CreateDevice(ctx, newDevice)
}

func (a *AuthService) DeactivateUserDevice(ctx context.Context, device_name string) error {
	return a.repo.DeleteDevice(ctx, device_name)
}

func (a *AuthService) CheckDevicePassword(ctx context.Context, device_name, password string, plain bool) bool {
	device, err := a.repo.GetDeviceByName(ctx, device_name)
	if err != nil {
		return false
	}
	toCheck := password
	if plain {
		toCheck = hashSyncPassword(password)
	}
	return subtle.ConstantTimeCompare([]byte(device.HashedPassword), []byte(toCheck)) == 1
}

func (a *AuthService) ListDevices(ctx context.Context) ([]Device, error) {
	return a.repo.ListDevices(ctx)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), err
}

func comparePasswords(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

func hashSyncPassword(sync_password string) string {
	// KOReader sync server uses md5 to hash the password
	hash := md5.Sum([]byte(sync_password))
	return hex.EncodeToString(hash[:])
}
