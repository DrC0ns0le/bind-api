package rdb

import (
	"database/sql"
	"time"
)

type Zone struct {
	UUID       string // Zone UUID
	Name       string // Zone name
	CreatedAt  uint64 // Zone creation time
	ModifiedAt uint64 // Zone modification time
	DeletedAt  uint64 // Zone deletion time
	Staging    bool   // Zone staging status
	PrimaryNS  string // Zone primary NS
	AdminEmail string // Zone admin email
	Refresh    uint16 // Zone refresh interval
	Retry      uint16 // Zone retry interval
	Expire     uint32 // Zone expire interval
	Minimum    uint16 // Zone minimum TTL
	TTL        uint16 // Zone TTL
	Tags       string // Zone tags
}

// Get retrieves all zones from the database.
//
// Returns:
//   - []Zone: A slice of Zone structs representing the retrieved zones.
//   - error: An error if the retrieval fails.
func (z *Zone) Get() ([]Zone, error) {
	rows, err := db.Query("SELECT uuid, name, created_at, modified_at, deleted_at, primary_ns, admin_email, refresh, retry, expire, minimum, ttl,staging FROM bind_dns.zones WHERE deleted_at = 0 OR staging = TRUE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []Zone
	for rows.Next() {
		var zone Zone
		err := rows.Scan(&zone.UUID, &zone.Name, &zone.CreatedAt, &zone.ModifiedAt, &zone.DeletedAt, &zone.PrimaryNS, &zone.AdminEmail, &zone.Refresh, &zone.Retry, &zone.Expire, &zone.Minimum, &zone.TTL, &zone.Staging)
		if err != nil {
			return nil, err
		}
		zones = append(zones, zone)
	}

	return zones, nil
}

// Create inserts a new zone into the database.
//
// Returns an error if the insertion fails.
func (z *Zone) Create() error {
	query := "INSERT INTO bind_dns.zones (uuid, name, created_at, modified_at, deleted_at, primary_ns, admin_email, refresh, retry, expire, minimum, staging) VALUES ($1, $2, $3, $3, 0, $4, $5, $6, $7, $8, $9, TRUE)"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	timeNow := time.Now().Unix()
	z.CreatedAt = uint64(timeNow)
	z.ModifiedAt = uint64(timeNow)
	_, err = stmt.Exec(z.UUID, z.Name, timeNow, z.PrimaryNS, z.AdminEmail, z.Refresh, z.Retry, z.Expire, z.Minimum)
	if err != nil {
		return err
	}

	return nil
}

// Update marks a zone as staging in the database.
//
// Returns an error if the update fails.
func (z *Zone) Update() error {
	query := "UPDATE bind_dns.zones SET name = $1, primary_ns = $2, admin_email = $3, refresh = $4, retry = $5, expire = $6, minimum = $7, staging = TRUE WHERE uuid = $8"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(z.Name, z.PrimaryNS, z.AdminEmail, z.Refresh, z.Retry, z.Expire, z.Minimum, z.UUID)
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

// Delete marks a zone as deleted in the database.
//
// Returns an error if the deletion fails.
func (z *Zone) Delete() error {
	query := "UPDATE bind_dns.zones SET deleted_at = $1, staging = TRUE WHERE uuid = $2"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(time.Now().Unix(), z.UUID)
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

// Find retrieves a zone from the database.
//
// Returns an error if the retrieval fails.
func (z *Zone) Find() error {
	query := "SELECT uuid, name, created_at, modified_at, deleted_at, primary_ns, admin_email, refresh, retry, expire, minimum, staging FROM bind_dns.zones WHERE uuid = $1 AND (deleted_at = 0 OR (deleted_at != 0 AND staging = TRUE))"
	row := db.QueryRow(query, z.UUID)
	err := row.Scan(&z.UUID, &z.Name, &z.CreatedAt, &z.ModifiedAt, &z.DeletedAt, &z.PrimaryNS, &z.AdminEmail, &z.Refresh, &z.Retry, &z.Expire, &z.Minimum, &z.Staging)
	if err != nil {
		return err
	}
	return nil
}

// GetStaging retrieves all zones in staging from the database.
//
// Returns:
//   - []Zone: A slice of Zone structs representing the retrieved zones.
//   - error: An error if the retrieval fails.
func (z *Zone) GetStaging() ([]Zone, error) {
	rows, err := db.Query("SELECT uuid, name, created_at, modified_at, deleted_at, primary_ns, admin_email, refresh, retry, expire, minimum, staging FROM bind_dns.zones WHERE staging = TRUE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []Zone
	for rows.Next() {
		var zone Zone
		err := rows.Scan(&zone.UUID, &zone.Name, &zone.CreatedAt, &zone.ModifiedAt, &zone.DeletedAt, &zone.PrimaryNS, &zone.AdminEmail, &zone.Refresh, &zone.Retry, &zone.Expire, &zone.Minimum, &zone.Staging)
		if err != nil {
			return nil, err
		}
		zones = append(zones, zone)
	}

	return zones, nil
}
