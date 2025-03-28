package rest

import (
	"auth-server-go/internal/middlewares"
	"auth-server-go/internal/models"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Token   string         `json:"token"`
	Account models.Account `json:"account"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func AddEmailAuthRoutes(rest REST, public chi.Router) {
	public.Post("/auth/register", rest.RegisterHandler)
	public.Post("/auth/login", rest.LoginHandler)
}

func (rest *REST) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if !isValidEmail(req.Email) {
		respondWithError(w, http.StatusBadRequest, "Invalid email format")
		return
	}

	if len(req.Password) < 8 {
		respondWithError(w, http.StatusBadRequest, "Password must be at least 8 characters long")
		return
	}

	account := models.Account{}
	existingAccount := rest.DB.Where("email = ?", req.Email).First(&account)

	if err := existingAccount.Error; err == nil {
		respondWithError(w, http.StatusConflict, "Email already in use")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to process registration")
		return
	}

	userId, err := gonanoid.New()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokenString, err := middlewares.GenerateToken(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	account = models.Account{
		UserID:       userId,
		DisplayName:  strings.Split(req.Email, "@")[0],
		Email:        req.Email,
		Provider:     "email",
		CreatedAt:    time.Now().Unix(),
		LastLoggedIn: time.Now().Unix(),
		Password:     string(hashedPassword),
	}

	if err := rest.DB.Create(&account).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, RegisterResponse{
		Token:   tokenString,
		Account: account,
	})
}

func (rest *REST) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if !isValidEmail(req.Email) {
		respondWithError(w, http.StatusBadRequest, "Invalid email format")
		return
	}

	account := models.Account{}
	if err := rest.DB.Where("email = ?", req.Email).First(&account).Error; err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(req.Password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	tokenString, err := middlewares.GenerateToken(account.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	account.LastLoggedIn = time.Now().Unix()
	if err := rest.DB.Save(&account).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, RegisterResponse{
		Token:   tokenString,
		Account: account,
	})
}

func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ErrorResponse{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
