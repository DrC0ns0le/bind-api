package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/DrC0ns0le/bind-api/commit"
	"github.com/DrC0ns0le/bind-api/rdb"
	"github.com/DrC0ns0le/bind-api/render"
)

// GetStagingHandler retrieves all zones and records in staging and returns them in a JSON response.
//
// Parameters:
// - w: http.ResponseWriter - the response writer used to write the response.
// - r: *http.Request - the HTTP request object.
//
// Returns:
// None.
func GetStagingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	Z, R, err := getAllStaging(r.Context())
	if err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Unable to retrieve changes in staging",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	response := responseBody{
		Code:    0,
		Message: "Changes in staging successfully retrieved",
		Data: struct {
			Zones   []Zone   `json:"zones"`
			Records []Record `json:"records"`
		}{
			Zones:   Z,
			Records: R,
		},
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func ApplyStagingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Render all zones
	if err := render.RenderZonesTemplate(r.Context()); err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Zone rendering failed",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Commit changes
	if err := commit.Push(); err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Unable to commit changes",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	// Commit all changes
	err := (&rdb.Record{}).CommitAll(r.Context())
	if err != nil {
		return
	}

	// set config_status to awaiting_deployment
	if err := (&rdb.Config{ConfigKey: "config_status", ConfigValue: "deployed"}).Update(r.Context(), "awaiting_deployment"); err != nil {
		errorMsg := responseBody{
			Code:    1,
			Message: "Unable to update deploy_status",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	responseBody := responseBody{
		Code:    0,
		Message: "Changes successfully committed",
		Data:    nil,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
}

func getAllStaging(ctx context.Context) ([]Zone, []Record, error) {
	// Get all zones in staging
	zones, err := (&rdb.Zone{}).GetStaging(ctx)
	if err != nil {
		return nil, nil, err
	}

	var Z Zones

	for _, zone := range zones {
		temp := Zone{
			UUID:       zone.UUID,
			Name:       zone.Name,
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
			Tags: zone.Tags,
		}
		Z = append(Z, temp)
	}

	// Get all records in staging
	records, err := (&rdb.Record{}).GetStaging(ctx)
	if err != nil {
		return nil, nil, err
	}

	var R Records

	for _, record := range records {
		temp := Record{
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
		}
		R = append(R, temp)
	}

	return Z, R, nil
}
