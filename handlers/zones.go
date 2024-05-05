package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/DrC0ns0le/bind-api/rdb"
	"github.com/google/uuid"
)

type Zone struct {
	ID         int    `json:"-"`
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	ModifiedAt int    `json:"modified_at"`
	DeletedAt  int    `json:"deleted_at"`
	Staging    bool   `json:"staging"`
}

type Zones []Zone

func GetZonesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	zones, err := bd.Zones.Get()

	if err != nil {
		log.Fatal(err)
	}

	var Z Zones

	for _, zone := range zones {
		temp := Zone{
			UUID:       zone.UUID,
			Name:       zone.Name,
			ModifiedAt: zone.ModifiedAt,
			DeletedAt:  zone.DeletedAt,
			Staging:    zone.Staging,
		}
		Z = append(Z, temp)
	}

	response := responseBody{
		Code:    0,
		Message: "Zones successfully retrieved",
		Data:    Z,
	}

	// Convert the slice to JSON
	data, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func GetZoneHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract zone UUID from URL
	zoneUUID := r.PathValue("zone_uuid")

	// Find the zone by UUID
	zone, err := bd.Zones.Select(zoneUUID)
	if err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Zone of UUID" + zoneUUID + " not found",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	temp := Zone{
		UUID:       zone.UUID,
		Name:       zone.Name,
		ModifiedAt: zone.ModifiedAt,
		DeletedAt:  zone.DeletedAt,
		Staging:    zone.Staging,
	}

	response := responseBody{
		Code:    0,
		Message: "Zones successfully retrieved",
		Data:    temp,
	}

	// Convert the slice to JSON
	data, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func CreateZoneHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var requestData struct {
		Name string `json:"name"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate UUID for the zone using UUID version 5
	uuid5 := uuid.NewSHA1(dnsNamespaceUUID, []byte(requestData.Name)).String()

	newZone := rdb.Zone{
		UUID:       uuid5,
		Name:       requestData.Name,
		ModifiedAt: int(time.Now().Unix()),
		Staging:    true,
	}

	// Create the zone
	err = bd.Zones.Create(newZone)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the created zone
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newZone)
}

func DeleteZoneHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract zone UUID from URL
	zoneUUID := r.PathValue("zone_uuid")

	// Find the zone by UUID
	zone, err := bd.Zones.Select(zoneUUID)
	if err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Zone of UUID" + zoneUUID + " not found",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Execute the delete query
	err = bd.Zones.Delete(zoneUUID)
	if err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Failed to delete zone of UUID" + zoneUUID,
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Respond with success message
	responseBody := responseBody{
		Code:    0,
		Message: "Zone deleted successfully",
		Data: Zone{
			UUID:       zone.UUID,
			Name:       zone.Name,
			ModifiedAt: zone.ModifiedAt,
			DeletedAt:  zone.DeletedAt,
			Staging:    zone.Staging,
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
}
