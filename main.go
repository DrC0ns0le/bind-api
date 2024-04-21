package main

import (
	"fmt"
	"net/http"

	"github.com/DrC0ns0le/bind-api/handlers"
	"github.com/DrC0ns0le/bind-api/rdb"
)

var bd *rdb.BindData

func main() {

	bd = &rdb.BindData{}
	bd.Init("10.2.1.2", 5432, "postgres", "jack2001", "bind_dns")
	handlers.Init(bd)

	mux := http.NewServeMux()

	// CRUD for zones
	mux.HandleFunc("GET /api/v1/zones/", handlers.GetZonesHandler)
	mux.HandleFunc("POST /api/v1/zones/", handlers.CreateZoneHandler)
	mux.HandleFunc("DELETE /api/v1/zones/{zone_uuid}", handlers.DeleteZoneHandler)

	//CRUD for records
	mux.HandleFunc("GET /api/v1/zones/{zone_uuid}/records", handlers.GetZoneRecordsHandler)
	mux.HandleFunc("POST /api/v1/zones/{zone_uuid}/records/{record_uuid}", handlers.GetRecordHandler)
	mux.HandleFunc("POST /api/v1/zones/{zone_uuid}/records", handlers.CreateRecordHandler)
	mux.HandleFunc("PUT /api/v1/zones/{zone_uuid}/records/{record_uuid}", handlers.UpdateRecordHandler)
	mux.HandleFunc("PATCH /api/v1/zones/{zone_uuid}/records/{record_uuid}", handlers.UpdateRecordHandler)
	mux.HandleFunc("DELETE /api/v1/zones/{zone_uuid}/records/{record_uuid}", handlers.DeleteRecordHandler)

	// Render Zones
	mux.HandleFunc("/api/v1/render", handlers.RenderZonesHandler)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

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
