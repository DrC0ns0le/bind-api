package rdb

import (
	"log"
)

type BindData struct {
	DB
	Zones   Zone
	Records Record
	Configs Config
}

func (bd *BindData) Init(host string, port int, user, password, dbname string) {

	// Connect to the database
	if err := bd.Connect(host, port, user, password, dbname, "disable"); err != nil {
		log.Fatal(err)
	}

	// Test the connection
	if err := bd.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to the database successfully.\n")

	bd.Configs = Config{db: bd.db}
	bd.Zones = Zone{db: bd.db}
	bd.Records = Record{db: bd.db}
}
