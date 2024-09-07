package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/DrC0ns0le/bind-api/render"
)

func GetRendersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	BeforeAfterMap := struct {
		Before map[string]string `json:"before"`
		After  map[string]string `json:"after"`
	}{
		Before: make(map[string]string),
		After:  make(map[string]string),
	}

	BeforeAfterMap.After, err = render.PreviewZoneRender(r.Context())
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

	for f := range BeforeAfterMap.After {
		// find the file in the output directory
		files, err := os.ReadDir("output")
		if err != nil {
			errorMsg := responseBody{
				Code:    2,
				Message: "Unable to find output directory",
				Data:    err.Error(),
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(errorMsg)
			return
		}

		// if parts := strings.Split(f, "."); parts[len(parts)-1] == "com" || parts[len(parts)-1] == "arpa" {
		// 	f = f + ".conf"
		// }

		for _, file := range files {
			if file.Name() == f {
				// store text content of file in beforeAfterMap.before as string
				fBytes, err := (os.ReadFile("output/" + f))
				if err != nil {
					errorMsg := responseBody{
						Code:    3,
						Message: "Unable to read file",
						Data:    err.Error(),
					}
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(errorMsg)
					return
				}
				BeforeAfterMap.Before[f] = string(fBytes)
			}
		}

		if _, ok := BeforeAfterMap.Before[f]; !ok {
			errorMsg := responseBody{
				Code:    4,
				Message: "File not found in output directory",
				Data:    f,
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(errorMsg)
			return
		}
	}

	responseBody := responseBody{
		Code:    0,
		Message: "Zones rendered successfully",
		Data:    BeforeAfterMap,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
}
