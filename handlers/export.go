package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/DrC0ns0le/bind-api/rdb"
)

// ExportAllHandler exports all zones and configs as JSON backup
func ExportAllHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := rdb.ExportAllJSON(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to export data: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("bind-api-backup-%s.json", time.Now().Format("2006-01-02-150405"))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// ExportZonesHandler exports all zones as JSON backup
func ExportZonesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := rdb.ExportZonesJSON(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to export zones: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("bind-api-zones-%s.json", time.Now().Format("2006-01-02-150405"))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// ExportZoneHandler exports a single zone by UUID as JSON backup
func ExportZoneHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	zoneUUID := r.PathValue("zone_uuid")

	if zoneUUID == "" {
		http.Error(w, `{"error": "zone_uuid is required"}`, http.StatusBadRequest)
		return
	}

	data, err := rdb.ExportZoneByUUIDJSON(ctx, zoneUUID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to export zone: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("bind-api-zone-%s-%s.json", zoneUUID[:8], time.Now().Format("2006-01-02-150405"))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// ExportConfigsHandler exports all configs as JSON backup
func ExportConfigsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := rdb.ExportConfigsJSON(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to export configs: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("bind-api-configs-%s.json", time.Now().Format("2006-01-02-150405"))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// ImportRequest represents the import request body options
type ImportRequest struct {
	SkipExisting      bool `json:"skip_existing"`
	OverwriteExisting bool `json:"overwrite_existing"`
	PreserveUUIDs     bool `json:"preserve_uuids"`
	DryRun            bool `json:"dry_run"`
}

// parseImportOptions extracts import options from query params
func parseImportOptions(r *http.Request) rdb.ImportOptions {
	return rdb.ImportOptions{
		SkipExisting:      r.URL.Query().Get("skip_existing") == "true",
		OverwriteExisting: r.URL.Query().Get("overwrite") == "true",
		PreserveUUIDs:     r.URL.Query().Get("preserve_uuids") == "true",
		DryRun:            r.URL.Query().Get("dry_run") == "true",
	}
}

// ImportAllHandler imports zones and configs from JSON backup
func ImportAllHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Read request body
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to read request body: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(data) == 0 {
		http.Error(w, `{"error": "request body is empty"}`, http.StatusBadRequest)
		return
	}

	opts := parseImportOptions(r)

	result, err := rdb.ImportAllJSON(ctx, data, opts)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to import data: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// ImportZonesHandler imports zones from JSON backup
func ImportZonesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to read request body: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(data) == 0 {
		http.Error(w, `{"error": "request body is empty"}`, http.StatusBadRequest)
		return
	}

	opts := parseImportOptions(r)

	result, err := rdb.ImportZonesJSON(ctx, data, opts)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to import zones: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// ImportZoneHandler imports a single zone from JSON backup
func ImportZoneHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to read request body: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(data) == 0 {
		http.Error(w, `{"error": "request body is empty"}`, http.StatusBadRequest)
		return
	}

	opts := parseImportOptions(r)

	result, err := rdb.ImportZoneJSON(ctx, data, opts)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to import zone: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// ImportConfigsHandler imports configs from JSON backup
func ImportConfigsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to read request body: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(data) == 0 {
		http.Error(w, `{"error": "request body is empty"}`, http.StatusBadRequest)
		return
	}

	opts := parseImportOptions(r)

	result, err := rdb.ImportConfigsJSON(ctx, data, opts)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to import configs: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
