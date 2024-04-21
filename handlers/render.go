package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/DrC0ns0le/bind-api/render"
)

func RenderZonesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	zones, err := render.CreateZones(bd)
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(zones)
}
