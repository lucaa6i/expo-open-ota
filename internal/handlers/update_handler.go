package handlers

import (
	"encoding/json"
	"expo-open-ota/internal/services"
	"net/http"
)

func GetReleaseChannelsHandler(w http.ResponseWriter, r *http.Request) {
	channels, err := services.FetchExpoChannels()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	jsonChannels, err := json.Marshal(channels)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(jsonChannels)
	w.WriteHeader(http.StatusOK)

}
