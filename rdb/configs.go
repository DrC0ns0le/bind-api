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

func (c *Config) Get() ([]Config, error) {
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
