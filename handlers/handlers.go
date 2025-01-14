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

type paginationMetadata struct {
	CurrentPage int  `json:"current_page"`
	PageSize    int  `json:"page_size"`
	TotalPages  int  `json:"total_pages"`
	TotalItems  int  `json:"total_items"`
	HasNextPage bool `json:"has_next_page"`
	HasPrevPage bool `json:"has_prev_page"`
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
