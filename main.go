package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/DrC0ns0le/bind-api/commit"
	"github.com/DrC0ns0le/bind-api/config"
	"github.com/DrC0ns0le/bind-api/rdb"
)

var (
	listenAddr = flag.String("listen.addr", "0.0.0.0", "listen address, env: LISTEN_ADDR")
	listenPort = flag.String("listen.port", "8080", "listen port, env: LISTEN_PORT")
)

func main() {

	config.Init()

	// Connect to the database
	err := rdb.Init()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the commit
	err = commit.Init()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	registerRoutes(mux)

	server := &http.Server{
		Addr:    net.JoinHostPort(config.GetEnv("LISTEN_ADDR", *listenAddr), config.GetEnv("LISTEN_PORT", *listenPort)),
		Handler: mux,
	}

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
