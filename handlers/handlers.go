package handlers

import (
	"encoding/json"
	"net/http"
)

type responseBody struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func CatchAllHandler(w http.ResponseWriter, r *http.Request) {
	responseBody := responseBody{
		Code:    1,
		Message: "Route not found",
		Data:    nil,
	}
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(responseBody)
}
