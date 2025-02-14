package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	updateDownloadsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "update_downloads_total",
			Help: "Total number of update downloads per runtime version, branch and update",
		},
		[]string{"runtime", "branch", "update", "updateType"},
	)
	runtimeVersionsUsedVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "runtime_versions_total",
			Help: "Total number of runtime versions used per runtime version and branch",
		},
		[]string{"runtime", "branch"},
	)
	activeUsersVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Total number of active users per runtime version, branch and update",
		},
		[]string{"runtime", "branch", "update"},
	)
)

func InitMetrics() {
	prometheus.MustRegister(updateDownloadsVec)
	prometheus.MustRegister(runtimeVersionsUsedVec)
	prometheus.MustRegister(activeUsersVec)
}

func TrackUpdateDownload(runtime, branch, update, updateType string) {
	updateDownloadsVec.WithLabelValues(runtime, branch, update, updateType).Inc()
}

func TrackRuntimeVersion(runtime, branch string) {
	runtimeVersionsUsedVec.WithLabelValues(runtime, branch).Inc()
}

func TrackActiveUser(runtime, branch, update string) {
	activeUsersVec.WithLabelValues(runtime, branch, update).Inc()
}

func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

func ResetMetricsForTest() {
	updateDownloadsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "update_downloads_total",
			Help: "Total number of update downloads per runtime version, branch and update",
		},
		[]string{"runtime", "branch", "update", "updateType"},
	)
	runtimeVersionsUsedVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "runtime_versions_total",
			Help: "Total number of runtime versions used per runtime version and branch",
		},
		[]string{"runtime", "branch"},
	)
	activeUsersVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Total number of active users per runtime version, branch and update",
		},
		[]string{"runtime", "branch", "update"},
	)
}
