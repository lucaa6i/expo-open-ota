package certs

import (
	"expo-open-ota/config"
)

type EnvironmentCertsStorage struct {
	publicExpoCertBase64Key        string
	privateExpoCertBase64Key       string
	privateCloudfrontCertBase64Key string
}

func (c *EnvironmentCertsStorage) GetPublicExpoCert() string {
	return config.GetEnv(c.publicExpoCertBase64Key)
}

func (c *EnvironmentCertsStorage) GetPrivateExpoCert() string {
	return config.GetEnv(c.publicExpoCertBase64Key)
}

func (c *EnvironmentCertsStorage) GetPrivateCloudfrontCert() string {
	return config.GetEnv(c.publicExpoCertBase64Key)
}
