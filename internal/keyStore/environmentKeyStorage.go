package keyStore

import (
	"encoding/base64"
	"expo-open-ota/config"
	"log"
)

type EnvironmentKeysStorage struct {
	publicExpoKeyBase64Key        string
	privateExpoKeyBase64Key       string
	privateCloudfrontKeyBase64Key string
}

func decodeKey(key string) string {
	if key == "" {
		return ""
	}
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		log.Printf("Failed to decode base64 key: %v", err)
		return ""
	}
	return string(decoded)
}

func (c *EnvironmentKeysStorage) GetPublicExpoKey() string {
	return decodeKey(config.GetEnv(c.publicExpoKeyBase64Key))
}

func (c *EnvironmentKeysStorage) GetPrivateExpoKey() string {
	return decodeKey(config.GetEnv(c.privateExpoKeyBase64Key))
}

func (c *EnvironmentKeysStorage) GetPrivateCloudfrontKey() string {
	return decodeKey(config.GetEnv(c.privateCloudfrontKeyBase64Key))
}
