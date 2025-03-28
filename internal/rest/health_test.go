package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetHealth(t *testing.T) {
	rest := &REST{}

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()

	rest.GetHealth(rr, req)

	// Check the status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}

	// Check the response body - trim whitespace and quotes for comparison
	expected := "OK"
	actual := strings.Trim(rr.Body.String(), "\" \n\r\t")
	if actual != expected {
		t.Errorf("Expected response body %q, got %q", expected, actual)
	}
}
