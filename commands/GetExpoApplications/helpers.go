package main

import (
	"encoding/json"
	"expo-open-ota/internal/types"
	"os"
)

func dotExpoHomeDir() *string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return nil
	}
	var dirPath string
	staging := os.Getenv("EXPO_STAGING")
	local := os.Getenv("EXPO_LOCAL")
	if staging != "" {
		dirPath = homeDir + "/.expo-staging"
	} else if local != "" {
		dirPath = homeDir + "/.expo-local"
	} else {
		dirPath = homeDir + "/.expo"
	}
	return &dirPath
}

func getStateJsonPath() *string {
	homeDir := dotExpoHomeDir()
	if homeDir == nil {
		return nil
	}
	var stateJsonPath string
	stateJsonPath = *homeDir + "/state.json"
	return &stateJsonPath
}

type ExpoStateAuth struct {
	UserId            string `json:"userId"`
	Username          string `json:"username"`
	CurrentConnection string `json:"currentConnection"`
	SessionSecret     string `json:"sessionSecret"`
}

type ExpoState struct {
	PATH              string         `json:"PATH"`
	Auth              *ExpoStateAuth `json:"auth"`
	Uuid              string         `json:"uuid"`
	AnalyticsDeviceId string         `json:"analyticsDeviceId"`
}

func getExpoState() *ExpoState {
	stateJsonPath := getStateJsonPath()
	if stateJsonPath == nil {
		return nil
	}
	file, err := os.Open(*stateJsonPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var state ExpoState
	err = json.NewDecoder(file).Decode(&state)
	if err != nil {
		return nil
	}
	return &state
}

func retrieveExpoAuth() *types.ExpoAuth {
	token := os.Getenv("EXPO_ACCESS_TOKEN")
	if token != "" {
		return &types.ExpoAuth{
			Token: &token,
		}
	}

	state := getExpoState()
	if state == nil || state.Auth == nil {
		return nil
	}

	if state.Auth == nil {
		return &types.ExpoAuth{
			SessionSecret: nil,
		}
	}

	return &types.ExpoAuth{
		SessionSecret: &state.Auth.SessionSecret,
	}
}
