package rdb

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type DB struct {
	db *sql.DB
}

// Connect establishes a connection to the database
func (db *DB) Connect(host string, port int, user, password, dbname string, sslmode string) error {
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	postgres, err := sql.Open("postgres", dbinfo)
	if err != nil {
		return err
	}
	db.db = postgres
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.db != nil {
		return db.db.Close()
	}
	return nil
}

func (db *DB) Ping() error {
	return db.db.Ping()
}
