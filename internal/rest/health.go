package rest

import (
	"encoding/json"
	"net/http"
)

func (rest *REST) GetHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("OK")
}
