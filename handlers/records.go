package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
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

	// Parse pagination parameters
	page := 1
	pageSize := 25 // Default page size
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 {
			pageSize = val
		}
	}

	// Get records with pagination
	records, total, err := (&rdb.Record{
		ZoneUUID: zoneUUID,
		Type:     r.URL.Query().Get("type"),
		Host:     r.URL.Query().Get("host"),
		Content:  r.URL.Query().Get("content"),
		Tags:     strings.Split(r.URL.Query().Get("tags"), ","),
	}).Get(r.Context(), page, pageSize, r.URL.Query().Get("search"))

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

	// Calculate pagination metadata
	totalPages := (total + pageSize - 1) / pageSize
	hasNext := page < totalPages
	hasPrev := page > 1

	responseBody := responseBody{
		Code:    0,
		Message: "Records retrieved successfully",
		Data: struct {
			Records    []Record           `json:"records"`
			Pagination paginationMetadata `json:"pagination"`
		}{
			Records: R,
			Pagination: paginationMetadata{
				CurrentPage: page,
				PageSize:    pageSize,
				TotalPages:  totalPages,
				TotalItems:  total,
				HasNextPage: hasNext,
				HasPrevPage: hasPrev,
			},
		},
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
		errorMsg := responseBody{
			Code:    2,
			Message: "Invalid request body",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Normalize record type to uppercase
	requestData.Type = strings.ToUpper(requestData.Type)

	// Set default TTL if not provided
	if requestData.TTL == 0 {
		requestData.TTL = 3600
	}

	// Validate record fields
	validationErrors := ValidateRecord(
		requestData.Type,
		requestData.Host,
		requestData.Content,
		requestData.TTL,
		requestData.AddPTR,
	)

	if validationErrors.HasErrors() {
		errorMsg := responseBody{
			Code:    3,
			Message: "Validation failed",
			Data:    validationErrors,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	newRecord := rdb.Record{
		UUID:     uuid.New().String(),
		Type:     requestData.Type,
		Host:     requestData.Host,
		Content:  requestData.Content,
		TTL:      requestData.TTL,
		AddPTR:   requestData.AddPTR,
		ZoneUUID: zone.UUID,
		Staging:  true,
		Tags:     requestData.Tags,
	}

	// Create the record
	if err := newRecord.Create(r.Context()); err != nil {
		errorMsg := responseBody{
			Code:    4,
			Message: "Failed to create record in database",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Respond with the created record
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
		Type    *string   `json:"type"`
		Host    *string   `json:"host"`
		Content *string   `json:"content"`
		TTL     *uint16   `json:"ttl"`
		AddPTR  *bool     `json:"add_ptr"`
		Tags    *[]string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		errorMsg := responseBody{
			Code:    3,
			Message: "Unable to parse request body",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Update record fields if provided in request
	needUpdate := false
	if requestData.Type != nil {
		record.Type = strings.ToUpper(*requestData.Type)
		needUpdate = true
	}
	if requestData.Host != nil {
		record.Host = *requestData.Host
		needUpdate = true
	}
	if requestData.Content != nil {
		record.Content = *requestData.Content
		needUpdate = true
	}
	if requestData.TTL != nil {
		record.TTL = *requestData.TTL
		needUpdate = true
	}
	if requestData.Tags != nil {
		record.Tags = *requestData.Tags
		needUpdate = true
	}
	if requestData.AddPTR != nil {
		record.AddPTR = *requestData.AddPTR
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

	// Validate updated record fields
	validationErrors := ValidateRecord(
		record.Type,
		record.Host,
		record.Content,
		record.TTL,
		record.AddPTR,
	)

	if validationErrors.HasErrors() {
		errorMsg := responseBody{
			Code:    5,
			Message: "Validation failed",
			Data:    validationErrors,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Update the record
	if err := record.Update(r.Context()); err != nil {
		errorMsg := responseBody{
			Code:    6,
			Message: "Failed to update record in database",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
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

// GetSupportedRecordTypesHandler returns all supported DNS record types
func GetSupportedRecordTypesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	types := make([]string, 0, len(supportedRecordTypes))
	for t := range supportedRecordTypes {
		types = append(types, t)
	}
	sort.Strings(types)

	response := responseBody{
		Code:    0,
		Message: "Supported record types retrieved successfully",
		Data: map[string][]string{
			"types": types,
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

