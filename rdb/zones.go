package rdb

import (
	"database/sql"
	"fmt"
	"time"
)

type Zone struct {
	db         *sql.DB
	UUID       string
	Name       string
	ModifiedAt int
	DeletedAt  int
	Staging    bool
}

// Get retrieves all zones from the database.
//
// Returns:
//   - []Zone: A slice of Zone structs representing the retrieved zones.
//   - error: An error if the retrieval fails.
func (z *Zone) Get() ([]Zone, error) {
	rows, err := z.db.Query("SELECT uuid, name, modified_at, deleted_at, staging FROM bind_dns.zones WHERE deleted_at = 0;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []Zone
	for rows.Next() {
		var zone Zone
		err := rows.Scan(&zone.UUID, &zone.Name, &zone.ModifiedAt, &zone.DeletedAt, &zone.Staging)
		if err != nil {
			return nil, err
		}
		fmt.Println(zone)
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
	query := "INSERT INTO bind_dns.zones (uuid, name, modified_at, deleted_at, staging) VALUES ($1, $2, $3, 0, TRUE)"
	stmt, err := z.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(newZone.UUID, newZone.Name, newZone.ModifiedAt, true)
	if err != nil {
		return err
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
	query := "SELECT uuid, name, modified_at, deleted_at, staging FROM bind_dns.zones WHERE uuid = $1"
	row := z.db.QueryRow(query, zoneUUID)
	err := row.Scan(&zone.UUID, &zone.Name, &zone.ModifiedAt, &zone.DeletedAt, &zone.Staging)
	if err != nil {
		return zone, err
	}
	return zone, nil
}
