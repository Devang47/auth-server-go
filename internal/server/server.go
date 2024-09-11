package server

import (
	"auth-server-go/internal/rest"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"
)

type Server struct {
	port int
}

func NewServer(db *gorm.DB) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port: port,
	}

	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"*", "http://localhost:5173"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowCredentials(),
		handlers.AllowedHeaders([]string{"*", "Content-Type", "Authorization"}),
	)

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      corsMiddleware(rest.SetupREST(db).Router),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
