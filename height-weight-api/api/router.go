package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lucasfepe/height-weight-api/config"
	"github.com/lucasfepe/height-weight-api/handlers"
	"github.com/rs/cors"
)

// SetupRouter initializes the router with all the routes
func SetupRouter(cfg *config.Config) http.Handler {
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/api/health", handlers.HealthCheckHandler).Methods(http.MethodGet)

	// API routes
	apiRouter := router.PathPrefix("/api").Subrouter()

	// New weight estimation endpoint using front image, side image, and height
	apiRouter.HandleFunc("/estimate-weight", handlers.EstimateWeight).Methods(http.MethodPost)

	// Legacy endpoints
	apiRouter.HandleFunc("/upload", handlers.NewImageUploadHandler(cfg)).Methods(http.MethodPost)
	apiRouter.HandleFunc("/estimate/{imageID}", handlers.GetEstimationHandler).Methods(http.MethodGet)

	// Configure CORS
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	return corsMiddleware.Handler(router)
}
