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

	corsSubrouter := r.PathPrefix("/auth").Subrouter()
	corsSubrouter.HandleFunc("/login", handlers.LoginHandler).Methods(http.MethodPost)
	corsSubrouter.HandleFunc("/refreshToken", handlers.RefreshTokenHandler).Methods(http.MethodPost)

	authSubrouter := r.PathPrefix("/dashboard").Subrouter()
	authSubrouter.Use(middleware.AuthMiddleware)
	authSubrouter.HandleFunc("/branches", handlers.GetBranchesHandler).Methods(http.MethodGet)
	authSubrouter.HandleFunc("/branch/{BRANCH}/runtimeVersions", handlers.GetRuntimeVersionsHandler).Methods(http.MethodGet)
	authSubrouter.HandleFunc("/branch/{BRANCH}/runtimeVersion/{RUNTIME_VERSION}/updates", handlers.GetUpdatesHandler).Methods(http.MethodGet)
	return r
}
