package main

import (
	"log"
	"net/http"

	"github.com/DrC0ns0le/bind-api/handlers"
	"github.com/DrC0ns0le/bind-api/middleware"
	"github.com/DrC0ns0le/bind-api/rdb"

	_ "github.com/DrC0ns0le/bind-api/commit"
)

func main() {

	// Connect to the database
	dbConfig := rdb.DBConfig{
		Host:     "10.2.1.2",
		Port:     5432,
		User:     "postgres",
		Password: "jack2001",
		DBName:   "bind_dns",
	}
	rdb.Init(dbConfig)

	mux := http.NewServeMux()

	middlewareChain := func(handler func(http.ResponseWriter, *http.Request)) http.Handler {
		return middleware.RESTMiddleware(middleware.LoggerMiddleware(middleware.CorsHandler(http.HandlerFunc(handler))))
	}

	// CRUD for zones
	mux.Handle("GET /api/v1/zones", middlewareChain(handlers.GetZonesHandler))
	mux.Handle("GET /api/v1/zones/{zone_uuid}", middlewareChain(handlers.GetZoneHandler))
	mux.Handle("POST /api/v1/zones", middlewareChain(handlers.CreateZoneHandler))
	mux.Handle("OPTIONS /api/v1/zones", middleware.CorsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	mux.Handle("PUT /api/v1/zones/{zone_uuid}", middlewareChain(handlers.UpdateZoneHandler))
	mux.Handle("PATCH /api/v1/zones/{zone_uuid}", middlewareChain(handlers.UpdateZoneHandler))
	mux.Handle("DELETE /api/v1/zones/{zone_uuid}", middlewareChain(handlers.DeleteZoneHandler))

	//CRUD for records
	mux.Handle("GET /api/v1/zones/{zone_uuid}/records", middlewareChain(handlers.GetZoneRecordsHandler))
	mux.Handle("GET /api/v1/zones/{zone_uuid}/records/{record_uuid}", middlewareChain(handlers.GetRecordHandler))
	mux.Handle("POST /api/v1/zones/{zone_uuid}/records", middlewareChain(handlers.CreateRecordHandler))
	mux.Handle("PUT /api/v1/zones/{zone_uuid}/records/{record_uuid}", middlewareChain(handlers.UpdateRecordHandler))
	mux.Handle("PATCH /api/v1/zones/{zone_uuid}/records/{record_uuid}", middlewareChain(handlers.UpdateRecordHandler))
	mux.Handle("DELETE /api/v1/zones/{zone_uuid}/records/{record_uuid}", middlewareChain(handlers.DeleteRecordHandler))

	//CRUD for configs
	mux.Handle("GET /api/v1/configs", middlewareChain(handlers.GetConfigsHandler))
	mux.Handle("GET /api/v1/configs/{config_key}", middlewareChain(handlers.GetConfigHandler))
	mux.Handle("POST /api/v1/configs", middlewareChain(handlers.CreateConfigHandler))
	mux.Handle("PUT /api/v1/configs", middlewareChain(handlers.UpdateConfigHandler))
	mux.Handle("PATCH /api/v1/configs", middlewareChain(handlers.UpdateConfigHandler))
	mux.Handle("DELETE /api/v1/configs", middlewareChain(handlers.DeleteConfigHandler))

	// Render Zones
	mux.Handle("GET /api/v1/render", middlewareChain(handlers.GetRendersHandler))

	// Stage
	mux.Handle("GET /api/v1/staging", middlewareChain(handlers.GetStagingHandler))
	mux.Handle("POST /api/v1/staging", middlewareChain(handlers.ApplyStagingHandler))

	// Deploy
	mux.Handle("GET /api/v1/deploy", middlewareChain(handlers.GetDeployHandler))
	mux.Handle("POST /api/v1/deploy", middlewareChain(handlers.DeployHandler))

	// Health check
	mux.Handle("GET /api/v1/health", middleware.CorsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	// Catch all
	mux.Handle("/api/v1/", middleware.RESTMiddleware(middleware.CorsHandler(http.HandlerFunc(handlers.CatchAllHandler))))

	const listenAddr = "0.0.0.0:9174"

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
