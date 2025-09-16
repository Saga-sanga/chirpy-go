package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestCheckPasswordHash(t *testing.T) {
	password := "hellokitty123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Errorf("Error hashing: %s", err)
	}

	if err := CheckPasswordHash(password, string(hash)); err != nil {
		t.Errorf("Hash failed: %s", err)
	}
}
