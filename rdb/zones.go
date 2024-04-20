package rdb

import (
	"database/sql"
)

type Zone struct {
	db           *sql.DB
	UUID         string
	Name         string
	LastModified int
}

// Get retrieves all zones from the database.
//
// Returns:
//   - []Zone: A slice of Zone structs representing the retrieved zones.
//   - error: An error if the retrieval fails.
func (z *Zone) Get() ([]Zone, error) {
	rows, err := z.db.Query("SELECT uuid, name, last_modified FROM bind_dns.zones")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []Zone
	for rows.Next() {
		var zone Zone
		err := rows.Scan(&zone.UUID, &zone.Name, &zone.LastModified)
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
	query := "INSERT INTO bind_dns.zones (uuid, name, last_modified) VALUES ($1, $2, $3)"
	stmt, err := z.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(newZone.UUID, newZone.Name, newZone.LastModified)
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
	query := "DELETE FROM bind_dns.zones WHERE uuid = $1"
	stmt, err := z.db.Prepare(query)
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
	query := "SELECT uuid, name, last_modified FROM bind_dns.zones WHERE uuid = $1"
	row := z.db.QueryRow(query, zoneUUID)
	err := row.Scan(&zone.UUID, &zone.Name, &zone.LastModified)
	if err != nil {
		return zone, err
	}
	return zone, nil
}
