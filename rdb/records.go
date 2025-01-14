package rdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
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

// Get retrieves a list of records from the database based on the provided criteria.
//
// Parameters:
//   - ctx: The context for the database operation.
//   - page: The page number for pagination (1-indexed). If 0, pagination is disabled.
//   - pageSize: The number of records per page. If 0, pagination is disabled.
//   - searchQuery: A general search query to filter records across multiple fields.
//
// The function applies filters based on the Record struct fields:
//   - ZoneUUID: Mandatory filter for the zone.
//   - Type: Optional filter for record type.
//   - Host: Optional filter for record host.
//   - Content: Optional filter for record content.
//   - Tags: Optional filter for record tags.
//
// It also includes non-deleted records or records in staging.
//
// Returns:
//   - []Record: A slice of Record structs matching the criteria.
//   - int: The total count of records matching the criteria (before pagination).
//   - error: An error if any occurred during the operation.
//
// The function uses case-insensitive partial matching (ILIKE) for string fields
// and supports pagination. It orders results by creation date in descending order.
func (r *Record) Get(ctx context.Context, page, pageSize int, searchQuery string) ([]Record, int, error) {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Build WHERE conditions and arguments
	whereConditions := []string{"z.uuid::text = $1"}
	args := []interface{}{r.ZoneUUID}
	argCount := 1

	// Add specific field filters
	if r.Type != "" {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("r.type ILIKE $%d", argCount))
		args = append(args, "%"+r.Type+"%")
	}
	if r.Host != "" {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("r.host ILIKE $%d", argCount))
		args = append(args, "%"+r.Host+"%")
	}
	if r.Content != "" {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("r.content ILIKE $%d", argCount))
		args = append(args, "%"+r.Content+"%")
	}

	// Add tags filter if provided and not empty
	if len(r.Tags) > 0 && r.Tags[0] != "" {
		tagPlaceholders := make([]string, len(r.Tags))
		for i := range r.Tags {
			argCount++
			tagPlaceholders[i] = fmt.Sprintf("$%d", argCount)
			args = append(args, r.Tags[i])
		}
		whereConditions = append(whereConditions, fmt.Sprintf(`
            EXISTS (
                SELECT 1 FROM bind_dns.tags t
                WHERE t.record_uuid = r.uuid
                AND t.tag IN (%s)
            )`, strings.Join(tagPlaceholders, ",")))
	}

	// Add general search query if provided
	if searchQuery != "" {
		argCount++
		searchArg := "%" + searchQuery + "%"

		searchConditions := []string{
			fmt.Sprintf("r.type ILIKE $%d", argCount),
			fmt.Sprintf("r.host ILIKE $%d", argCount),
			fmt.Sprintf("r.content ILIKE $%d", argCount),
			fmt.Sprintf(`
                EXISTS (
                    SELECT 1 FROM bind_dns.tags t
                    WHERE t.record_uuid = r.uuid
                    AND t.tag ILIKE $%d
                )`, argCount),
		}

		// Add the OR conditions as a grouped condition
		whereConditions = append(whereConditions,
			fmt.Sprintf("(%s)", strings.Join(searchConditions, " OR ")))

		args = append(args, searchArg)
	}

	// Add deletion/staging condition
	whereConditions = append(whereConditions, "(r.deleted_at IS NULL OR r.staging = TRUE)")

	// Combine all conditions
	whereClause := strings.Join(whereConditions, " AND ")

	// Get total count first
	countQuery := fmt.Sprintf(`
        SELECT COUNT(*)
        FROM bind_dns.records AS r
        JOIN bind_dns.zones AS z ON r.zone_uuid = z.uuid
        WHERE %s`, whereClause)

	var total int
	err = tx.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build the main query
	query := fmt.Sprintf(`
        SELECT r.uuid, r.type, r.host, r.content, r.ttl, r.add_ptr,
        r.created_at, r.modified_at, r.deleted_at, r.staging
        FROM bind_dns.records AS r
        JOIN bind_dns.zones AS z ON r.zone_uuid = z.uuid
        WHERE %s
        ORDER BY r.created_at DESC`, whereClause)

	// Add pagination if requested
	if page > 0 && pageSize > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, pageSize)
		argCount++
		offset := (page - 1) * pageSize
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	// Execute the query
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		if err := rows.Scan(
			&record.UUID,
			&record.Type,
			&record.Host,
			&record.Content,
			&record.TTL,
			&record.AddPTR,
			&record.CreatedAt,
			&record.ModifiedAt,
			&record.DeletedAt,
			&record.Staging,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan record: %w", err)
		}
		record.ZoneUUID = r.ZoneUUID

		// Get tags for the record
		tags, err := new(Tag).GetRecord(ctx, record.UUID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get tags for record %s: %w", record.UUID, err)
		}
		record.Tags = tags
		records = append(records, record)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error during row iteration: %w", err)
	}

	return records, total, nil
}

// GetAll retrieves all records from the database.
//
// It returns a slice of Record and an error if any.
func (r *Record) GetAll(ctx context.Context) ([]Record, error) {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
        SELECT uuid, type, host, content, ttl, add_ptr, created_at, modified_at, deleted_at, zone_uuid, staging 
        FROM bind_dns.records 
        WHERE deleted_at IS NULL OR staging = TRUE`)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		if err := rows.Scan(
			&record.UUID,
			&record.Type,
			&record.Host,
			&record.Content,
			&record.TTL,
			&record.AddPTR,
			&record.CreatedAt,
			&record.ModifiedAt,
			&record.DeletedAt,
			&record.ZoneUUID,
			&record.Staging,
		); err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}

		// Get tags for the record
		tags, err := new(Tag).GetZone(ctx, record.ZoneUUID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tags for zone %s: %w", record.ZoneUUID, err)
		}
		record.Tags = tags

		records = append(records, record)
	}

	// Check for errors that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return records, nil
}

// Create inserts a new record into the database.
//
// Returns an error if the insertion fails.
func (r *Record) Create(ctx context.Context) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	timeNow := time.Now()
	r.CreatedAt = timeNow
	r.ModifiedAt = timeNow

	// Insert the record
	commandTag, err := tx.Exec(ctx, `
        INSERT INTO bind_dns.records (
            uuid, type, host, content, ttl, add_ptr, created_at, modified_at, zone_uuid, staging
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $7, $8, TRUE
        )`,
		r.UUID, r.Type, r.Host, r.Content, r.TTL, r.AddPTR,
		timeNow, r.ZoneUUID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert record: %w", err)
	}

	// Check if the insert was successful
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows were inserted")
	}

	// Add tags if any exist
	if len(r.Tags) > 0 {
		if err := new(Tag).CreateRecord(ctx, r.UUID); err != nil {
			return fmt.Errorf("failed to create record tags: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Find retrieves a record with the given UUID from the database.
//
// Returns an error if the retrieval fails.
func (r *Record) Find(ctx context.Context) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
        SELECT type, host, content, ttl, add_ptr, created_at, modified_at, deleted_at, zone_uuid, staging 
        FROM bind_dns.records 
        WHERE uuid::text = $1 
        AND (deleted_at IS NULL OR staging = TRUE)`,
		r.UUID,
	).Scan(
		&r.Type, &r.Host, &r.Content, &r.TTL, &r.AddPTR,
		&r.CreatedAt, &r.ModifiedAt, &r.DeletedAt,
		&r.ZoneUUID, &r.Staging,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("record not found: %w", err)
		}
		return fmt.Errorf("failed to query record: %w", err)
	}

	// Get tags for the record
	tags, err := new(Tag).GetZone(ctx, r.ZoneUUID)
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}
	r.Tags = tags

	return nil
}

// Update updates an existing record in the database.
//
// Returns an error if the update fails.
func (r *Record) Update(ctx context.Context) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update modified time based on staging status
	if r.Staging && r.CreatedAt == r.ModifiedAt {
		r.ModifiedAt = time.Now()
		r.CreatedAt = r.ModifiedAt
	} else {
		r.ModifiedAt = time.Now()
	}

	// Update the record
	commandTag, err := tx.Exec(ctx, `
        UPDATE bind_dns.records 
        SET type = $1, host = $2, content = $3, ttl = $4, add_ptr = $5, created_at = $6, modified_at = $7, staging = TRUE 
        WHERE uuid::text = $8`,
		r.Type, r.Host, r.Content, r.TTL, r.AddPTR,
		r.CreatedAt, r.ModifiedAt, r.UUID,
	)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("record not found: %w", pgx.ErrNoRows)
	}

	// Handle tags
	if err := new(Tag).DeleteRecord(ctx, r.UUID); err != nil && err != ErrTagNotFound {
		return fmt.Errorf("failed to delete existing tags: %w", err)
	}

	if len(r.Tags) > 0 {
		if err := new(Tag).CreateRecord(ctx, r.UUID); err != nil {
			return fmt.Errorf("failed to create new tags: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Delete marks a record as deleted in the database.
//
// Returns an error if the deletion fails.
func (r *Record) Delete(ctx context.Context) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	commandTag, err := tx.Exec(ctx, `
        UPDATE bind_dns.records 
        SET deleted_at = $1, staging = TRUE 
        WHERE uuid::text = $2`,
		time.Now(), r.UUID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark record as deleted: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("record not found: %w", pgx.ErrNoRows)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CommitAll commits all staged changes to the database.
// Sets all records to staging = FALSE
//
// Returns an error if the commit fails.
func (r *Record) CommitAll(ctx context.Context) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check for any rows to commit
	var count int
	err = tx.QueryRow(ctx, `
        SELECT COUNT(*) 
        FROM bind_dns.records 
        WHERE staging = TRUE`,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count staged records: %w", err)
	}

	if count == 0 {
		return nil
	}

	// Apply changes
	commandTag, err := tx.Exec(ctx, `
        UPDATE bind_dns.records 
        SET staging = FALSE 
        WHERE staging = TRUE`,
	)
	if err != nil {
		return fmt.Errorf("failed to commit staged records: %w", err)
	}

	if commandTag.RowsAffected() != int64(count) {
		return fmt.Errorf(
			"commit affected %d rows, expected %d",
			commandTag.RowsAffected(), count,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetStaging retrieves all records in the staging area.
//
// Returns:
//   - []Record: A slice of Record structs representing the retrieved records.
//   - error: An error if the retrieval fails.
func (r *Record) GetStaging(ctx context.Context) ([]Record, error) {
	rows, err := db.Query(ctx, `
        SELECT uuid, type, host, content, ttl, add_ptr, created_at, modified_at, deleted_at, zone_uuid 
        FROM bind_dns.records 
        WHERE staging = TRUE`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query staging records: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		err := rows.Scan(
			&record.UUID, &record.Type, &record.Host,
			&record.Content, &record.TTL, &record.AddPTR,
			&record.CreatedAt, &record.ModifiedAt,
			&record.DeletedAt, &record.ZoneUUID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}

		record.Staging = true

		tags, err := new(Tag).GetZone(ctx, record.ZoneUUID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tags for record %s: %w", record.UUID, err)
		}
		record.Tags = tags

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return records, nil
}
