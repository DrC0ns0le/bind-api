package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/DrC0ns0le/bind-api/commit"
)

func CommitStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Commit status
	status, err := commit.Staging()
	if err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Unable to check commit status",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	responseBody := responseBody{
		Code:    0,
		Message: "Commit status",
		Data: struct {
			Staging bool `json:"staging"`
		}{
			Staging: status,
		},
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
}

func CommitHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Commit changes
	err := commit.Push()
	if err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Unable to commit changes",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	w.WriteHeader(http.StatusOK)
}
