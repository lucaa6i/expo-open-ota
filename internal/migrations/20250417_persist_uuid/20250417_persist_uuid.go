package _0250417_persist_uuid

import (
	"encoding/json"
	"expo-open-ota/internal/bucket"
	"expo-open-ota/internal/crypto"
	"expo-open-ota/internal/migration"
	"expo-open-ota/internal/types"
	update2 "expo-open-ota/internal/update"
	"fmt"
	"strings"
	"time"
)

func init() {
	migration.Register(migration.BaseMigration{
		Id:   "20250417_persist_uuid",
		Time: time.Date(2025, 4, 17, 0, 0, 0, 0, time.UTC),
		UpFunc: func(b bucket.Bucket) error {
			branches, err := b.GetBranches()
			if err != nil {
				return err
			}
			if len(branches) == 0 {
				return nil
			}
			for _, branch := range branches {
				runtimeVersions, err := b.GetRuntimeVersions(branch)
				if err != nil {
					continue
				}
				for _, runtimeVersion := range runtimeVersions {
					updates, err := b.GetUpdates(branch, runtimeVersion.RuntimeVersion)
					if err != nil {
						continue
					}
					for _, update := range updates {
						fmt.Println("Processing update:", update.UpdateId)
						storedMetadata, err := update2.RetrieveUpdateStoredMetadata(update)
						if err == nil && storedMetadata.UpdateUUID != "" {
							fmt.Println("Update UUID already exists, skipping:", update.UpdateId)
							continue
						}
						var metadata types.UpdateMetadata
						var metadataJson types.MetadataObject
						file, _ := b.GetFile(update, "metadata.json")
						if file == nil {
							continue
						}
						createdAt := file.CreatedAt
						err = json.NewDecoder(file.Reader).Decode(&metadataJson)
						defer file.Reader.Close()
						if err != nil {
							fmt.Println("error decoding metadata json:", err)
							continue
						}

						metadata.CreatedAt = createdAt.UTC().Format("2006-01-02T15:04:05.000Z")
						metadata.MetadataJSON = metadataJson
						stringifiedMetadata, err := json.Marshal(metadata.MetadataJSON)
						hashInput := string(stringifiedMetadata) + "::" + update.Branch + "::" + update.RuntimeVersion
						id, errHash := crypto.CreateHash([]byte(hashInput), "sha256", "hex")
						if errHash != nil {
							fmt.Println("error creating hash:", errHash)
							continue
						}
						updateUUID := crypto.ConvertSHA256HashToUUID(id)
						if updateUUID == "" {
							fmt.Println("error converting hash to UUID")
							continue
						}
						updateMetadataFile, _ := b.GetFile(update, "update-metadata.json")
						defer file.Reader.Close()
						storedMetadata = &types.UpdateStoredMetadata{}
						if updateMetadataFile == nil {
							err = json.NewDecoder(file.Reader).Decode(&storedMetadata)
							if err != nil {
								fmt.Println("error decoding update-metadata.json:", err)
								continue
							}
						}
						storedMetadata.UpdateUUID = updateUUID
						updatedMetadata, err := json.Marshal(storedMetadata)
						if err != nil {
							fmt.Println("error marshaling update-metadata.json:", err)
							continue
						}
						reader := strings.NewReader(string(updatedMetadata))
						err = b.UploadFileIntoUpdate(update, "update-metadata.json", reader)
						if err != nil {
							fmt.Println("error uploading update-metadata.json:", err)
							continue
						}
					}
				}
			}
			return nil
		},
		DownFunc: func(b bucket.Bucket) error {
			return nil
		},
	})
}
