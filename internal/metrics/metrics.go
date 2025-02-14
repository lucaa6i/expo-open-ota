package metrics

import (
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
		[]string{"clientId", "platform", "runtime", "branch", "update"},
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

func TrackActiveUser(clientId, platform, runtime, branch, update string) {
	if clientId == "" || update == "" || platform == "" || branch == "" {
		return
	}
	activeUsersVec.WithLabelValues(clientId, platform, runtime, branch, update).Set(1)
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
			Help: "Total number of unique active users per clientId, platform, runtime version, branch and update",
		},
		[]string{"clientId", "platform", "runtime", "branch", "update"},
	)
	updateDownloadsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "update_downloads_total",
			Help: "Total number of update downloads per platform, runtime version, branch and update",
		},
		[]string{"platform", "runtime", "branch", "update", "updateType"},
	)
}
