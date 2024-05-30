package rdb

import (
	"database/sql"
)

type Config struct {
	db          *sql.DB
	ConfigKey   string
	ConfigValue string
}

func (db *DB) Query(keys []string, table string) (*sql.Rows, error) {
	query := "SELECT"

	for _, key := range keys {
		query += " " + key
	}

	query += " FROM " + table
	rows, err := db.db.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rows, nil
}

func (c *Config) GetAll() ([]Config, error) {
	rows, err := c.db.Query("SELECT config_key, config_value FROM bind_dns.configs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []Config
	for rows.Next() {
		var config Config
		err := rows.Scan(&config.ConfigKey, &config.ConfigValue)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

// Get retrieves the value of a configuration key from the database.
//
// Parameters:
// - key: the key of the configuration to retrieve.
//
// Returns:
// - string: the value of the configuration key.
// - error: an error if the retrieval fails.
func (c *Config) Get(key string) (string, error) {
	row := c.db.QueryRow("SELECT config_value FROM bind_dns.configs WHERE config_key = ?", key)
	var configValue string
	err := row.Scan(&configValue)
	if err != nil {
		return "", err
	}
	return configValue, nil
}
