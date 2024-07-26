package rdb

import "database/sql"

type Tag string

func (t Tag) String() string {
	return string(t)
}

func (t Tag) GetRecord(recordUUID string) ([]string, error) {
	rows, err := db.Query("SELECT tag FROM bind_dns.tags WHERE record_uuid::text = $1", recordUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag Tag
		err := rows.Scan(&tag)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag.String())
	}

	return tags, nil
}

func (t Tag) GetZone(zoneUUID string) ([]string, error) {
	rows, err := db.Query("SELECT tag FROM bind_dns.tags WHERE zone_uuid::text = $1", zoneUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag Tag
		err := rows.Scan(&tag)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag.String())
	}

	return tags, nil
}

func (t Tag) CreateRecord(recordUUID string) error {
	query := "INSERT INTO bind_dns.tags (record_uuid, tag) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	_, err := db.Exec(query, recordUUID, t.String())
	if err != nil {
		return err
	}
	return nil
}

func (t Tag) CreateZone(zoneUUID string) error {
	query := "INSERT INTO bind_dns.tags (zone_uuid, tag) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	_, err := db.Exec(query, zoneUUID, t.String())
	if err != nil {
		return err
	}
	return nil
}

func (t Tag) DeleteRecord(recordUUID string) error {
	var result sql.Result
	if t.String() == "" {
		// delete all tags
		query := "DELETE FROM bind_dns.tags WHERE record_uuid = $1"
		stmt, err := db.Prepare(query)
		if err != nil {
			return err
		}
		defer stmt.Close()
		result, err = stmt.Exec(recordUUID)
		if err != nil {
			return err
		}
	} else {
		// delete specific tag
		query := "DELETE FROM bind_dns.tags WHERE record_uuid = $1 AND tag = $2"
		stmt, err := db.Prepare(query)
		if err != nil {
			return err
		}
		defer stmt.Close()
		result, err = stmt.Exec(recordUUID, t.String())
		if err != nil {
			return err
		}
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

// DeleteZone deletes either all tags or a specific tag for a given zone UUID.
//
// Parameters:
//   - t: The Tag object to delete.
//   - zoneUUID: The UUID of the zone to delete tags from.
//
// Returns:
//   - error: An error if the deletion fails.
func (t Tag) DeleteZone(zoneUUID string) error {
	var result sql.Result
	if t.String() == "" {
		// delete all tags
		query := "DELETE FROM bind_dns.tags WHERE zone_uuid::text = $1"
		stmt, err := db.Prepare(query)
		if err != nil {
			return err
		}
		defer stmt.Close()
		result, err = stmt.Exec(zoneUUID)
		if err != nil {
			return err
		}
	} else {
		// delete specific tag
		query := "DELETE FROM bind_dns.tags WHERE zone_uuid::text = $1 AND tag = $2"
		stmt, err := db.Prepare(query)
		if err != nil {
			return err
		}
		defer stmt.Close()
		result, err = stmt.Exec(zoneUUID, t.String())
		if err != nil {
			return err
		}
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
