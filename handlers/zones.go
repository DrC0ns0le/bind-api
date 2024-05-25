package handlers

import (
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
	Expire     uint16 `json:"expire"`
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
	type requestSOA struct {
		PrimaryNS  string `json:"primary_ns"`
		AdminEmail string `json:"admin_email"`
		Refresh    uint16 `json:"refresh"`
		Retry      uint16 `json:"retry"`
		Expire     uint16 `json:"expire"`
		Minimum    uint16 `json:"minimum"`
		TTL        uint16 `json:"ttl"`
	}
	var requestData struct {
		Name string     `json:"name"`
		SOA  requestSOA `json:"soa"`
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
