package infrastructure

import (
	"expo-open-ota/internal/handlers"
	"expo-open-ota/internal/middleware"
	"github.com/gorilla/mux"
	"net/http"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.HandleFunc("/hc", HealthCheck).Methods(http.MethodGet)
	r.HandleFunc("/manifest", handlers.ManifestHandler).Methods(http.MethodGet)
	r.HandleFunc("/assets", handlers.AssetsHandler).Methods(http.MethodGet)
	r.HandleFunc("/requestUploadUrl/{BRANCH}", handlers.RequestUploadUrlHandler).Methods(http.MethodPost)
	r.HandleFunc("/uploadLocalFile", handlers.RequestUploadLocalFileHandler).Methods(http.MethodPut)
	r.HandleFunc("/markUpdateAsUploaded/{BRANCH}", handlers.MarkUpdateAsUploadedHandler).Methods(http.MethodPost)
	return r
}
