package rest

import (
	"auth-server-go/internal/middlewares"
	"auth-server-go/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupEmailTestDB(t *testing.T, testSuffix string) *gorm.DB {
	// Use a unique database name for each test run
	dbName := fmt.Sprintf("file::memory:?cache=shared&_uuid=%d_%s", time.Now().UnixNano(), testSuffix)
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&models.Account{})
	require.NoError(t, err)

	// Create a test account with a hashed password and unique email/user ID
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	testAccount := models.Account{
		UserID:       fmt.Sprintf("existing-user-id-%s", testSuffix),
		DisplayName:  "existinguser",
		Email:        fmt.Sprintf("existing-%s@example.com", testSuffix),
		Provider:     "email",
		Password:     string(hashedPassword),
		CreatedAt:    time.Now().Unix() - 86400,
		LastLoggedIn: time.Now().Unix() - 3600,
	}

	err = db.Create(&testAccount).Error
	require.NoError(t, err)

	return db
}

func TestRegisterHandler(t *testing.T) {
	// Set up a JWT signing key for the tests
	os.Setenv("JWT_SIGNING_KEY", "test-signing-key")
	defer os.Unsetenv("JWT_SIGNING_KEY")

	db := setupEmailTestDB(t, "register")
	rest := REST{DB: db}

	t.Run("Valid registration", func(t *testing.T) {
		// Create a registration request
		registerReq := RegisterRequest{
			Email:    "newuser@example.com",
			Password: "password123",
		}
		reqBody, err := json.Marshal(registerReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		rest.RegisterHandler(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var response RegisterResponse
		err = json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		// Verify token and account
		assert.NotEmpty(t, response.Token)
		assert.Equal(t, "newuser", response.Account.DisplayName)
		assert.Equal(t, "newuser@example.com", response.Account.Email)
		assert.Equal(t, "email", response.Account.Provider)
		assert.NotEmpty(t, response.Account.UserID)

		// Verify token is valid
		token, claims, err := middlewares.ValidateToken(response.Token)
		assert.NoError(t, err)
		assert.True(t, token.Valid)
		assert.Equal(t, response.Account.UserID, claims.AccountID)

	})

	t.Run("Email already in use", func(t *testing.T) {
		// Create a registration request with existing email
		registerReq := RegisterRequest{
			Email:    "existing-register@example.com",
			Password: "password123",
		}
		reqBody, err := json.Marshal(registerReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		rest.RegisterHandler(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)

		var response ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "Email already in use", response.Error)

	})

	t.Run("Invalid email format", func(t *testing.T) {
		registerReq := RegisterRequest{
			Email:    "invalidemail",
			Password: "password123",
		}
		reqBody, err := json.Marshal(registerReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		rest.RegisterHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "Invalid email format", response.Error)
	})

	t.Run("Password too short", func(t *testing.T) {
		registerReq := RegisterRequest{
			Email:    "newuser@example.com",
			Password: "short",
		}
		reqBody, err := json.Marshal(registerReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		rest.RegisterHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "Password must be at least 8 characters long", response.Error)
	})

	t.Run("Invalid JSON payload", func(t *testing.T) {
		invalidJSON := []byte(`{"email": "invalid-json`)

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		rest.RegisterHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response ErrorResponse
		err := json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "Invalid request payload", response.Error)
	})

	t.Run("Method not allowed", func(t *testing.T) {
		// Create GET request (only POST is allowed)
		req := httptest.NewRequest(http.MethodGet, "/auth/register", nil)
		rr := httptest.NewRecorder()

		rest.RegisterHandler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestLoginHandler(t *testing.T) {
	// Set up a JWT signing key for the tests
	os.Setenv("JWT_SIGNING_KEY", "test-signing-key")
	defer os.Unsetenv("JWT_SIGNING_KEY")

	db := setupEmailTestDB(t, "login")
	rest := REST{DB: db}

	t.Run("Valid login", func(t *testing.T) {
		// Create a login request with valid credentials
		loginReq := RegisterRequest{
			Email:    "existing-login@example.com",
			Password: "password123",
		}
		reqBody, err := json.Marshal(loginReq)
		require.NoError(t, err)

		// Get account before login to check last login time
		var accountBefore models.Account
		err = db.Where("email = ?", "existing-login@example.com").First(&accountBefore).Error
		require.NoError(t, err)
		lastLoginBefore := accountBefore.LastLoggedIn

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		rest.LoginHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response RegisterResponse
		err = json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		// Verify token and account
		assert.NotEmpty(t, response.Token)
		assert.Equal(t, "existinguser", response.Account.DisplayName)
		assert.Equal(t, "existing-login@example.com", response.Account.Email)
		assert.Equal(t, "email", response.Account.Provider)
		assert.Equal(t, "existing-user-id-login", response.Account.UserID)

		// Verify last login time was updated
		assert.True(t, response.Account.LastLoggedIn > lastLoginBefore)

		// Verify token is valid
		token, claims, err := middlewares.ValidateToken(response.Token)
		assert.NoError(t, err)
		assert.True(t, token.Valid)
		assert.Equal(t, "existing-user-id-login", claims.AccountID)
	})

	t.Run("Non-existent email", func(t *testing.T) {
		// Create a login request with non-existent email
		loginReq := RegisterRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}
		reqBody, err := json.Marshal(loginReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		rest.LoginHandler(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "Invalid email or password", response.Error)
	})

	t.Run("Incorrect password", func(t *testing.T) {
		// Create a login request with wrong password
		loginReq := RegisterRequest{
			Email:    "existing-login@example.com",
			Password: "wrongpassword",
		}
		reqBody, err := json.Marshal(loginReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		rest.LoginHandler(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "Invalid email or password", response.Error)
	})

	t.Run("Invalid email format", func(t *testing.T) {
		// Create a login request with invalid email format
		loginReq := RegisterRequest{
			Email:    "invalidemail",
			Password: "password123",
		}
		reqBody, err := json.Marshal(loginReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		rest.LoginHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response ErrorResponse
		err = json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "Invalid email format", response.Error)
	})

	t.Run("Method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
		rr := httptest.NewRecorder()

		rest.LoginHandler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestIsValidEmail(t *testing.T) {
	testCases := []struct {
		email    string
		expected bool
		desc     string
	}{
		{"test@example.com", true, "Simple email"},
		{"test.user@example.com", true, "Email with dot in local part"},
		{"test+user@example.com", true, "Email with plus in local part"},
		{"test-user@example.co.uk", true, "Email with hyphen in domain"},
		{"test@localhost", false, "Missing TLD"},
		{"test@.com", false, "Missing domain"},
		{"@example.com", false, "Missing local part"},
		{"test@", false, "Missing domain and TLD"},
		{"test", false, "No @ symbol"},
		{"", false, "Empty string"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := isValidEmail(tc.email)
			assert.Equal(t, tc.expected, result, "Email: %s", tc.email)
		})
	}
}
