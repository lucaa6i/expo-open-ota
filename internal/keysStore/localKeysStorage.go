package keys

import (
	"fmt"
	"io"
	"os"
)

type LocalKeysStorage struct {
	privateExpoKeyPath       string
	publicExpoKeyPath        string
	privateCloudfrontKeyPath string
}

func retrieveFileContent(path string) string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ""
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return ""
	}

	return string(content)
}
func (c *LocalKeysStorage) GetPublicExpoKey() string {
	if c.publicExpoKeyPath == "" {
		return ""
	}
	return retrieveFileContent(c.publicExpoKeyPath)
}

func (c *LocalKeysStorage) GetPrivateExpoKey() string {
	if c.privateExpoKeyPath == "" {
		return ""
	}
	private := retrieveFileContent(c.privateExpoKeyPath)
	return private
}

func (c *LocalKeysStorage) GetPrivateCloudfrontKey() string {
	if c.privateCloudfrontKeyPath == "" {
		return ""
	}
	return retrieveFileContent(c.privateCloudfrontKeyPath)
}
