package auth_test

import (
	"testing"
	"time"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/EarlMatthews/chirpy/internal/auth"
	"net/http"
)

func TestHashPassword(t *testing.T) {
	password := "securepassword"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		t.Errorf("Hashed password does not match original password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "securepassword"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err := auth.CheckPasswordHash(password, string(hash)); err != nil {
		t.Errorf("CheckPasswordHash failed: %v", err)
	}
}

func TestMakeJWTAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "supersecretkey"
	expiresIn := time.Minute * 5

	token, err := auth.MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	parsedUserID, err := auth.ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}

	if parsedUserID != userID {
		t.Errorf("Parsed user ID does not match original user ID")
	}
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	invalidToken := "invalid.token.here"
	tokenSecret := "supersecretkey"

	_, err := auth.ValidateJWT(invalidToken, tokenSecret)
	if err == nil {
		t.Errorf("Expected error for invalid token, got none")
	}
}

func TestGetBearerToken(t *testing.T) {
	testCases := []struct {
		name          string
		headers       http.Header
		expectedToken string
		expectError   bool
	}{
		{
			name: "Valid Bearer Token",
			headers: func() http.Header {
				h := make(http.Header)
				h.Add("Authorization", "Bearer abc123")
				return h
			}(),
			expectedToken: "abc123",
			expectError:   false,
		},
		{
			name: "Invalid Bearer Token",
			headers: func() http.Header {
				h := make(http.Header)
				h.Add("Authorization", "Basic abc123")
				return h
			}(),
			expectedToken: "",
			expectError:   true,
		},
		{
			name: "No Authorization Header",
			headers: func() http.Header {
				h := make(http.Header)
				return h
			}(),
			expectedToken: "",
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := auth.GetBearerToken(tc.headers)
			if (err != nil) != tc.expectError {
				t.Errorf("GetBearerToken() error = %v, expectError = %v", err, tc.expectError)
			}
			if token != tc.expectedToken {
				t.Errorf("GetBearerToken() got = %v, want = %v", token, tc.expectedToken)
			}
		})
	}
}
