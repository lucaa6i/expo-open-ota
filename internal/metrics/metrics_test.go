package metrics_test

import (
	"net/http/httptest"
	"os"
	"regexp"
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
				for key, filterValue := range labelFilter {
					found := false
					for _, label := range m.Label {
						if label.GetName() == key {
							matched, err := regexp.MatchString(filterValue, label.GetValue())
							if err == nil && matched {
								found = true
								break
							}
						}
					}
					if !found {
						match = false
						break
					}
				}
				if match {
					if m.Gauge != nil {
						return m.Gauge.GetValue()
					}
					if m.Counter != nil {
						return m.Counter.GetValue()
					}
				}
			}
		}
	}
	return 0
}

func getActiveUsers(platform, runtime, branch, update string) float64 {
	return getMetricValue("active_users_total", map[string]string{
		"platform": platform,
		"runtime":  runtime,
		"branch":   branch,
		"update":   update,
	})
}

func getTotalUpdateDownloads(platform, runtime, branch, update, updateType string) float64 {
	return getMetricValue("update_downloads_total", map[string]string{
		"platform":   platform,
		"runtime":    runtime,
		"branch":     branch,
		"update":     update,
		"updateType": updateType,
	})
}

func TestTrackUpdateDownload(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	platform := "ios"
	runtime := "1.0.0"
	branch := "stable"
	update := "update42"
	updateType := "normal"
	metrics.TrackUpdateDownload(platform, runtime, branch, update, updateType)
	val := getTotalUpdateDownloads(platform, runtime, branch, update, updateType)
	if val != 1 {
		t.Errorf("Expected update_downloads_total to be 1, got %v", val)
	}
}

func TestTrackActiveUser(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	clientId := "client1"
	platform := "ios"
	runtime := "1.0.0"
	branch := "stable"
	update := "update42"
	metrics.TrackActiveUser(clientId, platform, runtime, branch, update)
	val := getActiveUsers(platform, runtime, branch, update)
	if val != 1 {
		t.Errorf("Expected active_users_total to be 1, got %v", val)
	}
}

func TestGetActiveUsers(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	clientId := "client1"
	platform := "ios"
	runtime := "1.0.0"
	branch := "stable"
	update := "update42"
	if got := getActiveUsers(platform, runtime, branch, update); got != 0 {
		t.Errorf("Expected getActiveUsers to return 0, got %v", got)
	}
	metrics.TrackActiveUser(clientId, platform, runtime, branch, update)
	if got := getActiveUsers(platform, runtime, branch, update); got != 1 {
		t.Errorf("Expected getActiveUsers to return 1, got %v", got)
	}
	metrics.TrackActiveUser("client2", platform, runtime, branch, update)
	if got := getActiveUsers(platform, runtime, branch, update); got != 1 {
		t.Errorf("Expected getActiveUsers to still be 1 (Gauge should not increment), got %v", got)
	}
}

func TestGetTotalUpdateDownloadsByUpdate(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	platform := "ios"
	runtime := "1.0.0"
	branch := "stable"
	update := "update42"
	updateType := "normal"
	if got := getTotalUpdateDownloads(platform, runtime, branch, update, updateType); got != 0 {
		t.Errorf("Expected total update downloads to be 0, got %v", got)
	}
	metrics.TrackUpdateDownload(platform, runtime, branch, update, updateType)
	if got := getTotalUpdateDownloads(platform, runtime, branch, update, updateType); got != 1 {
		t.Errorf("Expected total update downloads to be 1, got %v", got)
	}
	metrics.TrackUpdateDownload(platform, runtime, branch, update, updateType)
	if got := getTotalUpdateDownloads(platform, runtime, branch, update, updateType); got != 2 {
		t.Errorf("Expected total update downloads to be 2, got %v", got)
	}
}

func TestPrometheusHandler(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	platform := "ios"
	runtime := "1.0.0"
	branch := "stable"
	update := "update42"
	updateType := "normal"
	metrics.TrackUpdateDownload(platform, runtime, branch, update, updateType)
	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	handler := promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{})
	handler.ServeHTTP(rr, req)
	body := rr.Body.String()
	if !strings.Contains(body, "update_downloads_total") {
		t.Errorf("Expected update_downloads_total in metrics, got %s", body)
	}
}
