package main

import (
	"log"
	"net"
	"net/http"

	"github.com/DrC0ns0le/bind-api/commit"
	"github.com/DrC0ns0le/bind-api/rdb"

	_ "github.com/DrC0ns0le/bind-api/commit"
)

func main() {

	loadConfig()

	// Connect to the database
	dbConfig := rdb.DBConfig{
		Host:     *dbAddr,
		Port:     *dbPort,
		User:     *dbUser,
		Password: *dbPass,
		DBName:   *dbName,
	}
	rdb.Init(dbConfig)

	commit.Init(*gitToken)

	mux := http.NewServeMux()

	registerRoutes(mux)

	server := &http.Server{
		Addr:    net.JoinHostPort(*listenAddr, *listenPort),
		Handler: mux,
	}

	log.Printf("Listening at %s...\n", net.JoinHostPort(*listenAddr, *listenPort))
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
