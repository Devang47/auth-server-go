package middlewares

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	// Set a known JWT signing key for tests
	os.Setenv("JWT_SIGNING_KEY", "test-signing-key")
	defer os.Unsetenv("JWT_SIGNING_KEY")

	t.Run("Generates valid token", func(t *testing.T) {
		accountID := "test-account-123"

		// Generate token
		tokenString, err := GenerateToken(accountID)
		require.NoError(t, err, "Token generation should not error")
		require.NotEmpty(t, tokenString, "Token string should not be empty")

		// Validate the generated token
		token, claims, err := ValidateToken(tokenString)
		require.NoError(t, err, "Should validate the token we just created")
		require.True(t, token.Valid, "Token should be valid")

		// Check claims
		assert.Equal(t, accountID, claims.AccountID, "AccountID in claims should match")
		assert.Equal(t, "auth-server-go", claims.Issuer, "Issuer should be set correctly")

		// Check expiration time
		assert.True(t, time.Until(claims.ExpiresAt.Time) > 23*time.Hour,
			"Expiration should be about 24 hours in the future")
	})
}

func TestValidateToken(t *testing.T) {
	// Set a known JWT signing key for tests
	os.Setenv("JWT_SIGNING_KEY", "test-signing-key")
	defer os.Unsetenv("JWT_SIGNING_KEY")

	t.Run("Valid token", func(t *testing.T) {
		accountID := "test-account-123"

		// Generate a valid token
		tokenString, err := GenerateToken(accountID)
		require.NoError(t, err)

		// Validate the token
		token, claims, err := ValidateToken(tokenString)
		require.NoError(t, err, "Should validate without error")
		assert.True(t, token.Valid, "Token should be valid")
		assert.Equal(t, accountID, claims.AccountID, "AccountID should match")

	})

	t.Run("Invalid token", func(t *testing.T) {
		// Try to validate an invalid token
		token, claims, err := ValidateToken("invalid-token")
		assert.Error(t, err, "Should return an error for invalid token")
		assert.Nil(t, token, "Token should be nil")
		assert.NotNil(t, claims, "Claims should be initialized but empty")
		assert.Empty(t, claims.AccountID, "AccountID should be empty")

	})

	t.Run("Expired token", func(t *testing.T) {
		// Create an expired token
		claims := Claims{
			AccountID: "expired-account",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				Issuer:    "auth-server-go",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SIGNING_KEY")))
		require.NoError(t, err)

		// Validate the expired token
		validatedToken, validatedClaims, err := ValidateToken(tokenString)
		assert.Error(t, err, "Should return an error for expired token")
		assert.ErrorIs(t, err, jwt.ErrTokenExpired, "Error should be token expired")
		assert.False(t, validatedToken.Valid, "Token should be invalid")
		assert.Equal(t, "expired-account", validatedClaims.AccountID, "Should still extract AccountID")

	})

	t.Run("Wrong signing method", func(t *testing.T) {
		// Create a token with a different signing method
		token := jwt.New(jwt.SigningMethodNone)
		token.Claims = &Claims{
			AccountID: "test-account",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "auth-server-go",
			},
		}

		tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

		// Try to validate
		validatedToken, _, err := ValidateToken(tokenString)
		assert.Error(t, err, "Should return an error for wrong signing method")
		assert.False(t, validatedToken.Valid, "Token should be invalid")
	})
}

func TestProtectHandler(t *testing.T) {
	// Set a known JWT signing key for tests
	os.Setenv("JWT_SIGNING_KEY", "test-signing-key")
	defer os.Unsetenv("JWT_SIGNING_KEY")

	// Create a simple test handler that returns the account ID
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := r.Context().Value(AccountIDKey).(string)
		if !ok {
			http.Error(w, "No account ID in context", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(accountID))
	})

	// Apply the middleware
	protectedHandler := ProtectHandler(testHandler)

	t.Run("Valid token passes through", func(t *testing.T) {
		// Generate a valid token
		accountID := "test-account-456"
		tokenString, err := GenerateToken(accountID)
		require.NoError(t, err)

		// Create a request with the token
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		res := httptest.NewRecorder()

		protectedHandler.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "Should return 200 OK")
		assert.Equal(t, accountID, res.Body.String(), "Should return the account ID")

	})

	t.Run("Missing authorization header", func(t *testing.T) {
		// Create a request without Authorization header
		req := httptest.NewRequest("GET", "/protected", nil)
		res := httptest.NewRecorder()

		protectedHandler.ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code, "Should return 401 Unauthorized")
		assert.Contains(t, res.Body.String(), "Unauthorized", "Should return unauthorized message")

	})

	t.Run("Invalid authorization format", func(t *testing.T) {
		// Create a request with invalid Authorization format
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		res := httptest.NewRecorder()

		protectedHandler.ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code, "Should return 401 Unauthorized")
	})

	t.Run("Invalid token", func(t *testing.T) {
		// Create a request with invalid token
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		res := httptest.NewRecorder()

		protectedHandler.ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code, "Should return 401 Unauthorized")
	})

	t.Run("Expired token", func(t *testing.T) {
		// Create an expired token
		claims := Claims{
			AccountID: "expired-account",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				Issuer:    "auth-server-go",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SIGNING_KEY")))
		require.NoError(t, err)

		// Create a request with the expired token
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		res := httptest.NewRecorder()

		protectedHandler.ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code, "Should return 401 Unauthorized")
		assert.Contains(t, res.Body.String(), "Token expired", "Should mention token expired")
	})
}
