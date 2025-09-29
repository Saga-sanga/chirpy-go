package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	password1 := "correctPassword123!"
	password2 := "anotherPassword1234@"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password1,
			hash:     hash1,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "WrongPassword",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Password doesn't match different hash",
			password: password1,
			hash:     hash2,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password1,
			hash:     "invalidhash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	secret := "Thisisasecret"
	userID := uuid.New()
	jwt1, _ := MakeJWT(userID, secret, time.Hour)
	jwt2, _ := MakeJWT(userID, secret, 5*time.Second)

	t.Run("Create and Validate token", func(t *testing.T) {
		actualID, _ := ValidateJWT(jwt1, secret)
		if userID != actualID {
			t.Errorf("Want: %q, Got: %q", userID, actualID)
		}
	})

	t.Run("Reject wrong secret", func(t *testing.T) {
		actualID, err := ValidateJWT(jwt2, "wrongsecret")
		if userID == actualID {
			t.Errorf("Want: %q, Got: %q, err: %q", userID, actualID, err)
		}
	})

	t.Run("Reject expired token", func(t *testing.T) {
		time.Sleep(5 * time.Second)
		actualID, err := ValidateJWT(jwt2, secret)
		t.Logf("Got: %s, Error: %s", actualID, err)
		if err == nil {
			t.Errorf("Want Error, Got Error: %q", err)
		}

	})
}

func TestGetBearerToken(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	expected := "teststring"
	request.Header.Set("Authorization", "Bearer "+expected)
	actual, _ := GetBearerToken(request.Header)

	if actual != expected {
		t.Errorf("Want %q, Got %q", expected, actual)
	}
}
