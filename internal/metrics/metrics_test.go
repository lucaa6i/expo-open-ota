package metrics_test

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"expo-open-ota/internal/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupMetrics(t *testing.T) func() {
	os.Setenv("PROMETHEUS_ENABLED", "true")
	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg
	metrics.ResetMetricsForTest()
	metrics.InitMetrics()
	return func() {}
}

func getMetricValue(metricName string, labelFilter map[string]string) float64 {
	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return 0
	}
	for _, mf := range mfs {
		if mf.GetName() == metricName {
			for _, m := range mf.Metric {
				match := true
				for key, value := range labelFilter {
					found := false
					for _, label := range m.Label {
						if label.GetName() == key && label.GetValue() == value {
							found = true
							break
						}
					}
					if !found {
						match = false
						break
					}
				}
				if match {
					if m.Counter != nil {
						return m.Counter.GetValue()
					}
					if m.Gauge != nil {
						return m.Gauge.GetValue()
					}
				}
			}
		}
	}
	return 0
}

func getActiveUsers(runtime, branch, update string) float64 {
	return getMetricValue("active_users_total", map[string]string{
		"runtime": runtime,
		"branch":  branch,
		"update":  update,
	})
}

func getTotalActiveUsersByBranchAndRuntime(runtime, branch string, updates []string) float64 {
	total := 0.0
	for _, update := range updates {
		total += getActiveUsers(runtime, branch, update)
	}
	return total
}

func getTotalUpdateDownloads(runtime, branch, update, updateType string) float64 {
	return getMetricValue("update_downloads_total", map[string]string{
		"runtime":    runtime,
		"branch":     branch,
		"update":     update,
		"updateType": updateType,
	})
}

func TestTrackUpdateDownload(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	metrics.TrackUpdateDownload("1.0.0", "stable", "update42", "normal")
	val := getTotalUpdateDownloads("1.0.0", "stable", "update42", "normal")
	if val != 1 {
		t.Errorf("Expected update_downloads_total to be 1, got %v", val)
	}
}

func TestTrackRuntimeVersion(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	metrics.TrackRuntimeVersion("1.0.0", "stable")
	val := getMetricValue("runtime_versions_total", map[string]string{
		"runtime": "1.0.0",
		"branch":  "stable",
	})
	if val != 1 {
		t.Errorf("Expected runtime_versions_total to be 1, got %v", val)
	}
}

func TestTrackActiveUser(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	metrics.TrackActiveUser("1.0.0", "stable", "update42")
	val := getActiveUsers("1.0.0", "stable", "update42")
	if val != 1 {
		t.Errorf("Expected active_users_total to be 1, got %v", val)
	}
}

func TestGetActiveUsers(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	if got := getActiveUsers("1.0.0", "stable", "update42"); got != 0 {
		t.Errorf("Expected getActiveUsers to return 0, got %v", got)
	}
	metrics.TrackActiveUser("1.0.0", "stable", "update42")
	if got := getActiveUsers("1.0.0", "stable", "update42"); got != 1 {
		t.Errorf("Expected getActiveUsers to return 1, got %v", got)
	}
	metrics.TrackActiveUser("1.0.0", "stable", "update42")
	if got := getActiveUsers("1.0.0", "stable", "update42"); got != 2 {
		t.Errorf("Expected getActiveUsers to return 2, got %v", got)
	}
}

func TestGetTotalActiveUsersByBranchAndRuntime(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	updates := []string{"update42", "update43"}
	if got := getTotalActiveUsersByBranchAndRuntime("1.0.0", "stable", updates); got != 0 {
		t.Errorf("Expected total active users to be 0, got %v", got)
	}
	metrics.TrackActiveUser("1.0.0", "stable", "update42")
	metrics.TrackActiveUser("1.0.0", "stable", "update43")
	if got := getTotalActiveUsersByBranchAndRuntime("1.0.0", "stable", updates); got != 2 {
		t.Errorf("Expected total active users to be 2, got %v", got)
	}
}

func TestGetTotalUpdateDownloadsByUpdate(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	if got := getTotalUpdateDownloads("1.0.0", "stable", "update42", "normal"); got != 0 {
		t.Errorf("Expected total update downloads to be 0, got %v", got)
	}
	metrics.TrackUpdateDownload("1.0.0", "stable", "update42", "normal")
	if got := getTotalUpdateDownloads("1.0.0", "stable", "update42", "normal"); got != 1 {
		t.Errorf("Expected total update downloads to be 1, got %v", got)
	}
	metrics.TrackUpdateDownload("1.0.0", "stable", "update42", "normal")
	if got := getTotalUpdateDownloads("1.0.0", "stable", "update42", "normal"); got != 2 {
		t.Errorf("Expected total update downloads to be 2, got %v", got)
	}
}

func TestPrometheusHandler(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	metrics.TrackUpdateDownload("1.0.0", "stable", "update42", "normal")
	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	handler := promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{})
	handler.ServeHTTP(rr, req)
	body := rr.Body.String()
	if !strings.Contains(body, "update_downloads_total") {
		t.Errorf("Expected update_downloads_total in metrics, got %s", body)
	}
}
