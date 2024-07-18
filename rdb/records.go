package rdb

import (
	"database/sql"
	"fmt"
	"time"
)

type Record struct {
	UUID       string // Record UUID
	Type       string // Record type
	Host       string // Record host
	Content    string // Record content
	TTL        uint16 // Record TTL
	AddPTR     bool   // Add PTR record
	CreatedAt  uint64 // Record creation time
	ModifiedAt uint64 // Record modification time
	DeletedAt  uint64 // Record deletion time
	ZoneUUID   string // Record's zone UUID
	Staging    bool   // Record staging status
	Tags       string // Record tags
}

// Get retrieves records from the database based on the provided zone UUID.
//
// Parameters:
// - zoneUUID: The UUID of the zone to retrieve records from.
//
// Returns:
// - []Record: A slice of Record structs representing the retrieved records.
// - error: An error if the retrieval fails.
func (r *Record) Get(zoneUUID string) ([]Record, error) {
	rows, err := db.Query("SELECT r.uuid, r.type, r.host, r.content, r.ttl, r.add_ptr, r.created_at, r.modified_at, r.deleted_at, r.staging, r.tags FROM bind_dns.records AS r JOIN bind_dns.zones AS z ON r.zone_uuid = z.uuid WHERE z.uuid = $1 AND (r.deleted_at = 0 OR r.staging = TRUE)", zoneUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		err := rows.Scan(&record.UUID, &record.Type, &record.Host, &record.Content, &record.TTL, &record.AddPTR, &record.CreatedAt, &record.ModifiedAt, &record.DeletedAt, &record.Staging, &record.Tags)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

// GetAll retrieves all records from the database.
//
// It returns a slice of Record and an error if any.
func (r *Record) GetAll() ([]Record, error) {
	rows, err := db.Query("SELECT uuid, type, host, content, ttl, add_ptr, created_at, modified_at, deleted_at, zone_uuid, staging, tags FROM bind_dns.records WHERE deleted_at = 0 OR staging = TRUE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		err := rows.Scan(&record.UUID, &record.Type, &record.Host, &record.Content, &record.TTL, &record.AddPTR, &record.CreatedAt, &record.ModifiedAt, &record.DeletedAt, &record.ZoneUUID, &record.Staging, &record.Tags)
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
func (r *Record) Create() error {
	query := "INSERT INTO bind_dns.records (uuid, type, host, content, ttl, add_ptr, created_at, modified_at, deleted_at, zone_uuid, staging, tags) VALUES ($1, $2, $3, $4, $5, $6, $7, $7, 0, $8, TRUE, $9)"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	timeNow := time.Now().Unix()
	r.CreatedAt = uint64(timeNow)
	r.ModifiedAt = uint64(timeNow)
	result, err := stmt.Exec(r.UUID, r.Type, r.Host, r.Content, r.TTL, r.AddPTR, timeNow, r.ZoneUUID, r.Tags)
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

	return nil
}

// Find retrieves a record with the given UUID from the database.
//
// Returns an error if the retrieval fails.
func (r *Record) Find() error {
	query := "SELECT type, host, content, ttl, add_ptr, created_at, modified_at, deleted_at, zone_uuid, staging, tags FROM bind_dns.records WHERE uuid = $1 AND (deleted_at = 0 OR staging = TRUE)"
	row := db.QueryRow(query, r.UUID)
	err := row.Scan(&r.Type, &r.Host, &r.Content, &r.TTL, &r.AddPTR, &r.CreatedAt, &r.ModifiedAt, &r.DeletedAt, &r.ZoneUUID, &r.Staging, &r.Tags)
	if err != nil {
		return err
	}
	return nil
}

// Update updates an existing record in the database.
//
// Returns an error if the update fails.
func (r *Record) Update() error {
	query := "UPDATE bind_dns.records SET type = $1, host = $2, content = $3, ttl = $4, add_ptr = $5, created_at = $6, modified_at = $7, staging = TRUE, tags = $8 WHERE uuid = $8"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	r.ModifiedAt = uint64(time.Now().Unix())

	// check if record is new
	if r.Staging && r.CreatedAt == r.ModifiedAt {
		r.CreatedAt = r.ModifiedAt
	}

	result, err := stmt.Exec(r.Type, r.Host, r.Content, r.TTL, r.AddPTR, r.CreatedAt, r.ModifiedAt, r.Tags, r.UUID)
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

	return nil
}

// Delete deletes a record from the database.
//
// Returns an error if the deletion fails.
func (r *Record) Delete() error {
	query := "UPDATE bind_dns.records SET deleted_at = $1, staging = TRUE WHERE uuid = $2"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(time.Now().Unix(), r.UUID)
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

	return nil
}

// Commit commits the changes to the database.
// Sets all records to staging = FALSE
//
// Returns an error if the commit fails.
func (r *Record) CommitAll() error {
	//Check for any rows to commit
	query := "SELECT COUNT(*) FROM bind_dns.records WHERE staging = TRUE"
	row := db.QueryRow(query)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}

	// Apply changes
	query = "UPDATE bind_dns.records SET staging = FALSE WHERE staging = TRUE"
	result, err := db.Exec(query)
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
func (r *Record) GetStaging() ([]Record, error) {
	query := "SELECT uuid, type, host, content, ttl, add_ptr, created_at, modified_at, deleted_at, zone_uuid, tags FROM bind_dns.records WHERE staging = TRUE"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		err := rows.Scan(&record.UUID, &record.Type, &record.Host, &record.Content, &record.TTL, &record.AddPTR, &record.CreatedAt, &record.ModifiedAt, &record.DeletedAt, &record.ZoneUUID, &record.Tags)
		if err != nil {
			return nil, err
		}
		record.Staging = true
		records = append(records, record)
	}
	return records, nil
}
