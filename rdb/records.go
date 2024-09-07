package rdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Record struct {
	UUID       string       // Record UUID
	Type       string       // Record type
	Host       string       // Record host
	Content    string       // Record content
	TTL        uint16       // Record TTL
	AddPTR     bool         // Add PTR record
	CreatedAt  time.Time    // Record creation time
	ModifiedAt time.Time    // Record modification time
	DeletedAt  sql.NullTime // Record deletion time
	ZoneUUID   string       // Record's zone UUID
	Staging    bool         // Record staging status
	Tags       []string     // Record tags
}

// Get retrieves records from the database based on the provided zone UUID.
//
// Parameters:
// - zoneUUID: The UUID of the zone to retrieve records from.
//
// Returns:
//   - []Record: A slice of Record structs representing the retrieved records.
//   - error: An error if the retrieval fails.
func (r *Record) Get(ctx context.Context) ([]Record, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, "SELECT r.uuid, r.type, r.host, r.content, r.ttl, r.add_ptr, r.created_at, r.modified_at, r.deleted_at, r.staging FROM bind_dns.records AS r JOIN bind_dns.zones AS z ON r.zone_uuid = z.uuid WHERE z.uuid::text = $1 AND (r.deleted_at IS NULL OR r.staging = TRUE)", r.ZoneUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		if err := rows.Scan(&record.UUID, &record.Type, &record.Host, &record.Content, &record.TTL, &record.AddPTR, &record.CreatedAt, &record.ModifiedAt, &record.DeletedAt, &record.Staging); err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}

		record.Tags, err = new(Tag).GetZone(ctx, r.ZoneUUID)
		record.ZoneUUID = r.ZoneUUID
		if err != nil {
			return nil, fmt.Errorf("failed to get tags: %w", err)
		}
		records = append(records, record)
	}
	return records, nil
}

// GetAll retrieves all records from the database.
//
// It returns a slice of Record and an error if any.
func (r *Record) GetAll(ctx context.Context) ([]Record, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, "SELECT uuid, type, host, content, ttl, add_ptr, created_at, modified_at, deleted_at, zone_uuid, staging FROM bind_dns.records WHERE deleted_at IS NULL OR staging = TRUE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		if err := rows.Scan(&record.UUID, &record.Type, &record.Host, &record.Content, &record.TTL, &record.AddPTR, &record.CreatedAt, &record.ModifiedAt, &record.DeletedAt, &record.ZoneUUID, &record.Staging); err != nil {
			return nil, err
		}

		record.Tags, err = new(Tag).GetZone(ctx, record.ZoneUUID)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}
	return records, nil
}

// Create inserts a new record into the database.
//
// Returns an error if the insertion fails.
func (r *Record) Create(ctx context.Context) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := "INSERT INTO bind_dns.records (uuid, type, host, content, ttl, add_ptr, created_at, modified_at, zone_uuid, staging) VALUES ($1, $2, $3, $4, $5, $6, $7, $7, $8, TRUE)"
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	timeNow := time.Now()
	r.CreatedAt = timeNow
	r.ModifiedAt = timeNow
	result, err := stmt.ExecContext(ctx, r.UUID, r.Type, r.Host, r.Content, r.TTL, r.AddPTR, timeNow, r.ZoneUUID)
	if err != nil {
		return err
	}

	// Check if any rows were inserted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	// add tags if any
	if len(r.Tags) > 0 {
		err = new(Tag).CreateRecord(ctx, r.UUID)
		if err != nil {
			return err
		}
	}

	return nil
}

// Find retrieves a record with the given UUID from the database.
//
// Returns an error if the retrieval fails.
func (r *Record) Find(ctx context.Context) error {
	query := "SELECT type, host, content, ttl, add_ptr, created_at, modified_at, deleted_at, zone_uuid, staging FROM bind_dns.records WHERE uuid::text = $1 AND (deleted_at IS NULL OR staging = TRUE)"
	row := db.QueryRow(query, r.UUID)
	err := row.Scan(&r.Type, &r.Host, &r.Content, &r.TTL, &r.AddPTR, &r.CreatedAt, &r.ModifiedAt, &r.DeletedAt, &r.ZoneUUID, &r.Staging)
	if err != nil {
		return err
	}

	r.Tags, err = new(Tag).GetZone(ctx, r.ZoneUUID)
	if err != nil {
		return err
	}
	return nil
}

// Update updates an existing record in the database.
//
// Returns an error if the update fails.
func (r *Record) Update(ctx context.Context) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := "UPDATE bind_dns.records SET type = $1, host = $2, content = $3, ttl = $4, add_ptr = $5, created_at = $6, modified_at = $7, staging = TRUE WHERE uuid::text = $8"
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	r.ModifiedAt = time.Now()

	// check if record is new
	if r.Staging && r.CreatedAt == r.ModifiedAt {
		r.CreatedAt = r.ModifiedAt
	}

	result, err := stmt.ExecContext(ctx, r.Type, r.Host, r.Content, r.TTL, r.AddPTR, r.CreatedAt, r.ModifiedAt, r.UUID)
	if err != nil {
		return err
	}

	// Log the output
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	// delete tags
	err = new(Tag).DeleteRecord(ctx, r.UUID)
	if err != nil {
		return err
	}

	// add tags if any
	if len(r.Tags) > 0 {
		err = new(Tag).CreateRecord(ctx, r.UUID)
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete deletes a record from the database.
//
// Returns an error if the deletion fails.
func (r *Record) Delete(ctx context.Context) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := "UPDATE bind_dns.records SET deleted_at = $1, staging = TRUE WHERE uuid::text = $2"
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, time.Now(), r.UUID)
	if err != nil {
		return err
	}

	// Log the output
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	// delete tags
	err = new(Tag).DeleteRecord(ctx, r.UUID)
	if err != nil {
		return err
	}

	return nil
}

// Commit commits the changes to the database.
// Sets all records to staging = FALSE
//
// Returns an error if the commit fails.
func (r *Record) CommitAll(ctx context.Context) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	//Check for any rows to commit
	query := "SELECT COUNT(*) FROM bind_dns.records WHERE staging = TRUE"
	row := tx.QueryRowContext(ctx, query)
	var count int
	err = row.Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}

	// Apply changes
	query = "UPDATE bind_dns.records SET staging = FALSE WHERE staging = TRUE"
	result, err := tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	// Check for any rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != int64(count) {
		return fmt.Errorf("expected %d rows affected, got %d", count, rowsAffected)
	}

	return nil
}

// GetStaging retrieves all records in the staging area.
//
// Returns:
//   - []Record: A slice of Record structs representing the retrieved records.
//   - error: An error if the retrieval fails.
func (r *Record) GetStaging(ctx context.Context) ([]Record, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := "SELECT uuid, type, host, content, ttl, add_ptr, created_at, modified_at, deleted_at, zone_uuid FROM bind_dns.records WHERE staging = TRUE"
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		err := rows.Scan(&record.UUID, &record.Type, &record.Host, &record.Content, &record.TTL, &record.AddPTR, &record.CreatedAt, &record.ModifiedAt, &record.DeletedAt, &record.ZoneUUID)
		if err != nil {
			return nil, err
		}
		record.Staging = true
		record.Tags, err = new(Tag).GetZone(ctx, record.ZoneUUID)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}
