package rdb

import "database/sql"

type Record struct {
	db           *sql.DB
	UUID         string
	Type         string
	Host         string
	Content      string
	TTL          int
	LastModified int
	ZoneUUID     string
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
	rows, err := r.db.Query("SELECT r.uuid, r.type, r.host, r.content, r.ttl, r.last_modified FROM bind_dns.records AS r JOIN bind_dns.zones AS z ON r.zone_uuid = z.uuid WHERE z.uuid = $1;", zoneUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		err := rows.Scan(&record.UUID, &record.Type, &record.Host, &record.Content, &record.TTL, &record.LastModified)
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
	rows, err := r.db.Query("SELECT uuid, type, host, content, ttl, last_modified, zone_uuid FROM bind_dns.records")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		err := rows.Scan(&record.UUID, &record.Type, &record.Host, &record.Content, &record.TTL, &record.LastModified, &record.ZoneUUID)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

// Create inserts a new record into the bind_dns.zones table.
//
// It takes a newRecord of type Record as a parameter and returns an error.
func (r *Record) Create(newRecord Record) error {
	query := "INSERT INTO bind_dns.zones (uuid, type, host, content, ttl, last_modified, zone_uuid) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(newRecord.UUID, newRecord.Type, newRecord.Host, newRecord.Content, newRecord.TTL, newRecord.LastModified, newRecord.ZoneUUID)
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

// Update updates a record in the bind_dns.records table with the provided newRecord.
//
// Parameters:
// - newRecord: The new record to update. All fields in newRecord must be populated, except for the ID field.
//
// Returns:
// - error: An error if the update fails.
func (r *Record) Update(newRecord Record) error {
	query := "UPDATE bind_dns.records SET type = $1, host = $2, content = $3, ttl = $4, last_modified = $5 WHERE uuid = $6"
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(newRecord.Type, newRecord.Host, newRecord.Content, newRecord.TTL, newRecord.LastModified, newRecord.UUID)

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

// Delete deletes a record with the given UUID from the database.
//
// Parameters:
//   - uuid: the unique identifier of the record to be deleted
//
// Return type: error
func (r *Record) Delete(uuid string) error {
	query := "DELETE FROM bind_dns.records WHERE uuid = $1"
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(uuid)
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

func (z *Record) Select(recordUUID string) (Record, error) {
	var record Record
	query := "SELECT uuid, type, host, content, ttl, last_modified, zone_uuid FROM bind_dns.records WHERE uuid = $1"
	row := z.db.QueryRow(query, recordUUID)
	err := row.Scan(&record.UUID, &record.Type, &record.Host, &record.Content, &record.TTL, &record.LastModified, &record.ZoneUUID)
	if err != nil {
		return record, err
	}
	return record, nil
}
