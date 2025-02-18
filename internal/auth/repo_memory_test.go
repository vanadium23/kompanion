package auth

import (
	"context"
	"testing"
)

func TestGetUserByUsername(t *testing.T) {
	username := "test"
	user, memoryRepo := testInitMemoryRepo(username)
	ctx := context.Background()

	foundUser, err := memoryRepo.GetUserByUsername(ctx, username)
	if err != nil {
		t.Fatalf("Received unexpected error %v", err)
	}
	if user != foundUser {
		t.Fatal("user are not equal")
	}
}

func TestUserAlreadyCreated(t *testing.T) {
	username := "test"
	_, memoryRepo := testInitMemoryRepo(username)
	ctx := context.Background()

	err := memoryRepo.CreateUser(ctx, User{
		Username:       username,
		HashedPassword: "test",
	})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestUserNotFound(t *testing.T) {
	username := "test"
	_, memoryRepo := testInitMemoryRepo(username)
	ctx := context.Background()

	_, err := memoryRepo.GetUserByUsername(ctx, "notfound")
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestStoreSession(t *testing.T) {
	username := "test"
	user, memoryRepo := testInitMemoryRepo(username)
	ctx := context.Background()

	session := "test"
	err := memoryRepo.StoreSession(ctx, username, session, "test", nil)
	if err != nil {
		t.Fatalf("Received unexpected error %v", err)
	}

	foundUser, err := memoryRepo.GetUserBySession(ctx, session)
	if err != nil {
		t.Fatalf("Received unexpected error %v", err)
	}
	if user != foundUser {
		t.Fatal("user are not equal")
	}
}

func TestSessionNotFound(t *testing.T) {
	username := "test"
	_, memoryRepo := testInitMemoryRepo(username)
	ctx := context.Background()

	_, err := memoryRepo.GetUserBySession(ctx, "notfound")
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestSessionNotFoundUser(t *testing.T) {
	username := "test"
	_, memoryRepo := testInitMemoryRepo(username)
	ctx := context.Background()

	session := "test"
	err := memoryRepo.StoreSession(ctx, "notfound", session, "test", nil)
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestDeleteSession(t *testing.T) {
	username := "test"
	_, memoryRepo := testInitMemoryRepo(username)
	ctx := context.Background()

	session := "test"
	err := memoryRepo.StoreSession(ctx, username, session, "test", nil)
	if err != nil {
		t.Fatalf("Received unexpected error %v", err)
	}
	err = memoryRepo.DeleteSession(ctx, session)
	if err != nil {
		t.Fatalf("Received unexpected error %v", err)
	}
	_, err = memoryRepo.GetUserBySession(ctx, session)
	if err == nil {
		t.Fatal("Expected error")
	}
}

func testInitMemoryRepo(username string) (User, *MemoryRepo) {
	password := "test"

	memoryRepo := NewMemoryUserRepo()

	ctx := context.Background()
	user := User{
		Username:       username,
		HashedPassword: password,
	}
	memoryRepo.CreateUser(ctx, user)
	return user, memoryRepo
}
