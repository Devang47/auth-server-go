package rest

import (
	"auth-server-go/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func AddAccountRoutes(rest REST, public, secureRoutes chi.Router) {
	secureRoutes.Get("/account", rest.getAccountHandler)
	public.Get("/account/{id}", rest.getAccountByIdHandler)
}

func (rest *REST) getAccountHandler(w http.ResponseWriter, r *http.Request) {
	accountId := r.Context().Value("accountId").(string)

	if accountId == "" {
		http.Error(w, "Account ID is required", http.StatusBadRequest)
		return
	}

	log.Printf("Getting account with ID %v", accountId)

	account := models.Account{}

	if err := rest.DB.First(&account, "user_id = ?", accountId).Error; err != nil {
		log.Print(err)
		http.Error(w, fmt.Sprintf("account with ID %v not found", accountId), http.StatusNotFound)
		return
	}

	res, _ := json.Marshal(account)
	w.Write(res)
}

func (rest *REST) getAccountByIdHandler(w http.ResponseWriter, r *http.Request) {
	accountId := chi.URLParam(r, "id")
	if accountId == "" {
		http.Error(w, "Account ID is required", http.StatusBadRequest)
		return
	}

	log.Printf("Getting account with ID %v", accountId)

	account := models.Account{}

	if err := rest.DB.First(&account, "user_id = ?", accountId).Error; err != nil {
		log.Print(err)
		http.Error(w, fmt.Sprintf("account with ID %v not found", accountId), http.StatusNotFound)
		return
	}

	res, _ := json.Marshal(account)
	w.Write(res)
}
