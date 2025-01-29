package helpers

import (
	"expo-open-ota/internal/types"
	"net/http"
	"strings"
)

func GetExpoAuth(r *http.Request) types.ExpoAuth {
	bearerToken, _ := GetBearerToken(r)
	if bearerToken != "" {
		return types.ExpoAuth{
			Token: &bearerToken,
		}
	}
	sessionSecret := r.Header.Get("expo-session")
	if sessionSecret != "" {
		return types.ExpoAuth{
			SessionSecret: &sessionSecret,
		}
	}
	return types.ExpoAuth{}
}

func GetBearerToken(r *http.Request) (string, error) {
	bearerToken := r.Header.Get("Authorization")
	if bearerToken == "" {
		return "", nil
	}
	tokens := strings.Split(bearerToken, "Bearer ")
	if len(tokens) != 2 {
		return "", nil
	}
	return tokens[1], nil
}
