package handlers

import (
	"encoding/json"
	"expo-open-ota/config"
	"expo-open-ota/internal/bucket"
	cache2 "expo-open-ota/internal/cache"
	"expo-open-ota/internal/crypto"
	"expo-open-ota/internal/dashboard"
	"expo-open-ota/internal/services"
	"expo-open-ota/internal/types"
	update2 "expo-open-ota/internal/update"
	"github.com/gorilla/mux"
	"net/http"
	"sort"
	"strconv"
	"time"
)

type BranchMapping struct {
	BranchName     string  `json:"branchName"`
	ReleaseChannel *string `json:"releaseChannel"`
}

type UpdateItem struct {
	UpdateUUID string `json:"updateUUID"`
	UpdateId   string `json:"updateId"`
	CreatedAt  string `json:"createdAt"`
	CommitHash string `json:"commitHash"`
	Platform   string `json:"platform"`
}

type UpdateDetails struct {
	UpdateUUID string           `json:"updateUUID"`
	UpdateId   string           `json:"updateId"`
	CreatedAt  string           `json:"createdAt"`
	CommitHash string           `json:"commitHash"`
	Platform   string           `json:"platform"`
	Type       types.UpdateType `json:"type"`
	ExpoConfig string           `json:"expoConfig"`
}

type SettingsEnv struct {
	BASE_URL                               string `json:"BASE_URL"`
	EXPO_APP_ID                            string `json:"EXPO_APP_ID"`
	EXPO_ACCESS_TOKEN                      string `json:"EXPO_ACCESS_TOKEN"`
	CACHE_MODE                             string `json:"CACHE_MODE"`
	REDIS_HOST                             string `json:"REDIS_HOST"`
	REDIS_PORT                             string `json:"REDIS_PORT"`
	STORAGE_MODE                           string `json:"STORAGE_MODE"`
	S3_BUCKET_NAME                         string `json:"S3_BUCKET_NAME"`
	LOCAL_BUCKET_BASE_PATH                 string `json:"LOCAL_BUCKET_BASE_PATH"`
	KEYS_STORAGE_TYPE                      string `json:"KEYS_STORAGE_TYPE"`
	AWSSM_EXPO_PUBLIC_KEY_SECRET_ID        string `json:"AWSSM_EXPO_PUBLIC_KEY_SECRET_ID"`
	AWSSM_EXPO_PRIVATE_KEY_SECRET_ID       string `json:"AWSSM_EXPO_PRIVATE_KEY_SECRET_ID"`
	PUBLIC_EXPO_KEY_B64                    string `json:"PUBLIC_EXPO_KEY_B64"`
	PUBLIC_LOCAL_EXPO_KEY_PATH             string `json:"PUBLIC_LOCAL_EXPO_KEY_PATH"`
	PRIVATE_LOCAL_EXPO_KEY_PATH            string `json:"PRIVATE_LOCAL_EXPO_KEY_PATH"`
	AWS_REGION                             string `json:"AWS_REGION"`
	AWS_ACCESS_KEY_ID                      string `json:"AWS_ACCESS_KEY_ID"`
	CLOUDFRONT_DOMAIN                      string `json:"CLOUDFRONT_DOMAIN"`
	CLOUDFRONT_KEY_PAIR_ID                 string `json:"CLOUDFRONT_KEY_PAIR_ID"`
	CLOUDFRONT_PRIVATE_KEY_B64             string `json:"CLOUDFRONT_PRIVATE_KEY_B64"`
	AWSSM_CLOUDFRONT_PRIVATE_KEY_SECRET_ID string `json:"AWSSM_CLOUDFRONT_PRIVATE_KEY_SECRET_ID"`
	PRIVATE_LOCAL_CLOUDFRONT_KEY_PATH      string `json:"PRIVATE_LOCAL_CLOUDFRONT_KEY_PATH"`
	PROMETHEUS_ENABLED                     string `json:"PROMETHEUS_ENABLED"`
}

func GetSettingsHandler(w http.ResponseWriter, r *http.Request) {

	// Retrieve all in config.GetEnv & return as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SettingsEnv{
		BASE_URL:    config.GetEnv("BASE_URL"),
		EXPO_APP_ID: config.GetEnv("EXPO_APP_ID"),
		// Only retrieve the first 5 characters of the token
		EXPO_ACCESS_TOKEN:                      "***" + config.GetEnv("EXPO_ACCESS_TOKEN")[:5],
		CACHE_MODE:                             config.GetEnv("CACHE_MODE"),
		REDIS_HOST:                             config.GetEnv("REDIS_HOST"),
		REDIS_PORT:                             config.GetEnv("REDIS_PORT"),
		STORAGE_MODE:                           config.GetEnv("STORAGE_MODE"),
		S3_BUCKET_NAME:                         config.GetEnv("S3_BUCKET_NAME"),
		LOCAL_BUCKET_BASE_PATH:                 config.GetEnv("LOCAL_BUCKET_BASE_PATH"),
		KEYS_STORAGE_TYPE:                      config.GetEnv("KEYS_STORAGE_TYPE"),
		AWSSM_EXPO_PUBLIC_KEY_SECRET_ID:        config.GetEnv("AWSSM_EXPO_PUBLIC_KEY_SECRET_ID"),
		AWSSM_EXPO_PRIVATE_KEY_SECRET_ID:       config.GetEnv("AWSSM_EXPO_PRIVATE_KEY_SECRET_ID"),
		PUBLIC_EXPO_KEY_B64:                    config.GetEnv("PUBLIC_EXPO_KEY_B64"),
		PUBLIC_LOCAL_EXPO_KEY_PATH:             config.GetEnv("PUBLIC_LOCAL_EXPO_KEY_PATH"),
		PRIVATE_LOCAL_EXPO_KEY_PATH:            config.GetEnv("PRIVATE_LOCAL_EXPO_KEY_PATH"),
		AWS_REGION:                             config.GetEnv("AWS_REGION"),
		AWS_ACCESS_KEY_ID:                      config.GetEnv("AWS_ACCESS_KEY_ID"),
		CLOUDFRONT_DOMAIN:                      config.GetEnv("CLOUDFRONT_DOMAIN"),
		CLOUDFRONT_KEY_PAIR_ID:                 config.GetEnv("CLOUDFRONT_KEY_PAIR_ID"),
		CLOUDFRONT_PRIVATE_KEY_B64:             config.GetEnv("CLOUDFRONT_PRIVATE_KEY_B64"),
		AWSSM_CLOUDFRONT_PRIVATE_KEY_SECRET_ID: config.GetEnv("AWSSM_CLOUDFRONT_PRIVATE_KEY_SECRET_ID"),
		PRIVATE_LOCAL_CLOUDFRONT_KEY_PATH:      config.GetEnv("PRIVATE_LOCAL_CLOUDFRONT_KEY_PATH"),
		PROMETHEUS_ENABLED:                     config.GetEnv("PROMETHEUS_ENABLED"),
	})
}

func GetBranchesHandler(w http.ResponseWriter, r *http.Request) {
	resolvedBucket := bucket.GetBucket()
	branches, err := resolvedBucket.GetBranches()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cacheKey := dashboard.ComputeGetBranchesCacheKey()
	cache := cache2.GetCache()
	if cacheValue := cache.Get(cacheKey); cacheValue != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		var branches []BranchMapping
		json.Unmarshal([]byte(cacheValue), &branches)
		json.NewEncoder(w).Encode(branches)
		return
	}
	branchesMapping, err := services.FetchExpoBranchesMapping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var response []BranchMapping
	for _, branch := range branches {
		var releaseChannel *string
		for _, mapping := range branchesMapping {
			if mapping.BranchName == branch {
				releaseChannel = &mapping.ChannelName
				break
			}
		}
		response = append(response, BranchMapping{
			BranchName:     branch,
			ReleaseChannel: releaseChannel,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	marshaledResponse, _ := json.Marshal(response)
	cache.Set(cacheKey, string(marshaledResponse), nil)
}

func GetRuntimeVersionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	branchName := vars["BRANCH"]
	cacheKey := dashboard.ComputeGetRuntimeVersionsCacheKey(branchName)
	cache := cache2.GetCache()
	if cacheValue := cache.Get(cacheKey); cacheValue != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		var runtimeVersions []bucket.RuntimeVersionWithStats
		json.Unmarshal([]byte(cacheValue), &runtimeVersions)
		json.NewEncoder(w).Encode(runtimeVersions)
		return
	}
	resolvedBucket := bucket.GetBucket()
	runtimeVersions, err := resolvedBucket.GetRuntimeVersions(branchName)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	sort.Slice(runtimeVersions, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, runtimeVersions[i].CreatedAt)
		timeJ, _ := time.Parse(time.RFC3339, runtimeVersions[j].CreatedAt)
		return timeI.After(timeJ)
	})
	json.NewEncoder(w).Encode(runtimeVersions)
	marshaledResponse, _ := json.Marshal(runtimeVersions)
	cache.Set(cacheKey, string(marshaledResponse), nil)
}

func GetUpdateDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	branchName := vars["BRANCH"]
	runtimeVersion := vars["RUNTIME_VERSION"]
	updateId := vars["UPDATE_ID"]
	cacheKey := dashboard.ComputeGetUpdateDetailsCacheKey(branchName, runtimeVersion, updateId)
	cache := cache2.GetCache()
	if cacheValue := cache.Get(cacheKey); cacheValue != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		var updateDetailsResponse UpdateDetails
		json.Unmarshal([]byte(cacheValue), &updateDetailsResponse)
		json.NewEncoder(w).Encode(updateDetailsResponse)
		return
	}
	update, err := update2.GetUpdate(branchName, runtimeVersion, updateId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	metadata, err := update2.GetMetadata(*update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	numberUpdate, _ := strconv.ParseInt(update.UpdateId, 10, 64)
	commitHash, platform, _ := update2.RetrieveUpdateCommitHashAndPlatform(*update)
	expoConfig, err := update2.GetExpoConfig(*update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	updatesResponse := UpdateDetails{
		UpdateUUID: crypto.ConvertSHA256HashToUUID(metadata.ID),
		UpdateId:   update.UpdateId,
		CreatedAt:  time.UnixMilli(numberUpdate).UTC().Format(time.RFC3339),
		CommitHash: commitHash,
		Platform:   platform,
		Type:       update2.GetUpdateType(*update),
		ExpoConfig: string(expoConfig),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatesResponse)
	marshaledResponse, _ := json.Marshal(updatesResponse)
	cache.Set(cacheKey, string(marshaledResponse), nil)
}

func GetUpdatesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	branchName := vars["BRANCH"]
	runtimeVersion := vars["RUNTIME_VERSION"]
	cacheKey := dashboard.ComputeGetUpdatesCacheKey(branchName, runtimeVersion)
	cache := cache2.GetCache()
	if cacheValue := cache.Get(cacheKey); cacheValue != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		var updatesResponse []UpdateItem
		json.Unmarshal([]byte(cacheValue), &updatesResponse)
		json.NewEncoder(w).Encode(updatesResponse)
		return
	}
	resolvedBucket := bucket.GetBucket()
	updates, err := resolvedBucket.GetUpdates(branchName, runtimeVersion)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var updatesResponse []UpdateItem
	for _, update := range updates {
		isValid := update2.IsUpdateValid(update)
		if !isValid {
			continue
		}
		numberUpdate, _ := strconv.ParseInt(update.UpdateId, 10, 64)
		commitHash, platform, _ := update2.RetrieveUpdateCommitHashAndPlatform(update)
		updateType := update2.GetUpdateType(update)
		if updateType == types.Rollback {
			updatesResponse = append(updatesResponse, UpdateItem{
				UpdateUUID: "Rollback to embedded",
				UpdateId:   update.UpdateId,
				CreatedAt:  time.UnixMilli(numberUpdate).UTC().Format(time.RFC3339),
				CommitHash: commitHash,
				Platform:   platform,
			})
			continue
		}

		metadata, err := update2.GetMetadata(update)
		if err != nil {
			continue
		}
		updatesResponse = append(updatesResponse, UpdateItem{
			UpdateUUID: crypto.ConvertSHA256HashToUUID(metadata.ID),
			UpdateId:   update.UpdateId,
			CreatedAt:  time.UnixMilli(numberUpdate).UTC().Format(time.RFC3339),
			CommitHash: commitHash,
			Platform:   platform,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	sort.Slice(updatesResponse, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, updatesResponse[i].CreatedAt)
		timeJ, _ := time.Parse(time.RFC3339, updatesResponse[j].CreatedAt)
		return timeI.After(timeJ)
	})
	json.NewEncoder(w).Encode(updatesResponse)
	marshaledResponse, _ := json.Marshal(updatesResponse)
	ttl := 10 * time.Second
	ttlMs := int(ttl.Milliseconds())
	cache.Set(cacheKey, string(marshaledResponse), &ttlMs)
}
