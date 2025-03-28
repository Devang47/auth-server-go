package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"auth-server-go/internal/middlewares"
	"auth-server-go/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T, userIDSuffix string) *gorm.DB {
	// Use in-memory SQLite with a unique identifier to prevent sharing between tests
	dbName := fmt.Sprintf("file::memory:?cache=shared&_uuid=%d", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&models.Account{})
	require.NoError(t, err)

	// Create a test account with a unique ID based on the suffix
	testAccount := models.Account{
		UserID:      fmt.Sprintf("test-user-id-%s", userIDSuffix),
		DisplayName: "testuser",
		Email:       fmt.Sprintf("test-%s@example.com", userIDSuffix),
	}
	err = db.Create(&testAccount).Error
	require.NoError(t, err)

	return db
}

func TestGetAccountHandler(t *testing.T) {
	// Create a unique database for this test function
	db := setupTestDB(t, "get-handler")
	rest := REST{DB: db}

	t.Run("Success - Account found", func(t *testing.T) {
		// Create request with context containing account ID
		req, err := http.NewRequest("GET", "/account", nil)
		require.NoError(t, err)

		accountID := "test-user-id-get-handler"

		// Set account ID in context
		ctx := context.WithValue(req.Context(), middlewares.AccountIDKey, accountID)
		req = req.WithContext(ctx)

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call the handler
		rest.getAccountHandler(rr, req)

		// Check status code
		assert.Equal(t, http.StatusOK, rr.Code)

		// Parse response
		var account models.Account
		err = json.Unmarshal(rr.Body.Bytes(), &account)
		assert.NoError(t, err)

		// Check account details
		assert.Equal(t, accountID, account.UserID)
		assert.Equal(t, "testuser", account.DisplayName)
		assert.Equal(t, "test-get-handler@example.com", account.Email)

	})

	t.Run("Failure - Account not found", func(t *testing.T) {
		// Create request with context containing non-existent account ID
		req, err := http.NewRequest("GET", "/account", nil)
		require.NoError(t, err)

		// Set non-existent account ID in context
		ctx := context.WithValue(req.Context(), middlewares.AccountIDKey, "non-existent-id")
		req = req.WithContext(ctx)

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call the handler
		rest.getAccountHandler(rr, req)

		// Check status code
		assert.Equal(t, http.StatusNotFound, rr.Code)

		// Check error message
		assert.Contains(t, rr.Body.String(), "not found")

	})

	t.Run("Failure - Empty account ID", func(t *testing.T) {
		// Create request with context containing empty account ID
		req, err := http.NewRequest("GET", "/account", nil)
		require.NoError(t, err)

		// Set empty account ID in context
		ctx := context.WithValue(req.Context(), middlewares.AccountIDKey, "")
		req = req.WithContext(ctx)

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call the handler
		rest.getAccountHandler(rr, req)

		// Check status code
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		// Check error message
		assert.Contains(t, rr.Body.String(), "Account ID is required")

	})
}

func TestGetAccountByIdHandler(t *testing.T) {
	// Create a unique database for this test function
	db := setupTestDB(t, "get-by-id")
	rest := REST{DB: db}

	accountID := "test-user-id-get-by-id"

	t.Run("Success - Account found by ID", func(t *testing.T) {
		// Create a new router to test URL params
		r := chi.NewRouter()
		r.Get("/account/{id}", rest.getAccountByIdHandler)

		// Create test server
		ts := httptest.NewServer(r)
		defer ts.Close()

		// Make request to the test server
		resp, err := http.Get(ts.URL + "/account/" + accountID)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check status code
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var account models.Account
		err = json.NewDecoder(resp.Body).Decode(&account)
		assert.NoError(t, err)

		// Check account details
		assert.Equal(t, accountID, account.UserID)
		assert.Equal(t, "testuser", account.DisplayName)
		assert.Equal(t, "test-get-by-id@example.com", account.Email)

	})

	t.Run("Failure - Account not found by ID", func(t *testing.T) {
		// Create a new router to test URL params
		r := chi.NewRouter()
		r.Get("/account/{id}", rest.getAccountByIdHandler)

		// Create test server
		ts := httptest.NewServer(r)
		defer ts.Close()

		// Make request to the test server with non-existent ID
		resp, err := http.Get(ts.URL + "/account/non-existent-id")
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check status code
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	})

	t.Run("Failure - Empty ID parameter", func(t *testing.T) {
		// For this test, we'll mock the request directly since Chi doesn't route to empty URL params
		req := httptest.NewRequest("GET", "/account/", nil)
		rr := httptest.NewRecorder()

		// Create a Chi context with empty ID
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

		// Call handler directly
		rest.getAccountByIdHandler(rr, req)

		// Check status code
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		// Check error message
		assert.Contains(t, rr.Body.String(), "Account ID is required")

	})
}

func TestAddAccountRoutes(t *testing.T) {
	// Create a unique database for this test function
	db := setupTestDB(t, "routes")
	rest := REST{DB: db}

	// Create public and secure routers
	publicRouter := chi.NewRouter()
	secureRouter := chi.NewRouter()

	// Add routes
	AddAccountRoutes(rest, publicRouter, secureRouter)

	// Check that routes are registered (simple check - not comprehensive)
	t.Run("Routes are registered", func(t *testing.T) {
		// This is a simplified check - we're just making sure the function runs without errors
		// A more comprehensive test would involve checking the actual route patterns
	})

}
