package api

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func renderAPIError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)

	response := errorResponse{
		Message: message,
		Status:  http.StatusText(code),
	}

	json.NewEncoder(w).Encode(response)
}
