package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/DrC0ns0le/bind-api/rdb"
	"github.com/google/uuid"
)

type Record struct {
	ID         uint32   `json:"-"`
	UUID       string   `json:"uuid"`
	Type       string   `json:"type"`
	Host       string   `json:"host"`
	Content    string   `json:"content"`
	TTL        uint16   `json:"ttl"`
	AddPTR     bool     `json:"add_ptr"`
	CreatedAt  uint64   `json:"created_at"`
	ModifiedAt uint64   `json:"modified_at"`
	DeletedAt  uint64   `json:"deleted_at"`
	ZoneUUID   string   `json:"-"`
	Staging    bool     `json:"staging"`
	Tags       []string `json:"tags"`
}

type Records []Record

// Predefined namespace UUID for DNS purposes
var dnsNamespaceUUID = uuid.Must(uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))

func GetZoneRecordsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract zone UUID from URL
	zoneUUID := r.PathValue("zone_uuid")

	records, err := (&rdb.Record{ZoneUUID: zoneUUID}).Get(r.Context())

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
			AddPTR:     record.AddPTR,
			CreatedAt:  uint64(record.CreatedAt.Unix()),
			ModifiedAt: uint64(record.ModifiedAt.Unix()),
			DeletedAt: func(t sql.NullTime) uint64 {
				if t.Valid {
					return uint64(t.Time.Unix())
				}
				return 0
			}(record.DeletedAt),
			Staging: record.Staging,
			Tags:    record.Tags,
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
	record := rdb.Record{UUID: r.PathValue("record_uuid")}

	// Find the zone by UUID and record by UUID
	if err := record.Find(r.Context()); err != nil {
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
			Data:    map[string]string{"zone_uuid": record.ZoneUUID, "record_uuid": record.UUID},
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	responseBody := responseBody{
		Code:    0,
		Message: "Record retrieved successfully",
		Data: Record{
			UUID:       record.UUID,
			Type:       record.Type,
			Host:       record.Host,
			Content:    record.Content,
			TTL:        record.TTL,
			CreatedAt:  uint64(record.CreatedAt.Unix()),
			ModifiedAt: uint64(record.ModifiedAt.Unix()),
			DeletedAt: func(t sql.NullTime) uint64 {
				if t.Valid {
					return uint64(t.Time.Unix())
				}
				return 0
			}(record.DeletedAt),
			Staging: record.Staging,
			Tags:    record.Tags,
		},
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
}

func CreateRecordHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract record UUID from URL
	zone := rdb.Zone{UUID: r.PathValue("zone_uuid")}

	// Find the zone by UUID
	if err := zone.Find(r.Context()); err != nil {
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
		Type    string   `json:"type"`
		Host    string   `json:"host"`
		Content string   `json:"content"`
		TTL     uint16   `json:"ttl"`
		AddPTR  bool     `json:"add_ptr"`
		Tags    []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var missingFields []string
	if requestData.Type == "" {
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
		TTL: func() uint16 {
			if requestData.TTL == 0 {
				return 3600
			}
			return requestData.TTL
		}(),
		AddPTR:   requestData.AddPTR,
		ZoneUUID: zone.UUID,
		Staging:  true,
		Tags:     requestData.Tags,
	}

	// Create the record
	if err := newRecord.Create(r.Context()); err != nil {
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
	record := rdb.Record{UUID: r.PathValue("record_uuid")}

	// Find the zone by UUID and record by UUID
	if err := record.Find(r.Context()); err != nil {
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
			Data:    map[string]string{"zone_uuid": record.ZoneUUID, "record_uuid": record.UUID},
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Parse request body
	var requestData struct {
		Type    string   `json:"type"`
		Host    string   `json:"host"`
		Content string   `json:"content"`
		TTL     uint16   `json:"ttl"`
		AddPTR  bool     `json:"add_ptr"`
		Tags    []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
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
	newRecord := record
	needUpdate := false
	if requestData.Type != "" {
		record.Type = requestData.Type
		needUpdate = true
	}
	if requestData.Host != "" {
		record.Host = requestData.Host
		needUpdate = true
	}
	if requestData.Content != "" {
		record.Content = requestData.Content
		needUpdate = true
	}
	if requestData.TTL != 0 {
		record.TTL = requestData.TTL
		needUpdate = true
	}
	if len(requestData.Tags) > 0 {
		record.Tags = requestData.Tags
		needUpdate = true
	}

	if newRecord.AddPTR != requestData.AddPTR {
		record.AddPTR = requestData.AddPTR
		needUpdate = true
	}

	// Check if theres a need to update
	if !needUpdate {
		errorMsg := responseBody{
			Code:    4,
			Message: "No changes found, nothing to update",
			Data:    nil,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Update the record
	if err := record.Update(r.Context()); err != nil {
		errorMsg := responseBody{
			Code:    3,
			Message: "Failed to update record in database",
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
	record := rdb.Record{UUID: r.PathValue("record_uuid")}

	// Find the zone by UUID and record by UUID
	if err := record.Find(r.Context()); err != nil {
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
	if err := record.Delete(r.Context()); err != nil {
		errorMsg := responseBody{
			Code:    2,
			Message: "Failed to delete record of UUID " + record.UUID + " from database",
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
