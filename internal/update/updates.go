package update

import (
	"encoding/json"
	"expo-open-ota/config"
	"expo-open-ota/internal/bucket"
	cache2 "expo-open-ota/internal/cache"
	"expo-open-ota/internal/crypto"
	"expo-open-ota/internal/types"
	"fmt"
	"mime"
	"net/url"
	"sort"
	"strings"
	"sync"
)

func sortUpdates(updates []types.Update) []types.Update {
	sort.Slice(updates, func(i, j int) bool {
		return updates[i].CreatedAt > updates[j].CreatedAt
	})
	return updates
}

func GetAllUpdatesForRuntimeVersion(branch string, runtimeVersion string) ([]types.Update, error) {
	resolvedBucket := bucket.GetBucket()
	updates, errGetUpdates := resolvedBucket.GetUpdates(branch, runtimeVersion)
	if errGetUpdates != nil {
		return nil, errGetUpdates
	}
	updates = sortUpdates(updates)
	return updates, nil
}

func IsUpdateValid(Update types.Update) bool {
	resolvedBucket := bucket.GetBucket()
	// Search for .check file in the update
	file, _ := resolvedBucket.GetFile(Update, ".check")
	if file.Reader != nil {
		defer file.Reader.Close()
		return true
	}
	_, err := resolvedBucket.GetFile(Update, "rollback")
	if err != nil {
		meta, err := GetMetadata(Update)
		if err != nil {
			return false
		}
		if meta.MetadataJSON.FileMetadata.IOS.Bundle == "" || meta.MetadataJSON.FileMetadata.Android.Bundle == "" {
			return false
		}
		files := []string{meta.MetadataJSON.FileMetadata.IOS.Bundle, meta.MetadataJSON.FileMetadata.Android.Bundle}
		for _, asset := range meta.MetadataJSON.FileMetadata.IOS.Assets {
			files = append(files, asset.Path)
		}
		for _, asset := range meta.MetadataJSON.FileMetadata.Android.Assets {
			files = append(files, asset.Path)
		}

		for _, file := range files {
			_, err := resolvedBucket.GetFile(Update, file)
			if err != nil {
				return false
			}
		}
	}
	reader := strings.NewReader(".check")
	_ = resolvedBucket.UploadFileIntoUpdate(Update, ".check", reader)

	return true
}

func ComputeLastUpdateCacheKey(branch string, runtimeVersion string) string {
	return fmt.Sprintf("lastUpdate:%s:%s", branch, runtimeVersion)
}

func GetLatestUpdateBundlePathForRuntimeVersion(branch string, runtimeVersion string) (*types.Update, error) {
	cache := cache2.GetCache()
	cacheKey := fmt.Sprintf(ComputeLastUpdateCacheKey(branch, runtimeVersion))
	if cachedValue := cache.Get(cacheKey); cachedValue != "" {
		var update types.Update
		err := json.Unmarshal([]byte(cachedValue), &update)
		if err != nil {
			return nil, err
		}
		return &update, nil
	}
	updates, err := GetAllUpdatesForRuntimeVersion(branch, runtimeVersion)
	if err != nil {
		return nil, err
	}
	filteredUpdates := make([]types.Update, 0)
	for _, update := range updates {
		if IsUpdateValid(update) {
			filteredUpdates = append(filteredUpdates, update)
		}
	}
	if len(filteredUpdates) > 0 {
		cacheValue, err := json.Marshal(filteredUpdates[0])
		if err != nil {
			return nil, err
		}
		ttl := 1800
		err = cache.Set(cacheKey, string(cacheValue), &ttl)
		return &filteredUpdates[0], nil
	}
	return nil, nil
}

func GetUpdateType(update types.Update) types.UpdateType {
	resolvedBucket := bucket.GetBucket()
	file, err := resolvedBucket.GetFile(update, "rollback")
	if err == nil && file.Reader != nil {
		defer file.Reader.Close()
		return types.Rollback
	}
	return types.NormalUpdate
}

func GetExpoConfig(update types.Update) (json.RawMessage, error) {
	resolvedBucket := bucket.GetBucket()
	resp, err := resolvedBucket.GetFile(update, "expoConfig.json")
	if err != nil {
		return nil, err
	}
	defer resp.Reader.Close()
	var expoConfig json.RawMessage
	err = json.NewDecoder(resp.Reader).Decode(&expoConfig)
	if err != nil {
		return nil, err
	}
	return expoConfig, nil
}

func GetMetadata(update types.Update) (types.UpdateMetadata, error) {
	resolvedBucket := bucket.GetBucket()
	file, errFile := resolvedBucket.GetFile(update, "metadata.json")
	if errFile != nil {
		return types.UpdateMetadata{}, errFile
	}

	createdAt := file.CreatedAt
	var metadata types.UpdateMetadata
	var metadataJson types.MetadataObject
	err := json.NewDecoder(file.Reader).Decode(&metadataJson)
	defer file.Reader.Close()
	if err != nil {
		fmt.Println("error decoding metadata json:", err)
		return types.UpdateMetadata{}, err
	}
	metadata.CreatedAt = createdAt.UTC().Format("2006-01-02T15:04:05.000Z")
	metadata.MetadataJSON = metadataJson
	stringifiedMetadata, err := json.Marshal(metadata.MetadataJSON)
	if err != nil {
		return types.UpdateMetadata{}, err
	}
	id, errHash := crypto.CreateHash(stringifiedMetadata, "sha256", "hex")

	if errHash != nil {
		return types.UpdateMetadata{}, errHash
	}

	metadata.ID = id
	return metadata, nil
}

func BuildFinalManifestAssetUrlURL(baseURL, assetFilePath, runtimeVersion, platform string) (string, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	query := url.Values{}
	query.Set("asset", assetFilePath)
	query.Set("runtimeVersion", runtimeVersion)
	query.Set("platform", platform)
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

func GetAssetEndpoint() string {
	return config.GetEnv("BASE_URL") + "/assets"
}

func shapeManifestAsset(update types.Update, asset *types.Asset, isLaunchAsset bool, platform string) (types.ManifestAsset, error) {
	resolvedBucket := bucket.GetBucket()
	assetFilePath := asset.Path
	assetFile, errAssetFile := resolvedBucket.GetFile(update, asset.Path)
	if errAssetFile != nil {
		return types.ManifestAsset{}, errAssetFile
	}

	byteAsset, errAsset := bucket.ConvertReadCloserToBytes(assetFile.Reader)
	defer assetFile.Reader.Close()
	if errAsset != nil {
		return types.ManifestAsset{}, errAsset
	}
	assetHash, errHash := crypto.CreateHash(byteAsset, "sha256", "base64")
	if errHash != nil {
		return types.ManifestAsset{}, errHash
	}
	urlEncodedHash := crypto.GetBase64URLEncoding(assetHash)
	key, errKey := crypto.CreateHash(byteAsset, "md5", "hex")
	if errKey != nil {
		return types.ManifestAsset{}, errKey
	}

	keyExtensionSuffix := asset.Ext
	if isLaunchAsset {
		keyExtensionSuffix = "bundle"
	}
	keyExtensionSuffix = "." + keyExtensionSuffix
	contentType := "application/javascript"
	if isLaunchAsset {
		contentType = mime.TypeByExtension(asset.Ext)
	}
	finalUrl, errUrl := BuildFinalManifestAssetUrlURL(GetAssetEndpoint(), assetFilePath, update.RuntimeVersion, platform)
	if errUrl != nil {
		return types.ManifestAsset{}, errUrl
	}
	return types.ManifestAsset{
		Hash:          urlEncodedHash,
		Key:           key,
		FileExtension: keyExtensionSuffix,
		ContentType:   contentType,
		Url:           finalUrl,
	}, nil
}

func ComposeUpdateManifest(
	metadata *types.UpdateMetadata,
	update types.Update,
	platform string,
) (types.UpdateManifest, error) {
	expoConfig, errConfig := GetExpoConfig(update)
	if errConfig != nil {
		return types.UpdateManifest{}, errConfig
	}

	var platformSpecificMetadata types.PlatformMetadata
	switch platform {
	case "ios":
		platformSpecificMetadata = metadata.MetadataJSON.FileMetadata.IOS
	case "android":
		platformSpecificMetadata = metadata.MetadataJSON.FileMetadata.Android
	}
	var (
		assets = make([]types.ManifestAsset, len(platformSpecificMetadata.Assets))
		errs   = make(chan error, len(platformSpecificMetadata.Assets))
		wg     sync.WaitGroup
	)

	for i, a := range platformSpecificMetadata.Assets {
		wg.Add(1)
		go func(index int, asset types.Asset) {
			defer wg.Done()
			shapedAsset, errShape := shapeManifestAsset(update, &asset, false, platform)
			if errShape != nil {
				errs <- errShape
				return
			}
			assets[index] = shapedAsset
		}(i, a)
	}

	wg.Wait()
	close(errs)

	if len(errs) > 0 {
		return types.UpdateManifest{}, <-errs
	}

	launchAsset, errShape := shapeManifestAsset(update, &types.Asset{
		Path: platformSpecificMetadata.Bundle,
		Ext:  "",
	}, true, platform)
	if errShape != nil {
		return types.UpdateManifest{}, errShape
	}

	manifest := types.UpdateManifest{
		Id:             crypto.ConvertSHA256HashToUUID(metadata.ID),
		CreatedAt:      metadata.CreatedAt,
		RunTimeVersion: update.RuntimeVersion,
		Metadata:       json.RawMessage("{}"),
		Extra: types.ExtraManifestData{
			ExpoClient: expoConfig,
		},
		Assets:      assets,
		LaunchAsset: launchAsset,
	}
	return manifest, nil
}

func CreateRollbackDirective(update types.Update) (types.RollbackDirective, error) {
	resolvedBucket := bucket.GetBucket()
	object, err := resolvedBucket.GetFile(update, "rollback")
	if err != nil {
		return types.RollbackDirective{}, err
	}
	commitTime := object.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z")
	defer object.Reader.Close()
	return types.RollbackDirective{
		Type: "rollBackToEmbedded",
		Parameters: types.RollbackDirectiveParameters{
			CommitTime: commitTime,
		},
	}, nil
}

func CreateNoUpdateAvailableDirective() types.NoUpdateAvailableDirective {
	return types.NoUpdateAvailableDirective{
		Type: "noUpdateAvailable",
	}
}
