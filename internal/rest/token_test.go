package rest

import (
	"auth-server-go/internal/middlewares"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestRefreshToken(t *testing.T) {
	// Set JWT signing key for tests
	os.Setenv("JWT_SIGNING_KEY", "test-signing-key")
	defer os.Unsetenv("JWT_SIGNING_KEY")

	rest := &REST{}

	t.Run("No authorization header", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/refresh-token", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(rest.RefreshToken)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Equal(t, "Unauthorized\n", rr.Body.String())
	})

	t.Run("Invalid authorization format", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/refresh-token", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "InvalidFormat")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(rest.RefreshToken)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Equal(t, "Unauthorized\n", rr.Body.String())
	})

	t.Run("Invalid token", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/refresh-token", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer invalid-token")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(rest.RefreshToken)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Valid token returns same token", func(t *testing.T) {
		// Generate a valid token
		tokenString, err := middlewares.GenerateToken("test-user")
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", "/refresh-token", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(rest.RefreshToken)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response newTokenResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, tokenString, response.Token)
	})

	t.Run("Expired token gets refreshed", func(t *testing.T) {
		// Create expired token manually
		claims := middlewares.Claims{
			AccountID: "expired-user",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)), // Expired
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-48 * time.Hour)),
				Issuer:    "auth-server-go",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		expiredToken, err := token.SignedString([]byte(os.Getenv("JWT_SIGNING_KEY")))
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", "/refresh-token", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+expiredToken)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(rest.RefreshToken)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response newTokenResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Verify we got a new token
		assert.NotEqual(t, expiredToken, response.Token)

		// Verify the new token is valid
		token, newClaims, err := middlewares.ValidateToken(response.Token)
		assert.NoError(t, err)
		assert.True(t, token.Valid)
		assert.Equal(t, "expired-user", newClaims.AccountID)
	})

	t.Run("Malformed token returns error", func(t *testing.T) {
		// Create a token with different signing method to cause error
		token := jwt.New(jwt.SigningMethodNone)
		invalidToken, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

		req, err := http.NewRequest("POST", "/refresh-token", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+invalidToken)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(rest.RefreshToken)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

}
