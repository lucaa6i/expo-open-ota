package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"expo-open-ota/config"
	"expo-open-ota/internal/types"
	"net/http"
)

type ExpoUserAccount struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type ExpoChannelMapping struct {
	Id         string `json:"id"`
	BranchName string `json:"branchName"`
}

type ExpoBranchMapping struct {
	BranchName  string `json:"branchName"`
	ChannelName string `json:"channelName"`
}

type BranchMapping struct {
	Version int `json:"version"`
	Data    []struct {
		BranchId           string          `json:"branchId"`
		BranchMappingLogic json.RawMessage `json:"branchMappingLogic"`
	}
}

func ValidateExpoAuth(expoAuth types.ExpoAuth) (*ExpoUserAccount, error) {
	if expoAuth.Token == nil && expoAuth.SessionSecret == nil {
		return nil, errors.New("no valid Expo auth provided")
	}
	expoAccount, err := FetchExpoUserAccountInformations(expoAuth)
	if err != nil {
		return nil, err
	}
	if expoAccount == nil {
		return nil, errors.New("no expo account found")
	}
	selfExpoUsername := FetchSelfExpoUsername()
	if selfExpoUsername != expoAccount.Username {
		return nil, errors.New("expo account does not match self expo username")
	}
	return expoAccount, nil
}

func GetExpoAccessToken() string {
	return config.GetEnv("EXPO_ACCESS_TOKEN")
}

func GetExpoAppId() string {
	return config.GetEnv("EXPO_APP_ID")
}

func SetAuthHeaders(expoAuth types.ExpoAuth, req *http.Request) {
	if expoAuth.Token != nil {
		req.Header.Set("Authorization", "Bearer "+*expoAuth.Token)
	}
	if expoAuth.SessionSecret != nil {
		req.Header.Set("expo-session", *expoAuth.SessionSecret)
	}
}

func makeGraphQLRequest(ctx context.Context, query string, variables map[string]interface{}, expoAuth types.ExpoAuth, result interface{}, headers map[string]string) error {
	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.expo.dev/graphql", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	SetAuthHeaders(expoAuth, req)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("GraphQL request failed with status: " + resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func UpdateChannelBranchMapping(channelName, branchName string) error {
	query := `
		query UpdateChannelBranchMapping($channelId: ID!, $branchMapping: String!) {
			updateChannel {
				editUpdateChannel(channelId: $channelId, branchMapping: $branchMapping) {
					id
				}
			}
		}
	`
	branchMapping := BranchMapping{
		Version: 0,
		Data: []struct {
			BranchId           string          `json:"branchId"`
			BranchMappingLogic json.RawMessage `json:"branchMappingLogic"`
		}{
			{
				BranchId:           branchName,
				BranchMappingLogic: json.RawMessage(`"true"`),
			},
		},
	}
	variables := map[string]interface{}{
		"channelId":     channelName,
		"branchMapping": branchMapping,
	}
	token := GetExpoAccessToken()
	headers := map[string]string{}
	if config.IsTestMode() {
		headers["operationName"] = "UpdateChannelBranchMapping"
	}
	ctx := context.Background()
	resp := struct{}{}
	return makeGraphQLRequest(ctx, query, variables, types.ExpoAuth{
		Token: &token,
	}, &resp, headers)
}

func FetchExpoBranches() ([]string, error) {
	query := `
		query FetchAppChannel($appId: String!) {
			app {
				byId(appId: $appId) {
					id
					updateBranches(offset: 0, limit: 10000) {
						id
						name
					}
				}
			}
		}
	`
	appId := GetExpoAppId()
	expoToken := GetExpoAccessToken()
	variables := map[string]interface{}{
		"appId": appId,
	}
	var resp struct {
		Data struct {
			App struct {
				ById struct {
					UpdateBranches []struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"updateBranches"`
				} `json:"byId"`
			} `json:"app"`
		} `json:"data"`
	}
	headers := map[string]string{}
	if config.IsTestMode() {
		headers["operationName"] = "FetchExpoBranches"
	}
	ctx := context.Background()
	if err := makeGraphQLRequest(ctx, query, variables, types.ExpoAuth{
		Token: &expoToken,
	}, &resp, headers); err != nil {
		return nil, err
	}
	var branches []string
	for _, branch := range resp.Data.App.ById.UpdateBranches {
		branches = append(branches, branch.Name)
	}
	return branches, nil
}

func FetchExpoUserAccountInformations(expoAuth types.ExpoAuth) (*ExpoUserAccount, error) {
	query := `
		query GetCurrentUserAccount {
			me {
				id
				username
				email
			}
		}
	`

	var resp struct {
		Data struct {
			Me ExpoUserAccount `json:"me"`
		} `json:"data"`
	}

	headers := map[string]string{}
	if config.IsTestMode() {
		headers["operationName"] = "FetchExpoUserAccountInformations"
	}

	ctx := context.Background()
	if err := makeGraphQLRequest(ctx, query, nil, expoAuth, &resp, headers); err != nil {
		return nil, err
	}

	return &resp.Data.Me, nil
}

func FetchSelfExpoUsername() string {
	token := GetExpoAccessToken()
	expoAccount, err := FetchExpoUserAccountInformations(types.ExpoAuth{
		Token: &token,
	})
	if err != nil {
		return ""
	}
	return expoAccount.Username
}

func FetchExpoChannelMapping(channelName string) (*ExpoChannelMapping, error) {
	query := `
		query FetchAppChannel($appId: String!, $channelName: String!) {
			app {
				byId(appId: $appId) {
					id
					updateBranches(offset: 0, limit: 10000) {
						id
						name
					}
					updateChannelByName(name: $channelName) {
						id
						name
						branchMapping
					}
				}
			}
		}
	`

	appId := GetExpoAppId()
	expoToken := GetExpoAccessToken()
	variables := map[string]interface{}{
		"appId":       appId,
		"channelName": channelName,
	}

	var resp struct {
		Data struct {
			App struct {
				ById struct {
					UpdateBranches []struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"updateBranches"`
					UpdateChannelByName struct {
						ID            string `json:"id"`
						BranchMapping string `json:"branchMapping"`
					} `json:"updateChannelByName"`
				} `json:"byId"`
			} `json:"app"`
		} `json:"data"`
	}

	headers := map[string]string{}
	if config.IsTestMode() {
		headers["operationName"] = "FetchExpoChannelMapping"
	}
	ctx := context.Background()
	if err := makeGraphQLRequest(ctx, query, variables, types.ExpoAuth{Token: &expoToken}, &resp, headers); err != nil {
		return nil, err
	}

	var branchMapping BranchMapping
	if err := json.Unmarshal([]byte(resp.Data.App.ById.UpdateChannelByName.BranchMapping), &branchMapping); err != nil {
		return nil, err
	}

	var branchID string
	for _, mapping := range branchMapping.Data {
		var logic string
		if json.Unmarshal(mapping.BranchMappingLogic, &logic) == nil && logic == "true" {
			branchID = mapping.BranchId
			break
		}
	}
	if branchID == "" {
		return nil, nil
	}

	var branchName string
	for _, branch := range resp.Data.App.ById.UpdateBranches {
		if branch.ID == branchID {
			branchName = branch.Name
			break
		}
	}
	if branchName == "" {
		return nil, nil
	}

	return &ExpoChannelMapping{
		Id:         resp.Data.App.ById.UpdateChannelByName.ID,
		BranchName: branchName,
	}, nil
}

func FetchExpoBranchesMapping() ([]ExpoBranchMapping, error) {
	query := `
		query FetchAppChannel($appId: String!) {
			app {
				byId(appId: $appId) {
					id
					updateBranches(offset: 0, limit: 10000) {
						id
						name
					}
					updateChannels(offset: 0, limit: 10000) {
                		id
                		name
                		branchMapping
            		}
				}
			}
		}
	`
	appId := GetExpoAppId()
	expoToken := GetExpoAccessToken()
	variables := map[string]interface{}{
		"appId": appId,
	}
	var resp struct {
		Data struct {
			App struct {
				ById struct {
					UpdateBranches []struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"updateBranches"`
					UpdateChannels []struct {
						ID            string `json:"id"`
						Name          string `json:"name"`
						BranchMapping string `json:"branchMapping"`
					} `json:"updateChannels"`
				} `json:"byId"`
			} `json:"app"`
		} `json:"data"`
	}
	headers := map[string]string{}
	if config.IsTestMode() {
		headers["operationName"] = "FetchExpoBranches"
	}
	ctx := context.Background()
	if err := makeGraphQLRequest(ctx, query, variables, types.ExpoAuth{
		Token: &expoToken,
	}, &resp, headers); err != nil {
		return nil, err
	}
	var branchMappings []ExpoBranchMapping
	for _, channel := range resp.Data.App.ById.UpdateChannels {
		var branchMapping BranchMapping
		if err := json.Unmarshal([]byte(channel.BranchMapping), &branchMapping); err != nil {
			return nil, err
		}
		var branchID string
		for _, mapping := range branchMapping.Data {
			var logic string
			if json.Unmarshal(mapping.BranchMappingLogic, &logic) == nil && logic == "true" {
				branchID = mapping.BranchId
				break
			}
		}
		if branchID == "" {
			continue
		}
		var branchName string
		for _, branch := range resp.Data.App.ById.UpdateBranches {
			if branch.ID == branchID {
				branchName = branch.Name
				break
			}
		}
		if branchName == "" {
			continue
		}
		branchMappings = append(branchMappings, ExpoBranchMapping{
			BranchName:  branchName,
			ChannelName: channel.Name,
		})
	}
	return branchMappings, nil
}

func CreateBranch(branch string) error {
	query := `
		mutation CreateUpdateBranchForAppMutation($appId: ID!, $name: String!) {
		  updateBranch {
			createUpdateBranchForApp(appId: $appId, name: $name) {
			  id
			}
		  }
		}
	`
	appId := GetExpoAppId()
	variables := map[string]interface{}{
		"appId": appId,
		"name":  branch,
	}
	token := GetExpoAccessToken()
	headers := map[string]string{}
	if config.IsTestMode() {
		headers["operationName"] = "CreateBranch"
	}
	ctx := context.Background()
	resp := struct{}{}
	return makeGraphQLRequest(ctx, query, variables, types.ExpoAuth{
		Token: &token,
	}, &resp, headers)
}

type ExpoApp struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
