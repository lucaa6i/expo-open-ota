package test

import (
	"encoding/json"
	"expo-open-ota/internal/auth"
	"expo-open-ota/internal/handlers"
	infrastructure "expo-open-ota/internal/router"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestLoginDashboardNotEnabled(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	os.Setenv("USE_DASHBOARD", "false")
	router := infrastructure.NewRouter()
	respRec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", nil)
	router.ServeHTTP(respRec, req)
	assert.Equal(t, http.StatusNotFound, respRec.Code)

}

func TestLoginInvalidPassword(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	router := infrastructure.NewRouter()
	respRec := httptest.NewRecorder()
	formData := url.Values{}
	formData.Set("password", "wrongpassword")
	req, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(respRec, req)
	assert.Equal(t, http.StatusUnauthorized, respRec.Code)
}

func TestShouldRejectLoginIfAdminPasswordNotSet(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	os.Setenv("ADMIN_PASSWORD", "")
	router := infrastructure.NewRouter()
	respRec := httptest.NewRecorder()
	formData := url.Values{}
	formData.Set("password", "admin")
	req, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(respRec, req)
	assert.Equal(t, http.StatusUnauthorized, respRec.Code)
}

func TestLoginValidPassword(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	router := infrastructure.NewRouter()
	respRec := httptest.NewRecorder()
	formData := url.Values{}
	formData.Set("password", "admin")
	req, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(respRec, req)
	assert.Equal(t, http.StatusOK, respRec.Code)
	// Retrieve token & refreshToken from response
	body := respRec.Body.String()

	var response auth.AuthResponse
	err := json.Unmarshal([]byte(body), &response)
	assert.Nil(t, err)
	assert.NotEmpty(t, response.Token)
	assert.NotEmpty(t, response.RefreshToken)
}

func login() auth.AuthResponse {
	router := infrastructure.NewRouter()
	respRec := httptest.NewRecorder()
	formData := url.Values{}
	formData.Set("password", "admin")
	req, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(respRec, req)
	body := respRec.Body.String()
	var response auth.AuthResponse
	_ = json.Unmarshal([]byte(body), &response)
	return response
}

func TestRefreshToken(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	router := infrastructure.NewRouter()
	respRec := httptest.NewRecorder()
	formData := url.Values{}
	formData.Set("refreshToken", login().RefreshToken)
	req, _ := http.NewRequest("POST", "/auth/refreshToken", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(respRec, req)
	assert.Equal(t, http.StatusOK, respRec.Code)
	body := respRec.Body.String()
	var response auth.AuthResponse
	err := json.Unmarshal([]byte(body), &response)
	assert.Nil(t, err)
	assert.NotEmpty(t, response.Token)
	assert.NotEmpty(t, response.RefreshToken)
}

func TestSettings(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	router := infrastructure.NewRouter()
	respRec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dashboard/settings", nil)
	req.Header.Set("Authorization", "Bearer "+login().Token)
	router.ServeHTTP(respRec, req)
	assert.Equal(t, http.StatusOK, respRec.Code)
	var response handlers.SettingsEnv
	err := json.Unmarshal(respRec.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, "{\"BASE_URL\":\"http://localhost:3000\",\"EXPO_APP_ID\":\"EXPO_APP_ID\",\"EXPO_ACCESS_TOKEN\":\"***EXPO_\",\"CACHE_MODE\":\"\",\"REDIS_HOST\":\"\",\"REDIS_PORT\":\"\",\"STORAGE_MODE\":\"local\",\"S3_BUCKET_NAME\":\"\",\"LOCAL_BUCKET_BASE_PATH\":\"/Users/axelmarciano/Workspace/expo-open-ota/test/test-updates\",\"KEYS_STORAGE_TYPE\":\"local\",\"AWSSM_EXPO_PUBLIC_KEY_SECRET_ID\":\"\",\"AWSSM_EXPO_PRIVATE_KEY_SECRET_ID\":\"\",\"PUBLIC_EXPO_KEY_B64\":\"\",\"PUBLIC_LOCAL_EXPO_KEY_PATH\":\"/Users/axelmarciano/Workspace/expo-open-ota/test/keys/public-key-test.pem\",\"PRIVATE_LOCAL_EXPO_KEY_PATH\":\"/Users/axelmarciano/Workspace/expo-open-ota/test/keys/private-key-test.pem\",\"AWS_REGION\":\"eu-west-3\",\"AWS_ACCESS_KEY_ID\":\"\",\"CLOUDFRONT_DOMAIN\":\"\",\"CLOUDFRONT_KEY_PAIR_ID\":\"\",\"CLOUDFRONT_PRIVATE_KEY_B64\":\"\",\"AWSSM_CLOUDFRONT_PRIVATE_KEY_SECRET_ID\":\"\",\"PRIVATE_LOCAL_CLOUDFRONT_KEY_PATH\":\"\",\"PROMETHEUS_ENABLED\":\"\"}", strings.TrimSpace(string(respRec.Body.Bytes())))
}

func TestSettingsWithoutAuth(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	router := infrastructure.NewRouter()
	respRec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dashboard/settings", nil)
	router.ServeHTTP(respRec, req)
	assert.Equal(t, http.StatusUnauthorized, respRec.Code)
}

func TestBranches(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	router := infrastructure.NewRouter()
	respRec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dashboard/branches", nil)
	req.Header.Set("Authorization", "Bearer "+login().Token)
	router.ServeHTTP(respRec, req)
	assert.Equal(t, http.StatusOK, respRec.Code)
	var response []handlers.BranchMapping
	err := json.Unmarshal(respRec.Body.Bytes(), &response)
	assert.Nil(t, err)
	r, err := MockExpoBranchesMappingResponse([]map[string]interface{}{{"id": "master", "name": "master"}}, []map[string]interface{}{{"id": "develop", "name": "develop", "branchMapping": ""}})
	assert.Equal(t, "{\"branches\":[{\"name\":\"master\",\"id\":\"master\"}]}", strings.TrimSpace(string(respRec.Body.Bytes())))
}
