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
	appCache          cache.Cache
	prometheusEnabled bool

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
	appCache = cache.GetCache()
	prometheusEnabled = config.GetEnv("PROMETHEUS_ENABLED") == "true"
	if prometheusEnabled {
		prometheus.MustRegister(updateDownloadsVec)
		prometheus.MustRegister(runtimeVersionsUsedVec)
		prometheus.MustRegister(activeUsersVec)
	}
}

func TrackUpdateDownload(runtime, branch, update, updateType string) {
	if appCache == nil {
		return
	}
	key := fmt.Sprintf("update:downloads:%s:%s:%s", runtime, branch, update)
	count := getInt(appCache.Get(key)) + 1
	appCache.Set(key, fmt.Sprintf("%d", count), nil)
	if prometheusEnabled {
		updateDownloadsVec.WithLabelValues(runtime, branch, update, updateType).Inc()
	}
}

func TrackRuntimeVersion(runtime, branch string) {
	if appCache == nil {
		return
	}
	key := fmt.Sprintf("runtime:versions:%s:%s", runtime, branch)
	count := getInt(appCache.Get(key)) + 1
	appCache.Set(key, fmt.Sprintf("%d", count), nil)
	if prometheusEnabled {
		runtimeVersionsUsedVec.WithLabelValues(runtime, branch).Set(float64(count))
	}
}

func TrackActiveUser(runtime, branch, update string) {
	if appCache == nil {
		return
	}
	key := fmt.Sprintf("active_users:%s:%s:%s", runtime, branch, update)
	count := getInt(appCache.Get(key)) + 1
	appCache.Set(key, fmt.Sprintf("%d", count), nil)
	if prometheusEnabled {
		activeUsersVec.WithLabelValues(runtime, branch, update).Set(float64(count))
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

func GetTotalActiveUsersByBranchAndRuntime(branch, runtime string, updates []string) int {
	total := 0
	for _, update := range updates {
		key := fmt.Sprintf("active_users:%s:%s:%s", runtime, branch, update)
		total += getInt(appCache.Get(key))
	}
	return total
}

func GetActiveUsers(runtime, branch, update string) int {
	key := fmt.Sprintf("active_users:%s:%s:%s", runtime, branch, update)
	return getInt(appCache.Get(key))
}

func GetTotalUpdateDownloadsByUpdate(branch, runtime, update string) int {
	key := fmt.Sprintf("update:downloads:%s:%s:%s", runtime, branch, update)
	return getInt(appCache.Get(key))
}
