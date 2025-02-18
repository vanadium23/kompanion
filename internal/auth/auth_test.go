package auth_test

import (
	"context"
	"testing"

	"gitea.chrnv.ru/vanadium23/kompanion/internal/auth"
)

func TestAuthServiceUserOnInit(t *testing.T) {
	ctx := context.Background()

	memory_repo := auth.NewMemoryUserRepo()
	auth := auth.InitAuthService(memory_repo, "user", "password")

	err := auth.RegisterUser(ctx, "user", "password")
	if err == nil {
		t.Error("AuthService User was not registered on init")
	}
}

func TestAuthServiceUserRegister(t *testing.T) {
	ctx := context.Background()

	memory_repo := auth.NewMemoryUserRepo()
	auth := auth.InitAuthService(memory_repo, "", "")

	err := auth.RegisterUser(ctx, "user", "password")
	if err != nil {
		t.Error("RegisterUser failed")
	}

	err = auth.RegisterUser(ctx, "user", "password")
	if err == nil {
		t.Error("RegisterUser failed")
	}
}

func TestAuthServiceUserLogin(t *testing.T) {
	ctx := context.Background()

	memory_repo := auth.NewMemoryUserRepo()
	auth := auth.InitAuthService(memory_repo, "user", "password")

	sessionKey, err := auth.Login(ctx, "user", "password", "user-agent", nil)
	if err != nil {
		t.Error("Login failed")
	}

	if !auth.IsAuthenticated(ctx, sessionKey) {
		t.Error("IsAuthenticated failed")
	}
}
