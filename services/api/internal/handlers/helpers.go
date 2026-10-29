package handlers

import (
	"encoding/json"
	"net/http"

	"monitopro/services/api/internal/domain"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, msg, code string) {
	writeJSON(w, status, domain.ErrorResponse{
		Error: msg,
		Code:  code,
	})
}
