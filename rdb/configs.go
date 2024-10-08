package rdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Config struct {
	ConfigKey   string
	ConfigValue string
	CreatedAt   time.Time
	ModifiedAt  time.Time
	DeletedAt   sql.NullTime
	Staging     bool
}

// Get retrieves all configurations from the bind_dns.configs table where staging is true or deleted_at is null.
//
// It returns a slice of Config structs and an error if any.
func (c *Config) Get(ctx context.Context) ([]Config, error) {

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, "SELECT config_key, config_value, created_at, modified_at, staging FROM bind_dns.configs WHERE staging = TRUE OR deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	configs := []Config{}
	for rows.Next() {
		var config Config
		err := rows.Scan(&config.ConfigKey, &config.ConfigValue, &config.CreatedAt, &config.ModifiedAt, &config.Staging)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

// Find retrieves a list of Config objects from the database based on the provided ConfigKey.
//
// The function returns a slice of Config objects and an error.
func (c *Config) Find(ctx context.Context) ([]Config, error) {

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, "SELECT config_key, config_value, created_at, modified_at, deleted_at, staging FROM bind_dns.configs WHERE config_key = $1 AND (deleted_at IS NULL OR staging = TRUE)", c.ConfigKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	configs := []Config{}
	for rows.Next() {
		var config Config
		err := rows.Scan(&config.ConfigKey, &config.ConfigValue, &config.CreatedAt, &config.ModifiedAt, &config.DeletedAt, &config.Staging)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

// func (c *Config) Find() error {
// 	rows := db.QueryRow("SELECT config_value, created_at, modified_at, deleted_at, staging FROM bind_dns.configs WHERE config_key = ?", c.ConfigKey)
// 	err := rows.Scan(&c.ConfigValue, &c.CreatedAt, &c.ModifiedAt, &c.DeletedAt, &c.Staging)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// Delete removes a config from the database.
func (c *Config) Delete(ctx context.Context) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, "UPDATE bind_dns.configs SET deleted_at = NOW(), staging = $1 WHERE config_key = $2 and deleted_at IS NULL", c.Staging, c.ConfigKey)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}

func (c *Config) Create(ctx context.Context) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// make sure config doesn't already exist
	configs, err := c.Find(ctx)
	if err != nil {
		return err
	}
	for _, config := range configs {
		if config.ConfigValue == c.ConfigValue {
			return fmt.Errorf("config %s=%s already exists", config.ConfigKey, c.ConfigValue)
		}
	}

	query := "INSERT INTO bind_dns.configs (config_key, config_value, created_at, modified_at, staging) VALUES ($1, $2, $3, $4, $5)"
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	timeNow := time.Now()
	c.CreatedAt = timeNow
	c.ModifiedAt = timeNow
	result, err := stmt.ExecContext(ctx, c.ConfigKey, c.ConfigValue, timeNow, timeNow, c.Staging)
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
	return tx.Commit()
}

func (c *Config) Update(ctx context.Context, value string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// make sure config with value doesn't already exist
	configs, err := c.Find(ctx)
	if err != nil {
		return err
	}
	for _, config := range configs {
		if config.ConfigValue == c.ConfigValue {
			goto FOUND
		}
	}
	return fmt.Errorf("could not find %s=%s", c.ConfigKey, c.ConfigValue)

FOUND:
	query := "UPDATE bind_dns.configs SET config_value = $1, modified_at = $2, staging = $3 WHERE config_key = $4 and config_value = $5"
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	timeNow := time.Now()
	c.ModifiedAt = timeNow
	result, err := stmt.ExecContext(ctx, value, timeNow, c.Staging, c.ConfigKey, c.ConfigValue)
	if err != nil {
		return err
	}

	// Check if any rows were updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}
