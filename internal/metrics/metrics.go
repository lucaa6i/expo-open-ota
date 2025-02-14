package metrics

import (
	"expo-open-ota/config"
	"expo-open-ota/internal/cache"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
)

var (
	appCache             cache.Cache
	prometheusEnabled    bool
	totalUpdateDownloads = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "update_downloads_total",
			Help: "Total number of update downloads",
		})

	totalClientsOnUpdates = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "clients_on_updates_total",
			Help: "Total number of clients running on an update",
		})

	totalRuntimeVersions = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "runtime_versions_total",
			Help: "Total number of runtime versions used",
		})

	totalActiveUsers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Total number of active users",
		})
)

func InitMetrics() {
	appCache = cache.GetCache()
	prometheusEnabled = config.GetEnv("PROMETHEUS_ENABLED") == "true"
	if prometheusEnabled {
		prometheus.MustRegister(totalUpdateDownloads)
		prometheus.MustRegister(totalClientsOnUpdates)
		prometheus.MustRegister(totalRuntimeVersions)
		prometheus.MustRegister(totalActiveUsers)
	}
}

func TrackUpdateDownload() {
	if appCache == nil {
		return
	}
	key := "update:downloads"
	count := getInt(appCache.Get(key)) + 1
	appCache.Set(key, fmt.Sprintf("%d", count), nil)
	if prometheusEnabled {
		totalUpdateDownloads.Inc()
	}
}

func TrackClientOnUpdate() {
	if appCache == nil {
		return
	}
	key := "clients:on_updates"
	count := getInt(appCache.Get(key)) + 1
	appCache.Set(key, fmt.Sprintf("%d", count), nil)
	if prometheusEnabled {
		totalClientsOnUpdates.Set(float64(count))
	}
}

func TrackRuntimeVersion() {
	if appCache == nil {
		return
	}
	key := "runtime:versions"
	count := getInt(appCache.Get(key)) + 1
	appCache.Set(key, fmt.Sprintf("%d", count), nil)
	if prometheusEnabled {
		totalRuntimeVersions.Set(float64(count))
	}
}

func TrackActiveUser() {
	if appCache == nil {
		return
	}
	key := "active_users"
	count := getInt(appCache.Get(key)) + 1
	appCache.Set(key, fmt.Sprintf("%d", count), nil)
	if prometheusEnabled {
		totalActiveUsers.Set(float64(count))
	}
}

func getInt(value string) int {
	if value == "" {
		return 0
	}
	i, _ := strconv.Atoi(value)
	return i
}

func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}
