package test

import (
	"bytes"
	"encoding/json"
	"expo-open-ota/internal/bucket"
	cache2 "expo-open-ota/internal/cache"
	"expo-open-ota/internal/handlers"
	"expo-open-ota/internal/services"
	"expo-open-ota/internal/update"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestRequestUploadUrlWithoutBearer(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	mockExpoForRequestUploadUrlTest("staging")

	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Errorf("Error finding project root: %v", err)
	}
	os.Setenv("LOCAL_BUCKET_BASE_PATH", filepath.Join(projectRoot, "./updates"))
	q := "http://localhost:3000/requestUploadUrl/DO_NOT_USE?runtimeVersion=1"
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", q, nil)
	r.Header.Set("Authorization", "Bearer expo_alternative_token")
	r = mux.SetURLVars(r, map[string]string{
		"BRANCH": "DO_NOT_USE",
	})
	sampleUpdatePath := filepath.Join(projectRoot, "/test/test-updates/branch-1/1/1674170951")
	uploadRequestsInput := ComputeUploadRequestsInput(sampleUpdatePath)
	uploadRequestsInputJSON, err := json.Marshal(uploadRequestsInput)
	if err != nil {
		t.Errorf("Error marshalling uploadRequestsInput: %v", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(uploadRequestsInputJSON))
	handlers.RequestUploadUrlHandler(w, r)
	assert.Equal(t, 401, w.Code, "Expected status code 401")
	assert.Equal(t, "Invalid expo account\n", w.Body.String(), "Expected error message")
}

func TestRequestUploadUrlWithBadBearer(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	mockExpoForRequestUploadUrlTest("staging")

	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Errorf("Error finding project root: %v", err)
	}
	os.Setenv("LOCAL_BUCKET_BASE_PATH", filepath.Join(projectRoot, "./updates"))
	q := "http://localhost:3000/requestUploadUrl/DO_NOT_USE?runtimeVersion=1"
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", q, nil)
	r = mux.SetURLVars(r, map[string]string{
		"BRANCH": "DO_NOT_USE",
	})
	r.Header.Set("Authorization", "Bearer expo_bad_token")
	sampleUpdatePath := filepath.Join(projectRoot, "/test/test-updates/branch-1/1/1674170951")
	uploadRequestsInput := ComputeUploadRequestsInput(sampleUpdatePath)
	uploadRequestsInputJSON, err := json.Marshal(uploadRequestsInput)
	if err != nil {
		t.Errorf("Error marshalling uploadRequestsInput: %v", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(uploadRequestsInputJSON))
	handlers.RequestUploadUrlHandler(w, r)
	assert.Equal(t, 401, w.Code, "Expected status code 401")
	assert.Equal(t, "Error fetching expo account informations\n", w.Body.String(), "Expected error message")
}

func TestRequestUploadUrlWithoutRuntimeVersion(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	mockExpoForRequestUploadUrlTest("staging")
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Errorf("Error finding project root: %v", err)
	}
	os.Setenv("LOCAL_BUCKET_BASE_PATH", filepath.Join(projectRoot, "./updates"))
	q := "http://localhost:3000/requestUploadUrl/DO_NOT_USE"
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", q, nil)
	r = mux.SetURLVars(r, map[string]string{
		"BRANCH": "DO_NOT_USE",
	})
	r.Header.Set("Authorization", "Bearer expo_test_token")
	sampleUpdatePath := filepath.Join(projectRoot, "/test/test-updates/branch-1/1/1674170951")
	uploadRequestsInput := ComputeUploadRequestsInput(sampleUpdatePath)
	uploadRequestsInputJSON, err := json.Marshal(uploadRequestsInput)
	if err != nil {
		t.Errorf("Error marshalling uploadRequestsInput: %v", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(uploadRequestsInputJSON))
	handlers.RequestUploadUrlHandler(w, r)
	assert.Equal(t, 400, w.Code, "Expected status code 400")
	assert.Equal(t, "No runtime version provided\n", w.Body.String(), "Expected error message")
}

func TestRequestUploadUrlWithBadRequestBody(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	mockExpoForRequestUploadUrlTest("staging")
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Errorf("Error finding project root: %v", err)
	}
	os.Setenv("LOCAL_BUCKET_BASE_PATH", filepath.Join(projectRoot, "./updates"))
	q := "http://localhost:3000/requestUploadUrl/DO_NOT_USE?runtimeVersion=1"
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", q, nil)
	r = mux.SetURLVars(r, map[string]string{
		"BRANCH": "DO_NOT_USE",
	})
	r.Header.Set("Authorization", "Bearer expo_test_token")
	uploadRequestsInputJSON, err := json.Marshal(map[string]string{
		"id": "4",
	})
	if err != nil {
		t.Errorf("Error marshalling uploadRequestsInput: %v", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(uploadRequestsInputJSON))
	handlers.RequestUploadUrlHandler(w, r)
	assert.Equal(t, 400, w.Code, "Expected status code 400")
	assert.Equal(t, "No file names provided\n", w.Body.String(), "Expected error message")
}

func TestRequestUploadUrlWithBadFilenamesType(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	mockExpoForRequestUploadUrlTest("staging")
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Errorf("Error finding project root: %v", err)
	}
	os.Setenv("LOCAL_BUCKET_BASE_PATH", filepath.Join(projectRoot, "./updates"))
	q := "http://localhost:3000/requestUploadUrl/DO_NOT_USE?runtimeVersion=1"
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", q, nil)
	r = mux.SetURLVars(r, map[string]string{
		"BRANCH": "DO_NOT_USE",
	})
	r.Header.Set("Authorization", "Bearer expo_test_token")
	uploadRequestsInputJSON, err := json.Marshal(map[string]int{
		"fileNames": 1,
	})
	if err != nil {
		t.Errorf("Error marshalling uploadRequestsInput: %v", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(uploadRequestsInputJSON))
	handlers.RequestUploadUrlHandler(w, r)
	assert.Equal(t, 400, w.Code, "Expected status code 400")
	assert.Equal(t, "Invalid JSON body\n", w.Body.String(), "Expected error message")
}

func TestRequestUploadUrlWithSampleUpdate(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	mockExpoForRequestUploadUrlTest("staging")
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Errorf("Error finding project root: %v", err)
	}
	os.Setenv("LOCAL_BUCKET_BASE_PATH", filepath.Join(projectRoot, "./updates"))
	q := "http://localhost:3000/requestUploadUrl/DO_NOT_USE?runtimeVersion=1"
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", q, nil)
	r = mux.SetURLVars(r, map[string]string{
		"BRANCH": "DO_NOT_USE",
	})
	r.Header.Set("Authorization", "Bearer expo_test_token")
	sampleUpdatePath := filepath.Join(projectRoot, "/test/test-updates/branch-4/1/1674170952")
	uploadRequestsInput := ComputeUploadRequestsInput(sampleUpdatePath)
	uploadRequestsInputJSON, err := json.Marshal(uploadRequestsInput)
	if err != nil {
		t.Errorf("Error marshalling uploadRequestsInput: %v", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(uploadRequestsInputJSON))
	handlers.RequestUploadUrlHandler(w, r)
	assert.Equal(t, 200, w.Code, "Expected status code 200")
	type ResponseBody struct {
		UpdateId       int64                      `json:"updateId"`
		UploadRequests []bucket.FileUploadRequest `json:"uploadRequests"`
	}
	var responseBody ResponseBody
	err = json.NewDecoder(w.Body).Decode(&responseBody)
	if err != nil {
		assert.Fail(t, "Expected valid JSON response")
	}
	fileUploadRequests := responseBody.UploadRequests
	assert.Len(t, fileUploadRequests, 4, "Expected 4 file upload requests")
	updateId := w.Header().Get("expo-update-id")
	assert.NotEmpty(t, updateId, "Expected non-empty update ID")
	for _, uploadRequest := range fileUploadRequests {
		requestUploadUrl := uploadRequest.RequestUploadUrl
		parsedUrl, err := url.Parse(requestUploadUrl)
		assert.Nil(t, err, "Expected valid URL")
		assert.Equal(t, "http", parsedUrl.Scheme, "Expected HTTP scheme")
		assert.Equal(t, "localhost:3000", parsedUrl.Host, "Expected localhost:3000 host")
		assert.Equal(t, "/uploadLocalFile", parsedUrl.Path, "Expected /uploadLocalFile path")
		token := parsedUrl.Query().Get("token")
		assert.NotEmpty(t, token, "Expected non-empty token")
		claims := jwt.MapClaims{}
		decoded, err := services.DecodeAndExtractJWTToken("test_jwt_secret", token, claims)
		assert.Nil(t, err, "Expected valid JWT token")
		if !decoded.Valid {
			assert.Fail(t, "Expected valid JWT token")
		}
		filePath := claims["filePath"].(string)
		assert.NotEmpty(t, filePath, "Expected non-empty file path")
		sub := claims["sub"].(string)
		assert.Equal(t, "test_username", sub, "Expected test_username sub")
	}
	var (
		ws   = make([]*httptest.ResponseRecorder, len(fileUploadRequests))
		errs = make(chan error, len(fileUploadRequests))
		wg   sync.WaitGroup
	)
	for i, uploadRequest := range fileUploadRequests {
		wg.Add(1)
		go func(index int, request bucket.FileUploadRequest) {
			defer wg.Done()
			ws[index] = httptest.NewRecorder()
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			fileBuffer, err := os.Open(projectRoot + "/test/test-updates/branch-4/1/1674170952/" + request.FilePath)
			if err != nil {
				errs <- err
				return
			}

			part, err := writer.CreateFormFile(request.FileName, request.FileName)
			if err != nil {
				errs <- err
				return
			}
			_, err = io.Copy(part, fileBuffer)
			if err != nil {
				errs <- err
				return
			}
			err = writer.Close()
			parsedUrl, err := url.Parse(uploadRequest.RequestUploadUrl)
			token := parsedUrl.Query().Get("token")
			r := httptest.NewRequest("PUT", "/uploadLocalFile?token="+token, body)
			r.Header.Set("Content-Type", writer.FormDataContentType())
			r.Header.Set("Authorization", "Bearer expo_test_token")
			handlers.RequestUploadLocalFileHandler(ws[index], r)
			if ws[index].Code != 200 {
				errs <- assert.AnError
			}

		}(i, uploadRequest)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		assert.Nil(t, err, "Expected no errors")
	}
	for _, w := range ws {
		assert.Equal(t, 200, w.Code, "Expected status code 200")
		_, err := os.Open(projectRoot + "/updates/DO_NOT_USE/1/" + updateId + "/" + fileUploadRequests[0].FilePath)
		assert.Nil(t, err, "Expected no errors")
	}
	lastUpdate, err := update.GetLatestUpdateBundlePathForRuntimeVersion("DO_NOT_USE", "1")
	if err != nil {
		t.Errorf("Error getting latest update: %v", err)
	}
	assert.Nil(t, lastUpdate, "Expected nil")
	q = "http://localhost:3000/markUpdateAsUploaded/DO_NOT_USE?platform=android&runtimeVersion=1&updateId=" + updateId
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", q, nil)
	r.Header.Set("Authorization", "Bearer expo_test_token")
	r = mux.SetURLVars(r, map[string]string{
		"BRANCH": "DO_NOT_USE",
	})
	handlers.MarkUpdateAsUploadedHandler(w, r)
	assert.Equal(t, 200, w.Code, "Expected status code 200")
	lastUpdate, err = update.GetLatestUpdateBundlePathForRuntimeVersion("DO_NOT_USE", "1")
	if err != nil {
		t.Errorf("Error getting latest update: %v", err)
	}
	assert.NotNil(t, lastUpdate, "Expected non-nil")
	assert.Equal(t, updateId, lastUpdate.UpdateId, "Expected updated ID")

}

func TestRequestUploadUrlWithValidExpoSession(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	mockExpoForRequestUploadUrlTest("staging")

	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Errorf("Error finding project root: %v", err)
	}
	os.Setenv("LOCAL_BUCKET_BASE_PATH", filepath.Join(projectRoot, "./updates"))
	q := "http://localhost:3000/requestUploadUrl/DO_NOT_USE?runtimeVersion=1"
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", q, nil)
	r = mux.SetURLVars(r, map[string]string{
		"BRANCH": "DO_NOT_USE",
	})
	r.Header.Set("expo-session", "expo_test_session")
	sampleUpdatePath := filepath.Join(projectRoot, "/test/test-updates/branch-1/1/1674170951")
	uploadRequestsInput := ComputeUploadRequestsInput(sampleUpdatePath)
	uploadRequestsInputJSON, err := json.Marshal(uploadRequestsInput)
	if err != nil {
		t.Errorf("Error marshalling uploadRequestsInput: %v", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(uploadRequestsInputJSON))
	handlers.RequestUploadUrlHandler(w, r)
	assert.Equal(t, 200, w.Code, "Expected status code 200")
	assert.NotEmpty(t, w.Header().Get("expo-update-id"), "Expected non-empty update ID")
}

func TestShouldClearCache(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Errorf("Error finding project root: %v", err)
	}
	os.Setenv("LOCAL_BUCKET_BASE_PATH", filepath.Join(projectRoot, "/test/test-updates"))
	mockWorkingExpoResponse("staging")
	qManifest := "http://localhost:3000/manifest"

	wManifest := httptest.NewRecorder()
	rManifest := httptest.NewRequest("GET", qManifest, nil)
	rManifest.Header.Add("expo-platform", "ios")
	rManifest.Header.Add("expo-runtime-version", "1")
	rManifest.Header.Add("expo-protocol-version", "1")
	rManifest.Header.Add("expo-expect-signature", "true")
	rManifest.Header.Add("expo-channel-name", "staging")
	handlers.ManifestHandler(wManifest, rManifest)
	assert.Equal(t, 200, wManifest.Code, "Expected status code 200")

	os.Setenv("LOCAL_BUCKET_BASE_PATH", filepath.Join(projectRoot, "./updates"))
	q := "http://localhost:3000/requestUploadUrl/branch-1?runtimeVersion=1"
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", q, nil)
	r = mux.SetURLVars(r, map[string]string{
		"BRANCH": "branch-1",
	})
	r.Header.Set("expo-session", "expo_test_session")
	sampleUpdatePath := filepath.Join(projectRoot, "/test/test-updates/branch-1/1/1674170951")
	cache := cache2.GetCache()
	cacheKey := update.ComputeLastUpdateCacheKey("branch-1", "1")
	value := cache.Get(cacheKey)
	assert.Equal(t, "{\"Branch\":\"branch-1\",\"RuntimeVersion\":\"1\",\"UpdateId\":\"1674170951\",\"CreatedAt\":1674170951000000}", value, "Expected a specific value")
	uploadRequestsInput := ComputeUploadRequestsInput(sampleUpdatePath)
	uploadRequestsInputJSON, err := json.Marshal(uploadRequestsInput)
	if err != nil {
		t.Errorf("Error marshalling uploadRequestsInput: %v", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(uploadRequestsInputJSON))
	handlers.RequestUploadUrlHandler(w, r)
	assert.Equal(t, 200, w.Code, "Expected status code 200")
	assert.NotEmpty(t, w.Header().Get("expo-update-id"), "Expected non-empty update ID")
	value = cache.Get(cacheKey)
	assert.Empty(t, value, "Expected an empty value")
}

func TestRequestUploadUrlWithInvalidExpoSession(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	mockExpoForRequestUploadUrlTest("staging")

	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Errorf("Error finding project root: %v", err)
	}
	os.Setenv("LOCAL_BUCKET_BASE_PATH", filepath.Join(projectRoot, "./updates"))
	q := "http://localhost:3000/requestUploadUrl/DO_NOT_USE?runtimeVersion=1"
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", q, nil)
	r = mux.SetURLVars(r, map[string]string{
		"BRANCH": "DO_NOT_USE",
	})
	r.Header.Set("expo-session", "invalid_session_token")
	sampleUpdatePath := filepath.Join(projectRoot, "/test/test-updates/branch-1/1/1674170951")
	uploadRequestsInput := ComputeUploadRequestsInput(sampleUpdatePath)
	uploadRequestsInputJSON, err := json.Marshal(uploadRequestsInput)
	if err != nil {
		t.Errorf("Error marshalling uploadRequestsInput: %v", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(uploadRequestsInputJSON))
	handlers.RequestUploadUrlHandler(w, r)
	assert.Equal(t, 401, w.Code, "Expected status code 401")
	assert.Equal(t, "Error fetching expo account informations\n", w.Body.String(), "Expected error message")
}
