package test

import (
	"encoding/json"
	"expo-open-ota/internal/auth"
	"expo-open-ota/internal/bucket"
	"expo-open-ota/internal/handlers"
	infrastructure "expo-open-ota/internal/router"
	"github.com/jarcoal/httpmock"
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
	httpmock.RegisterResponder("POST", "https://api.expo.dev/graphql",
		func(req *http.Request) (*http.Response, error) {
			return MockExpoBranchesMappingResponse([]map[string]interface{}{{"id": "branch-1", "name": "branch-1"}, {"id": "branch-2", "name": "branch-2"}}, []map[string]interface{}{{"id": "staging", "name": "staging", "branchMapping": "{\"data\":[{\"branchId\":\"branch-1\",\"branchMappingLogic\":\"true\"}],\"version\":0}"}})
		})
	req, _ := http.NewRequest("GET", "/dashboard/branches", nil)
	req.Header.Set("Authorization", "Bearer "+login().Token)
	router.ServeHTTP(respRec, req)
	assert.Equal(t, http.StatusOK, respRec.Code)

	var response []handlers.BranchMapping
	err := json.Unmarshal(respRec.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, `[{"branchName":"branch-1","releaseChannel":"staging"},{"branchName":"branch-2","releaseChannel":null},{"branchName":"branch-3","releaseChannel":null},{"branchName":"branch-4","releaseChannel":null}]`, strings.TrimSpace(string(respRec.Body.Bytes())))
}

func TestBranchesWithoutAuth(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	router := infrastructure.NewRouter()
	respRec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dashboard/branches", nil)
	router.ServeHTTP(respRec, req)
	assert.Equal(t, http.StatusUnauthorized, respRec.Code)
}

func TestRuntimeVersionsWithoutAuth(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	router := infrastructure.NewRouter()
	respRec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dashboard/branch/branch-1/runtimeVersions", nil)
	router.ServeHTTP(respRec, req)
	assert.Equal(t, http.StatusUnauthorized, respRec.Code)
}

func TestRuntimeVersions(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	router := infrastructure.NewRouter()
	respRec := httptest.NewRecorder()
	httpmock.RegisterResponder("POST", "https://api.expo.dev/graphql",
		func(req *http.Request) (*http.Response, error) {
			return MockExpoBranchesMappingResponse([]map[string]interface{}{{"id": "branch-1", "name": "branch-1"}, {"id": "branch-2", "name": "branch-2"}}, []map[string]interface{}{{"id": "staging", "name": "staging", "branchMapping": "{\"data\":[{\"branchId\":\"branch-1\",\"branchMappingLogic\":\"true\"}],\"version\":0}"}})
		})
	req, _ := http.NewRequest("GET", "/dashboard/branch/branch-1/runtimeVersions", nil)
	req.Header.Set("Authorization", "Bearer "+login().Token)
	router.ServeHTTP(respRec, req)
	assert.Equal(t, http.StatusOK, respRec.Code)
	var response []bucket.RuntimeVersionWithStats
	err := json.Unmarshal(respRec.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, "[{\"runtimeVersion\":\"1\",\"lastUpdatedAt\":\"1970-01-20T09:02:50Z\",\"createdAt\":\"1970-01-20T09:02:50Z\",\"numberOfUpdates\":1}]", strings.TrimSpace(string(respRec.Body.Bytes())))
}

func TestUpdatesWithoutAuth(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	router := infrastructure.NewRouter()
	respRec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dashboard/branch/branch-1/runtimeVersion/1/updates", nil)
	router.ServeHTTP(respRec, req)
	assert.Equal(t, http.StatusUnauthorized, respRec.Code)
}

func TestUpdatesRegularBranch1(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	router := infrastructure.NewRouter()
	respRec := httptest.NewRecorder()
	httpmock.RegisterResponder("POST", "https://api.expo.dev/graphql",
		func(req *http.Request) (*http.Response, error) {
			return MockExpoBranchesMappingResponse([]map[string]interface{}{{"id": "branch-1", "name": "branch-1"}, {"id": "branch-2", "name": "branch-2"}}, []map[string]interface{}{{"id": "staging", "name": "staging", "branchMapping": "{\"data\":[{\"branchId\":\"branch-1\",\"branchMappingLogic\":\"true\"}],\"version\":0}"}})
		})
	req, _ := http.NewRequest("GET", "/dashboard/branch/branch-1/runtimeVersion/1/updates", nil)
	req.Header.Set("Authorization", "Bearer "+login().Token)
	router.ServeHTTP(respRec, req)
	assert.Equal(t, http.StatusOK, respRec.Code)
	assert.Equal(t, "[{\"updateUUID\":\"b15ed6d8-f39b-04ad-a248-fa3b95fd7e0e\",\"updateId\":\"1674170951\",\"createdAt\":\"1970-01-20T09:02:50Z\",\"commitHash\":\"1674170951\",\"platform\":\"ios\"}]", strings.TrimSpace(string(respRec.Body.Bytes())))
}

func TestUpdatesMultiBranch2(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	router := infrastructure.NewRouter()
	respRec := httptest.NewRecorder()
	httpmock.RegisterResponder("POST", "https://api.expo.dev/graphql",
		func(req *http.Request) (*http.Response, error) {
			return MockExpoBranchesMappingResponse([]map[string]interface{}{{"id": "branch-1", "name": "branch-1"}, {"id": "branch-2", "name": "branch-2"}}, []map[string]interface{}{{"id": "staging", "name": "staging", "branchMapping": "{\"data\":[{\"branchId\":\"branch-1\",\"branchMappingLogic\":\"true\"}],\"version\":0}"}})
		})
	req, _ := http.NewRequest("GET", "/dashboard/branch/branch-2/runtimeVersion/1/updates", nil)
	req.Header.Set("Authorization", "Bearer "+login().Token)
	router.ServeHTTP(respRec, req)
	assert.Equal(t, http.StatusOK, respRec.Code)
	assert.Equal(t, "[{\"updateUUID\":\"291580ca-a34f-73c4-fd82-7902c4129dda\",\"updateId\":\"1737455526\",\"createdAt\":\"1970-01-21T02:37:35Z\",\"commitHash\":\"\",\"platform\":\"\"},{\"updateUUID\":\"b15ed6d8-f39b-04ad-a248-fa3b95fd7e0e\",\"updateId\":\"1674170951\",\"createdAt\":\"1970-01-20T09:02:50Z\",\"commitHash\":\"\",\"platform\":\"\"},{\"updateUUID\":\"187e74b7-9dd7-e43e-75d0-64a843ffa00b\",\"updateId\":\"1666629107\",\"createdAt\":\"1970-01-20T06:57:09Z\",\"commitHash\":\"1674170951\",\"platform\":\"ios\"}]", strings.TrimSpace(string(respRec.Body.Bytes())))
}

func TestUpdatesSomeNotValidBranch4(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	router := infrastructure.NewRouter()
	respRec := httptest.NewRecorder()
	httpmock.RegisterResponder("POST", "https://api.expo.dev/graphql",
		func(req *http.Request) (*http.Response, error) {
			return MockExpoBranchesMappingResponse([]map[string]interface{}{{"id": "branch-1", "name": "branch-1"}, {"id": "branch-2", "name": "branch-2"}}, []map[string]interface{}{{"id": "staging", "name": "staging", "branchMapping": "{\"data\":[{\"branchId\":\"branch-1\",\"branchMappingLogic\":\"true\"}],\"version\":0}"}})
		})
	req, _ := http.NewRequest("GET", "/dashboard/branch/branch-4/runtimeVersion/1/updates", nil)
	req.Header.Set("Authorization", "Bearer "+login().Token)
	router.ServeHTTP(respRec, req)
	assert.Equal(t, http.StatusOK, respRec.Code)
	assert.Equal(t, "[{\"updateUUID\":\"b15ed6d8-f39b-04ad-a248-fa3b95fd7e0e\",\"updateId\":\"1674170951\",\"createdAt\":\"1970-01-20T09:02:50Z\",\"commitHash\":\"1674170951\",\"platform\":\"ios\"}]", strings.TrimSpace(string(respRec.Body.Bytes())))
}
