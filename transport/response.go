package transport

import (
	"encoding/json"
	"net/http"
	"personal/wassup/cerror"
)

func handleResponse(ws http.ResponseWriter, result any, err error) {
	if err != nil {
		if cerr, ok := err.(*cerror.AppError); ok {
			http.Error(ws, cerr.Error(), cerr.StatusCode)
			return
		}
		http.Error(ws, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.WriteHeader(http.StatusOK)
	ws.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(ws).Encode(result); err != nil {
		http.Error(ws, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
