package keyStore

import "expo-open-ota/internal/services"

type AWSSMKeysStorage struct {
	publicExpoKeySecretID        string
	privateExpoKeySecretID       string
	privateCloudfrontKeySecretID string
}

func (c *AWSSMKeysStorage) GetPublicExpoKey() string {
	if c.publicExpoKeySecretID == "" {
		return ""
	}
	return services.FetchSecret(c.publicExpoKeySecretID)
}

func (c *AWSSMKeysStorage) GetPrivateExpoKey() string {
	if c.privateExpoKeySecretID == "" {
		return ""
	}
	return services.FetchSecret(c.privateExpoKeySecretID)
}

func (c *AWSSMKeysStorage) GetPrivateCloudfrontKey() string {
	if c.privateCloudfrontKeySecretID == "" {
		return ""
	}
	return services.FetchSecret(c.privateCloudfrontKeySecretID)
}
