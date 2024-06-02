package rdb

import (
	"database/sql"
	"time"
)

type Zone struct {
	db         *sql.DB
	UUID       string
	Name       string
	CreatedAt  uint64
	ModifiedAt uint64
	DeletedAt  uint64
	Staging    bool
	PrimaryNS  string
	AdminEmail string
	Refresh    uint16
	Retry      uint16
	Expire     uint32
	Minimum    uint16
	TTL        uint16
}

// Get retrieves all zones from the database.
//
// Returns:
//   - []Zone: A slice of Zone structs representing the retrieved zones.
//   - error: An error if the retrieval fails.
func (z *Zone) Get() ([]Zone, error) {
	rows, err := z.db.Query("SELECT uuid, name, created_at, modified_at, deleted_at, primary_ns, admin_email, refresh, retry, expire, minimum, ttl,staging FROM bind_dns.zones WHERE deleted_at = 0 OR staging = TRUE")
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
// Parameters:
//   - newZone: A Zone struct representing the new zone to be inserted.
//
// Returns:
//   - error: An error if the insertion fails.
func (z *Zone) Create(newZone Zone) error {
	query := "INSERT INTO bind_dns.zones (uuid, name, created_at, modified_at, deleted_at, primary_ns, admin_email, refresh, retry, expire, minimum, staging) VALUES ($1, $2, $3, $3, 0, $4, $5, $6, $7, $8, $9, TRUE)"
	stmt, err := z.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(newZone.UUID, newZone.Name, time.Now().Unix(), newZone.PrimaryNS, newZone.AdminEmail, newZone.Refresh, newZone.Retry, newZone.Expire, newZone.Minimum)
	if err != nil {
		return err
	}

	return nil
}

// Update zones in the database with the provided newZone.
//
// Parameters:
// - newZone: The new zone to update. All fields in newZone must be populated, except for the UUID field.
//
// Returns:
// - error: An error if the update fails.
func (z *Zone) Update(newZone Zone) error {
	query := "UPDATE bind_dns.zones SET name = $1, primary_ns = $2, admin_email = $3, refresh = $4, retry = $5, expire = $6, minimum = $7, staging = TRUE WHERE uuid = $8"
	stmt, err := z.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(newZone.Name, newZone.PrimaryNS, newZone.AdminEmail, newZone.Refresh, newZone.Retry, newZone.Expire, newZone.Minimum, newZone.UUID)
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

// Delete deletes a zone from the database based on the provided UUID.
//
// Parameters:
// - uuid: The UUID of the zone to delete.
//
// Returns:
// - error: An error if the deletion fails.
func (z *Zone) Delete(uuid string) error {
	query := "UPDATE bind_dns.zones SET deleted_at = $1, staging = TRUE WHERE uuid = $2"
	stmt, err := z.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(time.Now().Unix(), uuid)
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

// Select retrieves a zone from the database based on the provided zone UUID.
//
// Parameters:
// - zoneUUID: The UUID of the zone to retrieve.
//
// Returns:
// - *Zone: A pointer to the retrieved zone.
// - error: An error if the retrieval fails.
func (z *Zone) Select(zoneUUID string) (Zone, error) {
	var zone Zone
	query := "SELECT uuid, name, created_at, modified_at, deleted_at, primary_ns, admin_email, refresh, retry, expire, minimum, staging FROM bind_dns.zones WHERE uuid = $1 AND (deleted_at = 0 OR (deleted_at != 0 AND staging = TRUE))"
	row := z.db.QueryRow(query, zoneUUID)
	err := row.Scan(&zone.UUID, &zone.Name, &zone.CreatedAt, &zone.ModifiedAt, &zone.DeletedAt, &zone.PrimaryNS, &zone.AdminEmail, &zone.Refresh, &zone.Retry, &zone.Expire, &zone.Minimum, &zone.Staging)
	if err != nil {
		return zone, err
	}
	return zone, nil
}

// GetStaging retrieves all staging zones from the database.
//
// It returns a slice of Zone structs representing the retrieved zones and an error if the retrieval fails.
func (z *Zone) GetStaging() ([]Zone, error) {
	rows, err := z.db.Query("SELECT uuid, name, created_at, modified_at, deleted_at, primary_ns, admin_email, refresh, retry, expire, minimum, staging FROM bind_dns.zones WHERE staging = TRUE")
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
