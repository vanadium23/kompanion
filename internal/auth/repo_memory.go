package auth

import (
	"context"
	"errors"
	"net"
	"sync"
)

type MemoryRepo struct {
	user     User
	sessions map[string]bool
	devices  map[string]Device
	mu       sync.RWMutex
}

func NewMemoryUserRepo() *MemoryRepo {
	return &MemoryRepo{
		sessions: make(map[string]bool),
		devices:  make(map[string]Device),
	}
}

func (mr *MemoryRepo) CreateUser(ctx context.Context, user User) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if mr.user.Username != "" {
		return UserAlreadyCreated
	}
	mr.user = user
	return nil
}

func (mr *MemoryRepo) GetUserByUsername(ctx context.Context, username string) (User, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	if mr.user.Username != username {
		return User{}, UserNotFound
	}
	return mr.user, nil
}

func (mr *MemoryRepo) GetUserBySession(ctx context.Context, sessionKey string) (User, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	if !mr.sessions[sessionKey] {
		return User{}, SessionNotFound
	}
	return mr.user, nil
}

func (mr *MemoryRepo) StoreSession(ctx context.Context, username, sessionKey, userAgent string, clientIP net.IP) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if mr.user.Username != username {
		return UserNotFound
	}

	mr.sessions[sessionKey] = true
	return nil
}

func (mr *MemoryRepo) DeleteSession(ctx context.Context, sessionKey string) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if !mr.sessions[sessionKey] {
		return SessionNotFound
	}
	delete(mr.sessions, sessionKey)
	return nil
}

func (mr *MemoryRepo) CreateDevice(ctx context.Context, device Device) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if _, ok := mr.devices[device.Name]; ok {
		return DeviceAlreadyCreated
	}
	mr.devices[device.Name] = device
	return nil
}

func (mr *MemoryRepo) GetDeviceByName(ctx context.Context, deviceName string) (Device, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	device, ok := mr.devices[deviceName]
	if !ok {
		return Device{}, errors.New("not found")
	}
	return device, nil
}

func (mr *MemoryRepo) DeleteDevice(ctx context.Context, deviceName string) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if _, ok := mr.devices[deviceName]; !ok {
		return errors.New("not found")
	}
	delete(mr.devices, deviceName)
	return nil
}

func (mr *MemoryRepo) ListDevices(ctx context.Context) ([]Device, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	devices := make([]Device, 0, len(mr.devices))
	for _, device := range mr.devices {
		devices = append(devices, device)
	}
	return devices, nil
}
