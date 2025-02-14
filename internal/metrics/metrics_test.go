package metrics_test

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"expo-open-ota/internal/cache"
	"expo-open-ota/internal/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

func setupMetrics(t *testing.T) func() {
	r := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = r
	prometheus.DefaultGatherer = r

	os.Setenv("CACHE_MODE", "local")
	os.Setenv("PROMETHEUS_ENABLED", "true")
	c := cache.GetCache()
	if err := c.Clear(); err != nil {
		t.Fatalf("failed to clear cache: %v", err)
	}
	metrics.InitMetrics()
	return func() {
		if err := c.Clear(); err != nil {
			t.Errorf("failed to clear cache on teardown: %v", err)
		}
	}
}

func TestTrackUpdateDownload(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	metrics.TrackUpdateDownload("1.0.0", "stable", "update42", "normal")
	c := cache.GetCache()
	expectedKey := "update:downloads:1.0.0:stable:update42"
	if c.Get(expectedKey) != "1" {
		t.Errorf("Expected %s to be 1, got %s", expectedKey, c.Get(expectedKey))
	}
}

func TestTrackRuntimeVersion(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	metrics.TrackRuntimeVersion("1.0.0", "stable")
	c := cache.GetCache()
	expectedKey := "runtime:versions:1.0.0:stable"
	if c.Get(expectedKey) != "1" {
		t.Errorf("Expected %s to be 1, got %s", expectedKey, c.Get(expectedKey))
	}
}

func TestTrackActiveUser(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	metrics.TrackActiveUser("1.0.0", "stable", "update42")
	c := cache.GetCache()
	expectedKey := "active_users:1.0.0:stable:update42"
	if c.Get(expectedKey) != "1" {
		t.Errorf("Expected %s to be 1, got %s", expectedKey, c.Get(expectedKey))
	}
}

func TestGetActiveUsers(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	if got := metrics.GetActiveUsers("1.0.0", "stable", "update42"); got != 0 {
		t.Errorf("Expected GetActiveUsers to return 0, got %d", got)
	}
	metrics.TrackActiveUser("1.0.0", "stable", "update42")
	if got := metrics.GetActiveUsers("1.0.0", "stable", "update42"); got != 1 {
		t.Errorf("Expected GetActiveUsers to return 1, got %d", got)
	}
	metrics.TrackActiveUser("1.0.0", "stable", "update42")
	if got := metrics.GetActiveUsers("1.0.0", "stable", "update42"); got != 2 {
		t.Errorf("Expected GetActiveUsers to return 2, got %d", got)
	}
}

func TestGetTotalActiveUsersByBranchAndRuntime(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	updates := []string{"update42", "update43"}
	if got := metrics.GetTotalActiveUsersByBranchAndRuntime("stable", "1.0.0", updates); got != 0 {
		t.Errorf("Expected total active users to be 0, got %d", got)
	}
	metrics.TrackActiveUser("1.0.0", "stable", "update42")
	metrics.TrackActiveUser("1.0.0", "stable", "update43")
	if got := metrics.GetTotalActiveUsersByBranchAndRuntime("stable", "1.0.0", updates); got != 2 {
		t.Errorf("Expected total active users to be 2, got %d", got)
	}
}

func TestGetTotalUpdateDownloadsByUpdate(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	if got := metrics.GetTotalUpdateDownloadsByUpdate("stable", "1.0.0", "update42"); got != 0 {
		t.Errorf("Expected total update downloads to be 0, got %d", got)
	}
	metrics.TrackUpdateDownload("1.0.0", "stable", "update42", "normal")
	if got := metrics.GetTotalUpdateDownloadsByUpdate("stable", "1.0.0", "update42"); got != 1 {
		t.Errorf("Expected total update downloads to be 1, got %d", got)
	}
	metrics.TrackUpdateDownload("1.0.0", "stable", "update42", "normal")
	if got := metrics.GetTotalUpdateDownloadsByUpdate("stable", "1.0.0", "update42"); got != 2 {
		t.Errorf("Expected total update downloads to be 2, got %d", got)
	}
}

func TestPrometheusHandler(t *testing.T) {
	teardown := setupMetrics(t)
	defer teardown()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	handler := metrics.PrometheusHandler()
	handler.ServeHTTP(rr, req)
	body := rr.Body.String()
	if !strings.Contains(body, "update_downloads_total") {
		t.Errorf("Expected update_downloads_total in metrics, got %s", body)
	}
}
