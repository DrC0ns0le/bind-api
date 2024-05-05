package main

import (
	"fmt"
	"net/http"

	"github.com/DrC0ns0le/bind-api/handlers"
	"github.com/DrC0ns0le/bind-api/middleware"
	"github.com/DrC0ns0le/bind-api/rdb"
)

var bd *rdb.BindData

func main() {

	bd = &rdb.BindData{}
	bd.Init("10.2.1.2", 5432, "postgres", "jack2001", "bind_dns")
	handlers.Init(bd)

	mux := http.NewServeMux()

	middlewareChain := func(handler func(http.ResponseWriter, *http.Request)) http.Handler {
		return middleware.LoggerMiddleware(middleware.CorsHandler(http.HandlerFunc(handler)))
	}

	// CRUD for zones
	mux.Handle("GET /api/v1/zones", middlewareChain(handlers.GetZonesHandler))
	mux.Handle("GET /api/v1/zones/{zone_uuid}", middlewareChain(handlers.GetZoneHandler))
	mux.Handle("POST /api/v1/zones", middlewareChain(handlers.CreateZoneHandler))
	mux.Handle("OPTIONS /api/v1/zones", middleware.CorsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	mux.Handle("DELETE /api/v1/zones/{zone_uuid}", middlewareChain(handlers.DeleteZoneHandler))

	//CRUD for records
	mux.Handle("GET /api/v1/zones/{zone_uuid}/records", middlewareChain(handlers.GetZoneRecordsHandler))
	mux.Handle("GET /api/v1/zones/{zone_uuid}/records/{record_uuid}", middlewareChain(handlers.GetRecordHandler))
	mux.Handle("POST /api/v1/zones/{zone_uuid}/records", middlewareChain(handlers.CreateRecordHandler))
	mux.Handle("PUT /api/v1/zones/{zone_uuid}/records/{record_uuid}", middlewareChain(handlers.UpdateRecordHandler))
	mux.Handle("PATCH /api/v1/zones/{zone_uuid}/records/{record_uuid}", middlewareChain(handlers.UpdateRecordHandler))
	mux.Handle("DELETE /api/v1/zones/{zone_uuid}/records/{record_uuid}", middlewareChain(handlers.DeleteRecordHandler))

	// Render Zones
	mux.Handle("/api/v1/render", middlewareChain(handlers.RenderZonesHandler))

	// Health check
	mux.Handle("GET /health", middleware.CorsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	// Catch all
	mux.Handle("/", middleware.CorsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})))

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
