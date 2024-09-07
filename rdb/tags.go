package rdb

import (
	"context"
	"database/sql"
)

type Tag string

func (t Tag) String() string {
	return string(t)
}

func (t Tag) GetRecord(ctx context.Context, recordUUID string) ([]string, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, "SELECT tag FROM bind_dns.tags WHERE record_uuid::text = $1", recordUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]string, 0)
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

func (t Tag) GetZone(ctx context.Context, zoneUUID string) ([]string, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, "SELECT tag FROM bind_dns.tags WHERE zone_uuid::text = $1", zoneUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]string, 0)
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

func (t Tag) CreateRecord(ctx context.Context, recordUUID string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := "INSERT INTO bind_dns.tags (record_uuid, tag) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, recordUUID, t.String())
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (t Tag) CreateZone(ctx context.Context, zoneUUID string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := "INSERT INTO bind_dns.tags (zone_uuid, tag) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, zoneUUID, t.String())
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (t Tag) DeleteRecord(ctx context.Context, recordUUID string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var result sql.Result
	if t.String() == "" {
		// delete all tags
		query := "DELETE FROM bind_dns.tags WHERE record_uuid = $1"
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return err
		}
		defer stmt.Close()
		result, err = stmt.ExecContext(ctx, recordUUID)
		if err != nil {
			return err
		}
	} else {
		// delete specific tag
		query := "DELETE FROM bind_dns.tags WHERE record_uuid = $1 AND tag = $2"
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return err
		}
		defer stmt.Close()
		result, err = stmt.ExecContext(ctx, recordUUID, t.String())
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
	return tx.Commit()
}

// DeleteZone deletes either all tags or a specific tag for a given zone UUID.
//
// Parameters:
//   - t: The Tag object to delete.
//   - zoneUUID: The UUID of the zone to delete tags from.
//
// Returns:
//   - error: An error if the deletion fails.
func (t Tag) DeleteZone(ctx context.Context, zoneUUID string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var result sql.Result
	if t.String() == "" {
		// delete all tags
		query := "DELETE FROM bind_dns.tags WHERE zone_uuid::text = $1"
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return err
		}
		defer stmt.Close()
		result, err = stmt.ExecContext(ctx, zoneUUID)
		if err != nil {
			return err
		}
	} else {
		// delete specific tag
		query := "DELETE FROM bind_dns.tags WHERE zone_uuid::text = $1 AND tag = $2"
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return err
		}
		defer stmt.Close()
		result, err = stmt.ExecContext(ctx, zoneUUID, t.String())
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
	return tx.Commit()
}
