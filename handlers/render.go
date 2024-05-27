package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/DrC0ns0le/bind-api/render"
)

func GetRendersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	renderMap, err := render.PreviewZoneRender(bd)
	if err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Zone rendering failed",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}
	responseBody := responseBody{
		Code:    0,
		Message: "Zones rendered successfully",
		Data:    renderMap,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
}
