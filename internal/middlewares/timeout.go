package middlewares

import (
	"net/http"
	"time"
)

// TimeoutMiddleware creates a new HTTP handler that runs h with the specified timeout
func TimeoutMiddleware(h http.Handler, timeout time.Duration) http.Handler {
	return http.TimeoutHandler(h, timeout, "Request timed out")
}
