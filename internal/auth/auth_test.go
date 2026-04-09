package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	userId := uuid.New()
	secret := "test-secret"

	token, err := MakeJWT(userId, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned unexpected error: %v", err)
	}

	got, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT returned unexpected error: %v", err)
	}

	if got != userId {
		t.Errorf("got user ID %v, want %v", got, userId)
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	userId := uuid.New()

	token, err := MakeJWT(userId, "correct-secret", time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned unexpected error: %v", err)
	}

	_, err = ValidateJWT(token, "wrong-secret")
	if err == nil {
		t.Error("expected error for wrong secret, got nil")
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	userId := uuid.New()

	token, err := MakeJWT(userId, "secret", -time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned unexpected error: %v", err)
	}

	_, err = ValidateJWT(token, "secret")
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}
