package rdb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ExportedZone represents a zone with all its records and tags for export
type ExportedZone struct {
	UUID       string           `json:"uuid"`
	Name       string           `json:"name"`
	PrimaryNS  string           `json:"primary_ns"`
	AdminEmail string           `json:"admin_email"`
	Refresh    uint16           `json:"refresh"`
	Retry      uint16           `json:"retry"`
	Expire     uint32           `json:"expire"`
	Minimum    uint16           `json:"minimum"`
	TTL        uint16           `json:"ttl"`
	Tags       []string         `json:"tags,omitempty"`
	Records    []ExportedRecord `json:"records,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	ModifiedAt time.Time        `json:"modified_at"`
}

// ExportedRecord represents a record for export
type ExportedRecord struct {
	UUID       string    `json:"uuid"`
	Type       string    `json:"type"`
	Host       string    `json:"host"`
	Content    string    `json:"content"`
	TTL        uint16    `json:"ttl"`
	AddPTR     bool      `json:"add_ptr"`
	Tags       []string  `json:"tags,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

// ExportedConfig represents a config entry for export
type ExportedConfig struct {
	ConfigKey   string    `json:"config_key"`
	ConfigValue string    `json:"config_value"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
}

// ExportData represents the complete export structure
type ExportData struct {
	ExportedAt time.Time        `json:"exported_at"`
	Version    string           `json:"version"`
	Zones      []ExportedZone   `json:"zones"`
	Configs    []ExportedConfig `json:"configs"`
}

// ExportZones exports all active zones with their records and tags
func ExportZones(ctx context.Context) ([]ExportedZone, error) {
	zones, err := (&Zone{}).Get(ctx)
	if err != nil {
		return nil, err
	}

	var exported []ExportedZone
	for _, z := range zones {
		// Skip deleted zones
		if z.DeletedAt.Valid {
			continue
		}

		// Get records for this zone
		records, _, err := (&Record{ZoneUUID: z.UUID}).Get(ctx, 0, 0, "")
		if err != nil {
			return nil, err
		}

		var exportedRecords []ExportedRecord
		for _, r := range records {
			// Skip deleted records
			if r.DeletedAt.Valid {
				continue
			}

			exportedRecords = append(exportedRecords, ExportedRecord{
				UUID:       r.UUID,
				Type:       r.Type,
				Host:       r.Host,
				Content:    r.Content,
				TTL:        r.TTL,
				AddPTR:     r.AddPTR,
				Tags:       r.Tags,
				CreatedAt:  r.CreatedAt,
				ModifiedAt: r.ModifiedAt,
			})
		}

		exported = append(exported, ExportedZone{
			UUID:       z.UUID,
			Name:       z.Name,
			PrimaryNS:  z.PrimaryNS,
			AdminEmail: z.AdminEmail,
			Refresh:    z.Refresh,
			Retry:      z.Retry,
			Expire:     z.Expire,
			Minimum:    z.Minimum,
			TTL:        z.TTL,
			Tags:       z.Tags,
			Records:    exportedRecords,
			CreatedAt:  z.CreatedAt,
			ModifiedAt: z.ModifiedAt,
		})
	}

	return exported, nil
}

// ExportConfigs exports all active configs
func ExportConfigs(ctx context.Context) ([]ExportedConfig, error) {
	configs, err := (&Config{}).Get(ctx)
	if err != nil {
		return nil, err
	}

	var exported []ExportedConfig
	for _, c := range configs {
		// Skip deleted configs
		if c.DeletedAt.Valid {
			continue
		}

		exported = append(exported, ExportedConfig{
			ConfigKey:   c.ConfigKey,
			ConfigValue: c.ConfigValue,
			CreatedAt:   c.CreatedAt,
			ModifiedAt:  c.ModifiedAt,
		})
	}

	return exported, nil
}

// ExportAll exports all zones and configs as a complete backup
func ExportAll(ctx context.Context) (*ExportData, error) {
	zones, err := ExportZones(ctx)
	if err != nil {
		return nil, err
	}

	configs, err := ExportConfigs(ctx)
	if err != nil {
		return nil, err
	}

	return &ExportData{
		ExportedAt: time.Now(),
		Version:    "1.0",
		Zones:      zones,
		Configs:    configs,
	}, nil
}

// ExportAllJSON exports all data as JSON bytes
func ExportAllJSON(ctx context.Context) ([]byte, error) {
	data, err := ExportAll(ctx)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(data, "", "  ")
}

// ExportZonesJSON exports zones as JSON bytes
func ExportZonesJSON(ctx context.Context) ([]byte, error) {
	zones, err := ExportZones(ctx)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(zones, "", "  ")
}

// ExportConfigsJSON exports configs as JSON bytes
func ExportConfigsJSON(ctx context.Context) ([]byte, error) {
	configs, err := ExportConfigs(ctx)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(configs, "", "  ")
}

// ExportZoneByUUID exports a single zone by UUID with all its records
func ExportZoneByUUID(ctx context.Context, uuid string) (*ExportedZone, error) {
	z := &Zone{UUID: uuid}
	if err := z.Find(ctx); err != nil {
		return nil, err
	}

	// Get records for this zone
	records, _, err := (&Record{ZoneUUID: z.UUID}).Get(ctx, 0, 0, "")
	if err != nil {
		return nil, err
	}

	var exportedRecords []ExportedRecord
	for _, r := range records {
		if r.DeletedAt.Valid {
			continue
		}

		exportedRecords = append(exportedRecords, ExportedRecord{
			UUID:       r.UUID,
			Type:       r.Type,
			Host:       r.Host,
			Content:    r.Content,
			TTL:        r.TTL,
			AddPTR:     r.AddPTR,
			Tags:       r.Tags,
			CreatedAt:  r.CreatedAt,
			ModifiedAt: r.ModifiedAt,
		})
	}

	return &ExportedZone{
		UUID:       z.UUID,
		Name:       z.Name,
		PrimaryNS:  z.PrimaryNS,
		AdminEmail: z.AdminEmail,
		Refresh:    z.Refresh,
		Retry:      z.Retry,
		Expire:     z.Expire,
		Minimum:    z.Minimum,
		TTL:        z.TTL,
		Tags:       z.Tags,
		Records:    exportedRecords,
		CreatedAt:  z.CreatedAt,
		ModifiedAt: z.ModifiedAt,
	}, nil
}

// ExportZoneByUUIDJSON exports a single zone as JSON bytes
func ExportZoneByUUIDJSON(ctx context.Context, uuid string) ([]byte, error) {
	zone, err := ExportZoneByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(zone, "", "  ")
}

// ImportResult contains the results of an import operation
type ImportResult struct {
	ZonesImported   int      `json:"zones_imported"`
	RecordsImported int      `json:"records_imported"`
	ConfigsImported int      `json:"configs_imported"`
	ZonesSkipped    int      `json:"zones_skipped"`
	RecordsSkipped  int      `json:"records_skipped"`
	ConfigsSkipped  int      `json:"configs_skipped"`
	Errors          []string `json:"errors,omitempty"`
}

// ImportOptions configures import behavior
type ImportOptions struct {
	SkipExisting   bool // Skip records that already exist (by UUID)
	OverwriteExisting bool // Overwrite existing records
	PreserveUUIDs  bool // Use UUIDs from import file (default: generate new)
	DryRun         bool // Don't actually import, just validate
}

// ImportAllJSON imports zones and configs from JSON bytes
func ImportAllJSON(ctx context.Context, data []byte, opts ImportOptions) (*ImportResult, error) {
	var exportData ExportData
	if err := json.Unmarshal(data, &exportData); err != nil {
		return nil, fmt.Errorf("failed to parse import data: %w", err)
	}

	return ImportAll(ctx, exportData.Zones, exportData.Configs, opts)
}

// ImportAll imports zones and configs
func ImportAll(ctx context.Context, zones []ExportedZone, configs []ExportedConfig, opts ImportOptions) (*ImportResult, error) {
	result := &ImportResult{}

	// Import configs first
	configResult, err := ImportConfigs(ctx, configs, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to import configs: %w", err)
	}
	result.ConfigsImported = configResult.ConfigsImported
	result.ConfigsSkipped = configResult.ConfigsSkipped
	result.Errors = append(result.Errors, configResult.Errors...)

	// Import zones
	zoneResult, err := ImportZones(ctx, zones, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to import zones: %w", err)
	}
	result.ZonesImported = zoneResult.ZonesImported
	result.ZonesSkipped = zoneResult.ZonesSkipped
	result.RecordsImported = zoneResult.RecordsImported
	result.RecordsSkipped = zoneResult.RecordsSkipped
	result.Errors = append(result.Errors, zoneResult.Errors...)

	return result, nil
}

// ImportZones imports zones with their records and tags
func ImportZones(ctx context.Context, zones []ExportedZone, opts ImportOptions) (*ImportResult, error) {
	result := &ImportResult{}

	for _, ez := range zones {
		zoneImported, recordsImported, recordsSkipped, err := importZone(ctx, ez, opts)
		if err != nil {
			if opts.SkipExisting {
				result.ZonesSkipped++
				result.Errors = append(result.Errors, fmt.Sprintf("zone %s: %v", ez.Name, err))
				continue
			}
			return result, err
		}
		if zoneImported {
			result.ZonesImported++
		} else {
			result.ZonesSkipped++
		}
		result.RecordsImported += recordsImported
		result.RecordsSkipped += recordsSkipped
	}

	return result, nil
}

// importZone imports a single zone with its records
func importZone(ctx context.Context, ez ExportedZone, opts ImportOptions) (zoneImported bool, recordsImported, recordsSkipped int, err error) {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, 0, 0, err
	}
	defer tx.Rollback(ctx)

	// Check if zone exists
	existingZone := &Zone{UUID: ez.UUID}
	zoneExists := existingZone.Find(ctx) == nil

	var zoneUUID string

	if zoneExists {
		if opts.SkipExisting && !opts.OverwriteExisting {
			// Skip this zone but still return success
			return false, 0, len(ez.Records), nil
		}
		if opts.OverwriteExisting {
			// Update existing zone
			existingZone.Name = ez.Name
			existingZone.PrimaryNS = ez.PrimaryNS
			existingZone.AdminEmail = ez.AdminEmail
			existingZone.Refresh = ez.Refresh
			existingZone.Retry = ez.Retry
			existingZone.Expire = ez.Expire
			existingZone.Minimum = ez.Minimum
			existingZone.TTL = ez.TTL
			existingZone.Tags = ez.Tags
			existingZone.Staging = true
			if !opts.DryRun {
				if err := existingZone.Update(ctx); err != nil {
					return false, 0, 0, fmt.Errorf("failed to update zone: %w", err)
				}
			}
			zoneUUID = existingZone.UUID
			zoneImported = true
		}
	} else {
		// Create new zone
		newZone := &Zone{
			Name:       ez.Name,
			PrimaryNS:  ez.PrimaryNS,
			AdminEmail: ez.AdminEmail,
			Refresh:    ez.Refresh,
			Retry:      ez.Retry,
			Expire:     ez.Expire,
			Minimum:    ez.Minimum,
			TTL:        ez.TTL,
			Tags:       ez.Tags,
			Staging:    true,
		}
		if opts.PreserveUUIDs && ez.UUID != "" {
			newZone.UUID = ez.UUID
		} else {
			newZone.UUID = generateUUID()
		}
		if !opts.DryRun {
			if err := newZone.Create(ctx); err != nil {
				return false, 0, 0, fmt.Errorf("failed to create zone: %w", err)
			}
		}
		zoneUUID = newZone.UUID
		zoneImported = true
	}

	// Import records
	for _, er := range ez.Records {
		imported, err := importRecord(ctx, zoneUUID, er, opts)
		if err != nil {
			if opts.SkipExisting {
				recordsSkipped++
				continue
			}
			return zoneImported, recordsImported, recordsSkipped, err
		}
		if imported {
			recordsImported++
		} else {
			recordsSkipped++
		}
	}

	if !opts.DryRun {
		if err := tx.Commit(ctx); err != nil {
			return false, 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return zoneImported, recordsImported, recordsSkipped, nil
}

// importRecord imports a single record
func importRecord(ctx context.Context, zoneUUID string, er ExportedRecord, opts ImportOptions) (bool, error) {
	// Check if record exists
	existingRecord := &Record{UUID: er.UUID}
	recordExists := existingRecord.Find(ctx) == nil

	if recordExists {
		if opts.SkipExisting && !opts.OverwriteExisting {
			return false, nil
		}
		if opts.OverwriteExisting {
			existingRecord.Type = er.Type
			existingRecord.Host = er.Host
			existingRecord.Content = er.Content
			existingRecord.TTL = er.TTL
			existingRecord.AddPTR = er.AddPTR
			existingRecord.Tags = er.Tags
			existingRecord.Staging = true
			if !opts.DryRun {
				if err := existingRecord.Update(ctx); err != nil {
					return false, fmt.Errorf("failed to update record: %w", err)
				}
			}
			return true, nil
		}
	}

	// Create new record
	newRecord := &Record{
		Type:     er.Type,
		Host:     er.Host,
		Content:  er.Content,
		TTL:      er.TTL,
		AddPTR:   er.AddPTR,
		Tags:     er.Tags,
		ZoneUUID: zoneUUID,
		Staging:  true,
	}
	if opts.PreserveUUIDs && er.UUID != "" {
		newRecord.UUID = er.UUID
	} else {
		newRecord.UUID = generateUUID()
	}
	if !opts.DryRun {
		if err := newRecord.Create(ctx); err != nil {
			return false, fmt.Errorf("failed to create record: %w", err)
		}
	}
	return true, nil
}

// ImportConfigs imports configs
func ImportConfigs(ctx context.Context, configs []ExportedConfig, opts ImportOptions) (*ImportResult, error) {
	result := &ImportResult{}

	for _, ec := range configs {
		imported, err := importConfig(ctx, ec, opts)
		if err != nil {
			if opts.SkipExisting {
				result.ConfigsSkipped++
				result.Errors = append(result.Errors, fmt.Sprintf("config %s: %v", ec.ConfigKey, err))
				continue
			}
			return result, err
		}
		if imported {
			result.ConfigsImported++
		} else {
			result.ConfigsSkipped++
		}
	}

	return result, nil
}

// importConfig imports a single config
func importConfig(ctx context.Context, ec ExportedConfig, opts ImportOptions) (bool, error) {
	// Check if config exists
	existingConfig := &Config{ConfigKey: ec.ConfigKey}
	configs, err := existingConfig.Find(ctx)
	if err != nil && err.Error() != "no rows in result set" {
		return false, err
	}

	// Check if exact key-value pair exists
	for _, c := range configs {
		if c.ConfigValue == ec.ConfigValue {
			if opts.SkipExisting && !opts.OverwriteExisting {
				return false, nil
			}
			// Config with same key-value already exists
			return false, nil
		}
	}

	// Create new config
	newConfig := &Config{
		ConfigKey:   ec.ConfigKey,
		ConfigValue: ec.ConfigValue,
		Staging:     true,
	}
	if !opts.DryRun {
		if err := newConfig.Create(ctx); err != nil {
			return false, fmt.Errorf("failed to create config: %w", err)
		}
	}
	return true, nil
}

// ImportZonesJSON imports zones from JSON bytes
func ImportZonesJSON(ctx context.Context, data []byte, opts ImportOptions) (*ImportResult, error) {
	var zones []ExportedZone
	if err := json.Unmarshal(data, &zones); err != nil {
		return nil, fmt.Errorf("failed to parse zones data: %w", err)
	}

	return ImportZones(ctx, zones, opts)
}

// ImportConfigsJSON imports configs from JSON bytes
func ImportConfigsJSON(ctx context.Context, data []byte, opts ImportOptions) (*ImportResult, error) {
	var configs []ExportedConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse configs data: %w", err)
	}

	return ImportConfigs(ctx, configs, opts)
}

// ImportZoneJSON imports a single zone from JSON bytes
func ImportZoneJSON(ctx context.Context, data []byte, opts ImportOptions) (*ImportResult, error) {
	var zone ExportedZone
	if err := json.Unmarshal(data, &zone); err != nil {
		return nil, fmt.Errorf("failed to parse zone data: %w", err)
	}

	return ImportZones(ctx, []ExportedZone{zone}, opts)
}

// generateUUID generates a new UUID string
func generateUUID() string {
	return uuid.New().String()
}
