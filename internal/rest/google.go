package rest

import (
	"auth-server-go/internal/middlewares"
	"auth-server-go/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth/gothic"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

func AddGoogleAuthRoutes(rest REST, public chi.Router) {
	public.Get("/auth/google", rest.loginUserHandler)
	public.Get("/auth/google/callback", rest.getAuthCallbackHandler)
	public.Get("/logout/google", rest.logoutHandler)
}

func (rest *REST) getAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	account := models.Account{}
	tokenString := ""

	existingAccount := rest.DB.Where("email = ?", user.Email).First(&account)

	if err := existingAccount.Error; err != nil {
		log.Printf("Account not found, creating new account")

		userId, err := gonanoid.New()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tokenString, err = middlewares.GenerateToken(userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		account = models.Account{
			UserID:       userId,
			DisplayName:  strings.Trim(user.FirstName+" "+user.LastName, " "),
			Name:         user.Name,
			Email:        user.Email,
			Provider:     "google",
			Picture:      user.AvatarURL,
			CreatedAt:    time.Now().Unix(),
			LastLoggedIn: time.Now().Unix(),
		}

		if err := rest.DB.Create(&account).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else {
		log.Printf("Account found, updating account")
		tokenString, err = middlewares.GenerateToken(account.UserID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		account.LastLoggedIn = time.Now().Unix()
		if err := rest.DB.Save(&account).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	type response struct {
		Token   string         `json:"token"`
		Account models.Account `json:"account"`
	}

	resp, err := json.Marshal(response{tokenString, account})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(resp)
}

func (rest *REST) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := gothic.CompleteUserAuth(w, r); err != nil {
		gothic.BeginAuthHandler(w, r)
	}
}

func (rest *REST) logoutHandler(w http.ResponseWriter, r *http.Request) {
	gothic.Logout(w, r)
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}
