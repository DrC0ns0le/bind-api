package rdb

import (
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
func (c *Config) Get() ([]Config, error) {
	rows, err := db.Query("SELECT config_key, config_value, created_at, modified_at, staging FROM bind_dns.configs WHERE staging = TRUE OR deleted_at IS NULL")
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
func (c *Config) Find() ([]Config, error) {
	rows, err := db.Query("SELECT config_key, config_value, created_at, modified_at, deleted_at, staging FROM bind_dns.configs WHERE config_key = $1 AND (deleted_at IS NULL OR staging = TRUE)", c.ConfigKey)
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
func (c *Config) Delete() error {
	result, err := db.Exec("UPDATE bind_dns.configs SET deleted_at = NOW(), staging = $1 WHERE config_key = $2 and deleted_at IS NULL", c.Staging, c.ConfigKey)
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
	return nil
}

func (c *Config) Create() error {
	// make sure config doesn't already exist
	configs, err := c.Find()
	if err != nil {
		return err
	}
	for _, config := range configs {
		if config.ConfigValue == c.ConfigValue {
			return fmt.Errorf("config %s=%s already exists", config.ConfigKey, c.ConfigValue)
		}
	}

	query := "INSERT INTO bind_dns.configs (config_key, config_value, created_at, modified_at, staging) VALUES ($1, $2, $3, $4, $5)"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	timeNow := time.Now()
	c.CreatedAt = timeNow
	c.ModifiedAt = timeNow
	result, err := stmt.Exec(c.ConfigKey, c.ConfigValue, timeNow, timeNow, c.Staging)
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

func (c *Config) Update(value string) error {
	// make sure config with value doesn't already exist
	configs, err := c.Find()
	if err != nil {
		return err
	}
	for _, config := range configs {
		if config.ConfigValue == value {
			goto FOUND
		}
	}
	return fmt.Errorf("could not find %s=%s", c.ConfigKey, c.ConfigValue)

FOUND:
	query := "UPDATE bind_dns.configs SET config_value = $1, modified_at = $2, staging = $3 WHERE config_key = $4 and config_value = $5"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(value, time.Now(), c.Staging, c.ConfigKey, c.ConfigValue)
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
	return nil
}
