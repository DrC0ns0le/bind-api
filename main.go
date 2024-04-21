package main

import (
	"fmt"
	"net/http"

	"github.com/DrC0ns0le/bind-api/rdb"
)

var bd *rdb.BindData

func main() {

	bd = &rdb.BindData{}
	bd.Init("10.2.1.2", 5432, "postgres", "jack2001", "bind_dns")

	mux := http.NewServeMux()

	// CRUD for zones
	mux.HandleFunc("GET /api/v1/zones/", GetZonesHandler)
	mux.HandleFunc("POST /api/v1/zones/", CreateZoneHandler)
	mux.HandleFunc("DELETE /api/v1/zones/{zone_uuid}", DeleteZoneHandler)

	//CRUD for records
	mux.HandleFunc("GET /api/v1/zones/{zone_uuid}/records", GetZoneRecordsHandler)
	mux.HandleFunc("POST /api/v1/zones/{zone_uuid}/records/{record_uuid}", GetRecordHandler)
	mux.HandleFunc("POST /api/v1/zones/{zone_uuid}/records", CreateRecordHandler)
	mux.HandleFunc("PUT /api/v1/zones/{zone_uuid}/records/{record_uuid}", UpdateRecordHandler)
	mux.HandleFunc("PATCH /api/v1/zones/{zone_uuid}/records/{record_uuid}", UpdateRecordHandler)
	mux.HandleFunc("DELETE /api/v1/zones/{zone_uuid}/records/{record_uuid}", DeleteRecordHandler)

	http.ListenAndServe(":8080", mux)

	const listenAddr = "0.0.0.0:8090"

	server := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	fmt.Printf("Listening at %s...\n", listenAddr)
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
