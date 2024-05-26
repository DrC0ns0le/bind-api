package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/DrC0ns0le/bind-api/rdb"
	"github.com/google/uuid"
)

type Zone struct {
	ID         uint32 `json:"-"`
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	CreatedAt  uint64 `json:"created_at"`
	ModifiedAt uint64 `json:"modified_at"`
	DeletedAt  uint64 `json:"deleted_at"`
	Staging    bool   `json:"staging"`
	SOA        SOA    `json:"soa"`
}

type SOA struct {
	PrimaryNS  string `json:"primary_ns"`
	AdminEmail string `json:"admin_email"`
	Refresh    uint16 `json:"refresh"`
	Retry      uint16 `json:"retry"`
	Expire     uint32 `json:"expire"`
	Minimum    uint16 `json:"minimum"`
	TTL        uint16 `json:"ttl"`
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

	response := responseBody{
		Code:    0,
		Message: "Zones successfully retrieved",
		Data: Zone{
			UUID:       zone.UUID,
			Name:       zone.Name,
			CreatedAt:  zone.CreatedAt,
			ModifiedAt: zone.ModifiedAt,
			DeletedAt:  zone.DeletedAt,
			Staging:    zone.Staging,
			SOA: SOA{
				PrimaryNS:  zone.PrimaryNS,
				AdminEmail: zone.AdminEmail,
				Refresh:    zone.Refresh,
				Retry:      zone.Retry,
				Expire:     zone.Expire,
				Minimum:    zone.Minimum,
				TTL:        zone.TTL,
			},
		},
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
		SOA  SOA    `json:"soa"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Invalid request body",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Check name is not empty
	if requestData.Name == "" {
		errorMsg := responseBody{
			Code:    2,
			Message: "Name cannot be empty",
			Data:    nil,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Generate UUID for the zone using UUID version 5
	uuid5 := uuid.NewSHA1(dnsNamespaceUUID, []byte(requestData.Name)).String()

	// Check if the zone already exists
	_, err = bd.Zones.Select(uuid5)
	if err == nil {
		errorMsg := responseBody{
			Code:    2,
			Message: "Zone already exists",
			Data:    nil,
		}
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	if err != sql.ErrNoRows {
		errorMsg := responseBody{
			Code:    3,
			Message: "Error checking if zone exists",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	newZone := rdb.Zone{
		UUID:       uuid5,
		Name:       requestData.Name,
		PrimaryNS:  requestData.SOA.PrimaryNS,
		AdminEmail: requestData.SOA.AdminEmail,
		Refresh:    requestData.SOA.Refresh,
		Retry:      requestData.SOA.Retry,
		Expire:     requestData.SOA.Expire,
		Minimum:    requestData.SOA.Minimum,
		TTL:        requestData.SOA.TTL,
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

func UpdateZoneHandler(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var requestData struct {
		Name string `json:"name"`
		SOA  SOA    `json:"soa"`
	}
	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Invalid request body",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Check which fields need to be updated
	if requestData.Name != "" {
		zone.Name = requestData.Name
	}
	if requestData.SOA.PrimaryNS != "" {
		zone.PrimaryNS = requestData.SOA.PrimaryNS
	}
	if requestData.SOA.AdminEmail != "" {
		zone.AdminEmail = requestData.SOA.AdminEmail
	}
	if requestData.SOA.Refresh != 0 {
		zone.Refresh = requestData.SOA.Refresh
	}
	if requestData.SOA.Retry != 0 {
		zone.Retry = requestData.SOA.Retry
	}
	if requestData.SOA.Expire != 0 {
		zone.Expire = requestData.SOA.Expire
	}
	if requestData.SOA.Minimum != 0 {
		zone.Minimum = requestData.SOA.Minimum
	}
	if requestData.SOA.TTL != 0 {
		zone.TTL = requestData.SOA.TTL
	}

	// Update the zone
	err = bd.Zones.Update(zone)
	if err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Failed to update zone of UUID" + zoneUUID,
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Respond with success message
	responseBody := responseBody{
		Code:    0,
		Message: "Zone updated successfully",
		Data:    zone,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
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
