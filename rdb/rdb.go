package rdb

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func Init(config DBConfig) {

	// Connect to the database
	if err := connect(config.Host, config.Port, config.User, config.Password, config.DBName, "disable"); err != nil {
		log.Fatal(err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to the database successfully.\n")
}

// Connect establishes a connection to the database
func connect(host string, port int, user, password, dbname string, sslmode string) error {
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	postgres, err := sql.Open("postgres", dbinfo)
	if err != nil {
		return err
	}
	db = postgres
	return nil
}
