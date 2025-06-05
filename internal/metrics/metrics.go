package metrics

import (
	"expo-open-ota/internal/cache"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	activeUsersVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Total number of unique active users per clientId, platform, runtime version, branch and update",
		},
		[]string{"platform", "runtime", "branch", "update"},
	)
	updateDownloadsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "update_downloads_total",
			Help: "Total number of update downloads per platform, runtime version, branch and update",
		},
		[]string{"platform", "runtime", "branch", "update", "updateType"},
	)
)

func InitMetrics() {
	prometheus.MustRegister(activeUsersVec)
	prometheus.MustRegister(updateDownloadsVec)
}

func CleanupMetrics() {
	prometheus.Unregister(activeUsersVec)
	prometheus.Unregister(updateDownloadsVec)
}

func TrackActiveUser(clientId, platform, runtime, branch, update string) {
	if clientId == "" || platform == "" || branch == "" || update == "" {
		return
	}

	resolvedCache := cache.GetCache()
	key := fmt.Sprintf("seen_users:%s:%s", branch, platform)
	ttl := 86400

	_ = resolvedCache.Sadd(key, []string{clientId}, &ttl)

	count, err := resolvedCache.Scard(key)
	if err != nil {
		return
	}
	activeUsersVec.WithLabelValues(platform, runtime, branch, update).Set(float64(count))
}

func TrackUpdateDownload(platform, runtime, branch, update, updateType string) {
	if update == "" || platform == "" || branch == "" {
		return
	}
	updateDownloadsVec.WithLabelValues(platform, runtime, branch, update, updateType).Inc()
}

func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

func ResetMetricsForTest() {
	activeUsersVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Total number of unique active users per platform, runtime version, branch and update",
		},
		[]string{"platform", "runtime", "branch", "update"},
	)
	updateDownloadsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "update_downloads_total",
			Help: "Total number of update downloads per platform, runtime version, branch and update",
		},
		[]string{"platform", "runtime", "branch", "update", "updateType"},
	)
}
