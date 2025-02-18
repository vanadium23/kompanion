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
	mu       *sync.RWMutex
}

func NewMemoryUserRepo() *MemoryRepo {
	return &MemoryRepo{
		sessions: make(map[string]bool),
		devices:  make(map[string]Device),
		mu:       &sync.RWMutex{},
	}
}

func (mr *MemoryRepo) CreateUser(ctx context.Context, user User) error {
	if mr.user.Username != "" {
		return UserAlreadyCreated
	}
	mr.user = user
	return nil
}

func (mr *MemoryRepo) GetUserByUsername(ctx context.Context, username string) (User, error) {
	if mr.user.Username != username {
		return User{}, errors.New("not found")
	}
	return mr.user, nil
}

func (mr *MemoryRepo) GetUserBySession(ctx context.Context, sessionKey string) (User, error) {
	mr.mu.RLock()
	_, ok := mr.sessions[sessionKey]
	mr.mu.RUnlock()

	if !ok {
		return User{}, errors.New("not found")
	}

	return mr.user, nil
}

func (mr *MemoryRepo) StoreSession(
	ctx context.Context,
	username string,
	sessionKey string,
	userAgent string,
	clientIP net.IP,
) error {
	mr.mu.RLock()
	if mr.user.Username != username {
		mr.mu.RUnlock()
		return UserNotFound
	}
	mr.mu.RUnlock()

	mr.mu.Lock()
	mr.sessions[sessionKey] = true
	mr.mu.Unlock()

	return nil
}

func (mr *MemoryRepo) DeleteSession(ctx context.Context, sessionKey string) error {
	mr.mu.Lock()
	delete(mr.sessions, sessionKey)
	mr.mu.Unlock()

	return nil
}

func (mr *MemoryRepo) CreateDevice(ctx context.Context, device Device) error {
	mr.mu.RLock()
	_, ok := mr.devices[device.Name]
	mr.mu.RUnlock()

	if ok {
		return DeviceAlreadyCreated
	}

	mr.mu.Lock()
	mr.devices[device.Name] = device
	mr.mu.Unlock()

	return nil
}

func (mr *MemoryRepo) GetDeviceByName(ctx context.Context, device_name string) (Device, error) {
	mr.mu.RLock()
	device, ok := mr.devices[device_name]
	mr.mu.RUnlock()

	if !ok {
		return Device{}, errors.New("not found")
	}

	return device, nil
}

func (mr *MemoryRepo) DeleteDevice(ctx context.Context, device_name string) error {
	mr.mu.Lock()
	delete(mr.devices, device_name)
	mr.mu.Unlock()

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
