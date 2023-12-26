package auth

import (
	"os"
	"testing"
)

func TestTokenCreation(t *testing.T) {
	token, err := GenerateToken(128)
	if err != nil {
		t.Errorf("error generating token: %s", err)
	}

	if len(token) != 256 {
		t.Errorf("token length is not 256, got %d", len(token))
	}

	token, err = GenerateToken(64)
	if err != nil {
		t.Errorf("error generating token: %s", err)
	}

	if len(token) != 128 {
		t.Errorf("token length is not 128, got %d", len(token))
	}

	token, err = GenerateToken(32)
	if err != nil {
		t.Errorf("error generating token: %s", err)
	}

	if len(token) != 64 {
		t.Errorf("token length is not 64, got %d", len(token))
	}
}

func TestHmacComparison(t *testing.T) {
	token1, err := GenerateToken(128)
	if err != nil {
		t.Errorf("error generating token: %s", err)
	}

	secret := os.Getenv("SESSION_SECRET")

	mac1 := makeMac(token1, []byte(secret))
	valid := CheckMac(token1, mac1)
	if !valid {
		t.Errorf("mac1 is not valid")
	}
}
