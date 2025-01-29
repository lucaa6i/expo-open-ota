package certs

import (
	"expo-open-ota/config"
)

type EnvironmentCertsStorage struct {
	publicExpoCertKey        string
	privateExpoCertKey       string
	privateCloudfrontCertKey string
}

func (c *EnvironmentCertsStorage) GetPublicExpoCert() string {
	return config.GetEnv(c.publicExpoCertKey)
}

func (c *EnvironmentCertsStorage) GetPrivateExpoCert() string {
	return config.GetEnv(c.privateExpoCertKey)
}

func (c *EnvironmentCertsStorage) GetPrivateCloudfrontCert() string {
	return config.GetEnv(c.privateCloudfrontCertKey)
}
