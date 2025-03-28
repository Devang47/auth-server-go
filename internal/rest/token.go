package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"auth-server-go/internal/middlewares"

	"github.com/golang-jwt/jwt/v5"
)

type newTokenResponse struct {
	Token string
}

func (rest *REST) RefreshToken(w http.ResponseWriter, r *http.Request) {
	authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
	if len(authHeader) != 2 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenString := authHeader[1]

	token, claims, err := middlewares.ValidateToken(tokenString)

	if err != nil || !token.Valid {
		if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			tokenString, _ = middlewares.GenerateToken(claims.AccountID)
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	res, err := json.Marshal(newTokenResponse{Token: tokenString})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
}
