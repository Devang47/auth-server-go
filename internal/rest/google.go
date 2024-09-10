package rest

import (
	"auth-server-go/internal/middlewares"
	"auth-server-go/internal/models"
	"context"
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
	public.Get("/auth/{provider}", rest.loginUserHandler)
	public.Get("/auth/{provider}/callback", rest.getAuthCallbackHandler)
	public.Get("/logout/{provider}", rest.logoutHandler)
}

func (rest *REST) getAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(context.Background(), "provider", provider))

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	tokenString, err := middlewares.GenerateToken(user.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	account := models.Account{}

	userId, err := gonanoid.New()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := rest.DB.Where("email = ?", user.Email).First(&account).Error; err != nil {
		log.Printf("Account not found, creating new account")
		account = models.Account{
			UserID:       userId,
			DisplayName:  strings.Trim(user.FirstName+" "+user.LastName, " "),
			Name:         user.Name,
			Email:        user.Email,
			Provider:     provider,
			Picture:      user.AvatarURL,
			CreatedAt:    time.Now().Unix(),
			LastLoggedIn: time.Now().Unix(),
		}

		if err := rest.DB.Create(&account).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else {
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
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(context.Background(), "provider", provider))

	if _, err := gothic.CompleteUserAuth(w, r); err != nil {
		gothic.BeginAuthHandler(w, r)
	}
}

func (rest *REST) logoutHandler(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(context.Background(), "provider", provider))

	gothic.Logout(w, r)
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}
