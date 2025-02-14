package handlers

import (
	"encoding/json"
	"expo-open-ota/internal/bucket"
	cache2 "expo-open-ota/internal/cache"
	"expo-open-ota/internal/crypto"
	"expo-open-ota/internal/dashboard"
	"expo-open-ota/internal/metrics"
	"expo-open-ota/internal/services"
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
	UpdateUUID     string `json:"updateUUID"`
	UpdateId       string `json:"updateId"`
	CreatedAt      string `json:"createdAt"`
	CommitHash     string `json:"commitHash"`
	Platform       string `json:"platform"`
	ActiveUsers    int    `json:"activeUsers"`
	TotalDownloads int    `json:"totalDownloads"`
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
		w.WriteHeader(http.StatusInternalServerError)
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
		metadata, err := update2.GetMetadata(update)
		if err != nil {
			continue
		}
		numberUpdate, _ := strconv.ParseInt(update.UpdateId, 10, 64)
		commitHash, platform, _ := update2.RetrieveUpdateCommitHashAndPlatform(update)
		updatesResponse = append(updatesResponse, UpdateItem{
			UpdateUUID:     crypto.ConvertSHA256HashToUUID(metadata.ID),
			UpdateId:       update.UpdateId,
			CreatedAt:      time.UnixMilli(numberUpdate).Format(time.RFC3339),
			CommitHash:     commitHash,
			Platform:       platform,
			ActiveUsers:    metrics.GetActiveUsers(runtimeVersion, branchName, update.UpdateId),
			TotalDownloads: metrics.GetTotalUpdateDownloadsByUpdate(runtimeVersion, branchName, update.UpdateId),
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
	cache.Set(cacheKey, string(marshaledResponse), nil)
}
