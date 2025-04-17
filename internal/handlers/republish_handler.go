package handlers

import (
	"encoding/json"
	"expo-open-ota/internal/branch"
	update2 "expo-open-ota/internal/update"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func RepublishHandler(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	vars := mux.Vars(r)
	branchName := vars["BRANCH"]
	platform := r.URL.Query().Get("platform")
	if platform == "" || (platform != "ios" && platform != "android") {
		log.Printf("[RequestID: %s] Invalid platform: %s", requestID, platform)
		http.Error(w, "Invalid platform", http.StatusBadRequest)
		return
	}
	if branchName == "" {
		log.Printf("[RequestID: %s] No branch provided", requestID)
		http.Error(w, "No branch provided", http.StatusBadRequest)
		return
	}
	err := branch.UpsertBranch(branchName)
	if err != nil {
		log.Printf("[RequestID: %s] Error upserting branch: %v", requestID, err)
		http.Error(w, "Error upserting branch", http.StatusInternalServerError)
		return
	}
	runtimeVersion := r.URL.Query().Get("runtimeVersion")
	if runtimeVersion == "" {
		log.Printf("[RequestID: %s] No runtime version provided", requestID)
		http.Error(w, "No runtime version provided", http.StatusBadRequest)
		return
	}
	commitHash := r.URL.Query().Get("commitHash")
	updateId := r.URL.Query().Get("updateId")
	if updateId == "" {
		log.Printf("[RequestID: %s] No updateId provided", requestID)
		http.Error(w, "No updateId provided", http.StatusBadRequest)
		return
	}
	update, err := update2.GetUpdate(branchName, runtimeVersion, updateId)
	if err != nil {
		log.Printf("[RequestID: %s] Error getting update: %v", requestID, err)
		http.Error(w, "Error getting update", http.StatusInternalServerError)
		return
	}
	storedMetadata, err := update2.RetrieveUpdateStoredMetadata(*update)
	if err != nil {
		log.Printf("[RequestID: %s] Error retrieving update commit hash and platform: %v", requestID, err)
		http.Error(w, "Error retrieving update commit hash and platform", http.StatusInternalServerError)
		return
	}
	isValid := update2.IsUpdateValid(*update)
	if !isValid {
		log.Printf("[RequestID: %s] Update is not valid", requestID)
		http.Error(w, "Update is not valid", http.StatusBadRequest)
		return
	}
	if storedMetadata.Platform != platform {
		log.Printf("[RequestID: %s] Update platform mismatch: %s != %s", requestID, storedMetadata.Platform, platform)
		http.Error(w, "Update platform mismatch", http.StatusBadRequest)
		return
	}
	newUpdate, err := update2.RepublishUpdate(update, platform, commitHash)
	if err != nil {
		log.Printf("[RequestID: %s] Error republishing update: %v", requestID, err)
		http.Error(w, "Error republishing update", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newUpdate)
}
