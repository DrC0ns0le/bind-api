package rdb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Tag string

var ErrTagNotFound error = fmt.Errorf("tags not found")

func (t Tag) String() string {
	return string(t)
}

func (t Tag) GetRecord(ctx context.Context, recordUUID string) ([]string, error) {
	rows, err := db.Query(ctx, "SELECT tag FROM bind_dns.tags WHERE record_uuid::text = $1", recordUUID)
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
	rows, err := db.Query(ctx, "SELECT tag FROM bind_dns.tags WHERE zone_uuid::text = $1", zoneUUID)
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
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := "INSERT INTO bind_dns.tags (record_uuid, tag) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	commandTag, err := tx.Exec(ctx, query, recordUUID, t.String())
	if err != nil {
		return fmt.Errorf("failed to insert tag: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrNotCreated
	}

	return tx.Commit(ctx)
}

func (t Tag) CreateZone(ctx context.Context, zoneUUID string) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := "INSERT INTO bind_dns.tags (zone_uuid, tag) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	commandTag, err := tx.Exec(ctx, query, zoneUUID, t.String())
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrNotCreated
	}

	return tx.Commit(ctx)
}

func (t Tag) DeleteRecord(ctx context.Context, recordUUID string) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// check if record exists
	var count int
	query := "SELECT COUNT(*) FROM bind_dns.tags WHERE record_uuid::text = $1"
	err = tx.QueryRow(ctx, query, recordUUID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check record existence: %w", err)
	}
	if count == 0 {
		return ErrTagNotFound
	}

	var commandTag pgconn.CommandTag
	if t.String() == "" {
		// delete all tags
		query := "DELETE FROM bind_dns.tags WHERE record_uuid::text = $1"
		commandTag, err = tx.Exec(ctx, query, recordUUID)
	} else {
		// delete specific tag
		query := "DELETE FROM bind_dns.tags WHERE record_uuid::text = $1 AND tag = $2"
		commandTag, err = tx.Exec(ctx, query, recordUUID, t.String())
	}

	if err != nil {
		return fmt.Errorf("failed to delete tag(s): %w", err)
	}

	// Check affected rows
	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return tx.Commit(ctx)
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
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// check if zone exists
	var count int
	query := "SELECT COUNT(*) FROM bind_dns.tags WHERE zone_uuid::text = $1"
	err = tx.QueryRow(ctx, query, zoneUUID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check zone existence: %w", err)
	}
	if count == 0 {
		return ErrTagNotFound
	}

	var commandTag pgconn.CommandTag
	if t.String() == "" {
		// delete all tags
		query := "DELETE FROM bind_dns.tags WHERE zone_uuid::text = $1"
		commandTag, err = tx.Exec(ctx, query, zoneUUID)
	} else {
		// delete specific tag
		query := "DELETE FROM bind_dns.tags WHERE zone_uuid::text = $1 AND tag = $2"
		commandTag, err = tx.Exec(ctx, query, zoneUUID, t.String())
	}

	if err != nil {
		return fmt.Errorf("failed to delete tag(s): %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return tx.Commit(ctx)
}
