package server

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

func SetupGoogleAuth() {
	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_CALLBACK_URL")

	goth.UseProviders(google.New(googleClientId, googleClientSecret, redirectURL))
}
