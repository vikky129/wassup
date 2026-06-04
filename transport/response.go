package transport

import (
	"encoding/json"
	"errors"
	"net/http"
	"personal/wassup/cerror"

	"log"
)

func handleResponse(ws http.ResponseWriter, result any, err error) {
	if err != nil {
		var errResp ErrResponse
		var cerr *cerror.AppError
		if errors.As(err, &cerr) {
			errResp = ErrResponse{Error: cerr.Error()}
			ws.Header().Set("Content-Type", "application/json")
			ws.WriteHeader(cerr.StatusCode)
		} else {
			errResp = ErrResponse{Error: err.Error()}
			ws.Header().Set("Content-Type", "application/json")
			ws.WriteHeader(http.StatusInternalServerError)
		}
		if err := json.NewEncoder(ws).Encode(errResp); err != nil {
			log.Println("Failed to encode error response:", err)
		}
		return
	}

	ws.Header().Set("Content-Type", "application/json")
	ws.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(ws).Encode(result); err != nil {
		log.Println("Failed to encode response:", err)

		return
	}
}

type ErrResponse struct {
	Error string `json:"error"`
}
