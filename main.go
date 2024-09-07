package main

import (
	"log"
	"net/http"

	"github.com/DrC0ns0le/bind-api/rdb"

	_ "github.com/DrC0ns0le/bind-api/commit"
)

func main() {

	const listenAddr = "0.0.0.0:9174"

	// Connect to the database
	dbConfig := rdb.DBConfig{
		Host:     "10.2.1.2",
		Port:     5643,
		User:     "postgres",
		Password: "jack2001",
		DBName:   "bind_dns",
	}
	rdb.Init(dbConfig)

	mux := http.NewServeMux()

	registerRoutes(mux)

	server := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	log.Printf("Listening at %s...\n", listenAddr)
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
