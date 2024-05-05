package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/DrC0ns0le/bind-api/rdb"
	"github.com/google/uuid"
)

type Record struct {
	ID         int    `json:"-"`
	UUID       string `json:"uuid"`
	Type       string `json:"type"`
	Host       string `json:"host"`
	Content    string `json:"content"`
	TTL        int    `json:"ttl"`
	ModifiedAt int    `json:"modified_at"`
	DeletedAt  int    `json:"deleted_at"`
	ZoneUUID   string `json:"-"`
	Staging    bool   `json:"staging"`
}

type Records []Record

// Predefined namespace UUID for DNS purposes
var dnsNamespaceUUID = uuid.Must(uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))

func GetZoneRecordsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract zone UUID from URL
	zoneUUID := r.PathValue("zone_uuid")

	records, err := bd.Records.Get(zoneUUID)

	if err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Unable to retrieve records",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	var R Records

	for _, record := range records {
		temp := Record{
			UUID:       record.UUID,
			Type:       record.Type,
			Host:       record.Host,
			Content:    record.Content,
			TTL:        record.TTL,
			ModifiedAt: record.ModifiedAt,
			DeletedAt:  record.DeletedAt,
			Staging:    record.Staging,
		}
		R = append(R, temp)
	}

	responseBody := responseBody{
		Code:    0,
		Message: "Records retrieved successfully",
		Data:    R,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
}

func GetRecordHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Extract record UUID from URL
	zoneUUID := r.PathValue("zone_uuid")
	recordUUID := r.PathValue("record_uuid")

	// Find the zone by UUID and record by UUID
	record, err := bd.Records.Select(recordUUID)

	R := Record{
		UUID:       record.UUID,
		Type:       record.Type,
		Host:       record.Host,
		Content:    record.Content,
		TTL:        record.TTL,
		ModifiedAt: record.ModifiedAt,
		DeletedAt:  record.DeletedAt,
		Staging:    record.Staging,
	}

	if err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Record not found",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	if record.ZoneUUID != zoneUUID {
		errorMsg := responseBody{
			Code:    2,
			Message: "Record found, but zone does not match",
			Data:    map[string]string{"zone_uuid": record.ZoneUUID, "record_uuid": recordUUID},
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	responseBody := responseBody{
		Code:    0,
		Message: "Record retrieved successfully",
		Data:    R,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
}

func CreateRecordHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract record UUID from URL
	zoneUUID := r.PathValue("zone_uuid")

	// Find the zone by UUID
	_, err := bd.Zones.Select(zoneUUID)
	if err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Zone not found",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Parse request body
	var requestData struct {
		Type    string `json:"type"`
		Host    string `json:"host"`
		Content string `json:"content"`
		TTL     int    `json:"ttl"`
	}
	err = json.NewDecoder(r.Body).Decode(&requestData)
	missingFields := []string{}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if requestData.Type == "" {
		missingFields = append(missingFields, "type")
	} else if requestData.Host == "" {
		missingFields = append(missingFields, "host")
	} else if requestData.Content == "" {
		missingFields = append(missingFields, "content")
	}

	if len(missingFields) > 0 {
		errorMsg := responseBody{
			Code:    2,
			Message: "Missing fields",
			Data: map[string]string{
				"missing_fields": strings.Join(missingFields, ", "),
			},
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	newRecord := rdb.Record{
		UUID:    uuid.New().String(),
		Type:    requestData.Type,
		Host:    requestData.Host,
		Content: requestData.Content,
		TTL: func() int {
			if requestData.TTL == 0 {
				return 3600
			}
			return requestData.TTL
		}(),
		ModifiedAt: int(time.Now().Unix()),
		ZoneUUID:   zoneUUID,
		Staging:    true,
	}

	// Create the record
	err = bd.Records.Create(newRecord)
	if err != nil {
		errorMsg := responseBody{
			Code:    3,
			Message: "Faild to create record in database",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Respond with the created zone
	responseBody := responseBody{
		Code:    0,
		Message: "Record created successfully",
		Data:    newRecord,
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseBody)
}

func UpdateRecordHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Extract record UUID from URL
	zoneUUID := r.PathValue("zone_uuid")
	recordUUID := r.PathValue("record_uuid")

	// Find the zone by UUID and record by UUID
	record, err := bd.Records.Select(recordUUID)

	if err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Record not found",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	if record.ZoneUUID != zoneUUID {
		errorMsg := responseBody{
			Code:    2,
			Message: "Record found, but zone does not match",
			Data:    map[string]string{"zone_uuid": record.ZoneUUID, "record_uuid": recordUUID},
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Parse request body
	var requestData struct {
		Type    string `json:"type"`
		Host    string `json:"host"`
		Content string `json:"content"`
		TTL     int    `json:"ttl"`
	}
	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		errorMsg := responseBody{
			Code:    3,
			Message: "Unable to parse request body",
			Data:    err.Error(),
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Update record fields if provided in request
	if requestData.Type != "" {
		record.Type = requestData.Type
	}
	if requestData.Host != "" {
		record.Host = requestData.Host
	}
	if requestData.Content != "" {
		record.Content = requestData.Content
	}
	if requestData.TTL != 0 {
		record.TTL = requestData.TTL
	}

	// Update ModifiedAt timestamp
	record.ModifiedAt = int(time.Now().Unix())

	// Update the record
	err = bd.Records.Update(record)
	if err != nil {
		errorMsg := responseBody{
			Code:    3,
			Message: "Faild to update record in database",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Respond with the updated zone
	response := responseBody{
		Code:    0,
		Message: "Record updated successfully",
		Data:    record,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func DeleteRecordHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Extract record UUID from URL
	zoneUUID := r.PathValue("zone_uuid")
	recordUUID := r.PathValue("record_uuid")

	// Find the zone by UUID and record by UUID
	record, err := bd.Records.Select(recordUUID)

	if err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Record not found",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	if record.ZoneUUID != zoneUUID {
		errorMsg := responseBody{
			Code:    1,
			Message: "Record found, but zone does not match",
			Data:    nil,
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Delete record from database
	err = bd.Records.Delete(recordUUID)

	if err != nil {
		errorMsg := responseBody{
			Code:    2,
			Message: "Failed to delete record of UUID" + recordUUID,
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Respond with success message
	successMsg := responseBody{
		Code:    0,
		Message: "Record deleted successfully",
		Data:    nil,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(successMsg)

}
