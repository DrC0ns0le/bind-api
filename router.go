package main

import (
	"net/http"

	"github.com/DrC0ns0le/bind-api/handlers"
	"github.com/DrC0ns0le/bind-api/middleware"
)

func registerRoutes(mux *http.ServeMux) {
	// Public chain (no auth), useful for health and preflight
	publicChain := func(handler func(http.ResponseWriter, *http.Request)) http.Handler {
		return middleware.CorsHandler(
			middleware.RESTMiddleware(
				middleware.LoggerMiddleware(http.HandlerFunc(handler)),
			),
		)
	}

	// Protected chain: requires valid auth token
	protectedChain := func(handler func(http.ResponseWriter, *http.Request)) http.Handler {
		return middleware.CorsHandler(
			middleware.AuthMiddleware(
				middleware.RESTMiddleware(
					middleware.LoggerMiddleware(http.HandlerFunc(handler)),
				),
			),
		)
	}

	// CRUD for zones
	mux.Handle("GET /api/v1/zones", protectedChain(handlers.GetZonesHandler))
	mux.Handle("GET /api/v1/zones/{zone_uuid}", protectedChain(handlers.GetZoneHandler))
	mux.Handle("POST /api/v1/zones", protectedChain(handlers.CreateZoneHandler))
	mux.Handle("OPTIONS /api/v1/zones", middleware.CorsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	mux.Handle("PUT /api/v1/zones/{zone_uuid}", protectedChain(handlers.UpdateZoneHandler))
	mux.Handle("PATCH /api/v1/zones/{zone_uuid}", protectedChain(handlers.UpdateZoneHandler))
	mux.Handle("DELETE /api/v1/zones/{zone_uuid}", protectedChain(handlers.DeleteZoneHandler))

	//CRUD for records
	mux.Handle("GET /api/v1/zones/{zone_uuid}/records", protectedChain(handlers.GetZoneRecordsHandler))
	mux.Handle("GET /api/v1/zones/{zone_uuid}/records/{record_uuid}", protectedChain(handlers.GetRecordHandler))
	mux.Handle("POST /api/v1/zones/{zone_uuid}/records", protectedChain(handlers.CreateRecordHandler))
	mux.Handle("PUT /api/v1/zones/{zone_uuid}/records/{record_uuid}", protectedChain(handlers.UpdateRecordHandler))
	mux.Handle("PATCH /api/v1/zones/{zone_uuid}/records/{record_uuid}", protectedChain(handlers.UpdateRecordHandler))
	mux.Handle("DELETE /api/v1/zones/{zone_uuid}/records/{record_uuid}", protectedChain(handlers.DeleteRecordHandler))
	mux.Handle("GET /api/v1/records/types", publicChain(handlers.GetSupportedRecordTypesHandler))

	//CRUD for configs
	mux.Handle("GET /api/v1/configs", protectedChain(handlers.GetConfigsHandler))
	mux.Handle("GET /api/v1/configs/{config_key}", protectedChain(handlers.GetConfigHandler))
	mux.Handle("POST /api/v1/configs", protectedChain(handlers.CreateConfigHandler))
	mux.Handle("PUT /api/v1/configs", protectedChain(handlers.UpdateConfigHandler))
	mux.Handle("PATCH /api/v1/configs", protectedChain(handlers.UpdateConfigHandler))
	mux.Handle("DELETE /api/v1/configs", protectedChain(handlers.DeleteConfigHandler))

	// Render Zones
	mux.Handle("GET /api/v1/render", protectedChain(handlers.GetRendersHandler))

	// Stage
	mux.Handle("GET /api/v1/staging", protectedChain(handlers.GetStagingHandler))
	mux.Handle("POST /api/v1/staging", protectedChain(handlers.ApplyStagingHandler))

	// Deploy
	mux.Handle("GET /api/v1/deploy", protectedChain(handlers.GetDeployHandler))
	mux.Handle("POST /api/v1/deploy", protectedChain(handlers.DeployHandler))

	// Export/Backup
	mux.Handle("GET /api/v1/export", protectedChain(handlers.ExportAllHandler))
	mux.Handle("GET /api/v1/export/zones", protectedChain(handlers.ExportZonesHandler))
	mux.Handle("GET /api/v1/export/zones/{zone_uuid}", protectedChain(handlers.ExportZoneHandler))
	mux.Handle("GET /api/v1/export/configs", protectedChain(handlers.ExportConfigsHandler))

	// Import/Restore
	mux.Handle("POST /api/v1/import", protectedChain(handlers.ImportAllHandler))
	mux.Handle("POST /api/v1/import/zones", protectedChain(handlers.ImportZonesHandler))
	mux.Handle("POST /api/v1/import/zone", protectedChain(handlers.ImportZoneHandler))
	mux.Handle("POST /api/v1/import/configs", protectedChain(handlers.ImportConfigsHandler))

	// Health check (public)
	mux.Handle("GET /api/v1/health", publicChain(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Catch all
	mux.Handle("/api/v1/", protectedChain(handlers.CatchAllHandler))
}
