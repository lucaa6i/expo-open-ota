package infrastructure

import (
	"expo-open-ota/internal/dashboard"
	"expo-open-ota/internal/handlers"
	"expo-open-ota/internal/metrics"
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

	r.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics.PrometheusHandler().ServeHTTP(w, r)
	}).Methods(http.MethodGet)

	r.HandleFunc("/hc", HealthCheck).Methods(http.MethodGet)
	r.HandleFunc("/manifest", handlers.ManifestHandler).Methods(http.MethodGet)
	r.HandleFunc("/assets", handlers.AssetsHandler).Methods(http.MethodGet)
	r.HandleFunc("/requestUploadUrl/{BRANCH}", handlers.RequestUploadUrlHandler).Methods(http.MethodPost)
	r.HandleFunc("/uploadLocalFile", handlers.RequestUploadLocalFileHandler).Methods(http.MethodPut)
	r.HandleFunc("/markUpdateAsUploaded/{BRANCH}", handlers.MarkUpdateAsUploadedHandler).Methods(http.MethodPost)

	corsSubrouter := r.PathPrefix("/auth").Subrouter()
	corsSubrouter.HandleFunc("/login", handlers.LoginHandler).Methods(http.MethodPost)
	corsSubrouter.HandleFunc("/refreshToken", handlers.RefreshTokenHandler).Methods(http.MethodPost)

	if dashboard.IsDashboardEnabled() {
		r.PathPrefix("/dashboard").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/dashboard" {
				target := "/dashboard/"
				if r.URL.RawQuery != "" {
					target += "?" + r.URL.RawQuery
				}
				http.Redirect(w, r, target, http.StatusMovedPermanently)
				return
			}

			staticExtensions := []string{".css", ".js", ".svg", ".png", ".json", ".ico"}
			for _, ext := range staticExtensions {
				if len(r.URL.Path) > len(ext) && r.URL.Path[len(r.URL.Path)-len(ext):] == ext {
					http.ServeFile(w, r, "./dashboard/dist/"+r.URL.Path[len("/dashboard/"):])
					return
				}
			}

			http.ServeFile(w, r, "./dashboard/dist/index.html")
		}))
	}

	authSubrouter := r.PathPrefix("/api").Subrouter()
	authSubrouter.Use(middleware.AuthMiddleware)
	authSubrouter.HandleFunc("/settings", handlers.GetSettingsHandler).Methods(http.MethodGet)
	authSubrouter.HandleFunc("/branches", handlers.GetBranchesHandler).Methods(http.MethodGet)
	authSubrouter.HandleFunc("/branch/{BRANCH}/runtimeVersions", handlers.GetRuntimeVersionsHandler).Methods(http.MethodGet)
	authSubrouter.HandleFunc("/branch/{BRANCH}/runtimeVersion/{RUNTIME_VERSION}/updates", handlers.GetUpdatesHandler).Methods(http.MethodGet)
	return r
}
