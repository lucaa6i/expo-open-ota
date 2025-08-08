package bucket

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"expo-open-ota/config"
	"expo-open-ota/internal/types"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type GCSBucket struct {
	BucketName string
	AccessKey  string
	SecretKey  string
	BaseURL    string
}

// ListBucketResult represents the XML response from GCS ListObjects
type ListBucketResult struct {
	XMLName        xml.Name       `xml:"ListBucketResult"`
	Contents       []Object       `xml:"Contents"`
	CommonPrefixes []CommonPrefix `xml:"CommonPrefixes"`
	IsTruncated    bool           `xml:"IsTruncated"`
	NextMarker     string         `xml:"NextMarker"`
}

type Object struct {
	Key          string    `xml:"Key"`
	LastModified time.Time `xml:"LastModified"`
	Size         int64     `xml:"Size"`
}

type CommonPrefix struct {
	Prefix string `xml:"Prefix"`
}

func NewGCSBucket() *GCSBucket {
	return &GCSBucket{
		BucketName: config.GetEnv("S3_BUCKET_NAME"),
		AccessKey:  config.GetEnv("AWS_ACCESS_KEY_ID"),
		SecretKey:  config.GetEnv("AWS_SECRET_ACCESS_KEY"),
		BaseURL:    "https://storage.googleapis.com",
	}
}

// generateSignature creates an AWS signature v2 for GCS compatibility
func (b *GCSBucket) generateSignature(method, resource string, contentType string, date time.Time) (string, error) {
	if b.AccessKey == "" || b.SecretKey == "" {
		return "", errors.New("access key and secret key must be set")
	}

	dateStr := date.Format("Mon, 02 Jan 2006 15:04:05 GMT")

	// Create string to sign for AWS signature v2
	// Format: HTTP-Verb + "\n" + Content-MD5 + "\n" + Content-Type + "\n" + Date + "\n" + CanonicalizedAmzHeaders + CanonicalizedResource
	contentMD5 := ""
	if contentType == "" {
		contentType = ""
	}
	
	// Build canonicalized AMZ headers (none for basic GCS usage)
	canonicalizedAmzHeaders := ""
	
	stringToSign := method + "\n" + contentMD5 + "\n" + contentType + "\n" + dateStr + "\n" + canonicalizedAmzHeaders + resource


	// Calculate HMAC-SHA1 signature (GCS expects SHA1, not SHA256)
	h := hmac.New(sha1.New, []byte(b.SecretKey))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Create authorization header (AWS signature v2 style)
	authHeader := fmt.Sprintf("AWS %s:%s", b.AccessKey, signature)
	return authHeader, nil
}

func (b *GCSBucket) makeRequest(method, path string, body io.Reader) (*http.Response, error) {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("error reading body: %w", err)
		}
	}

	now := time.Now().UTC()
	dateStr := now.Format("Mon, 02 Jan 2006 15:04:05 GMT")

	// Determine content type
	contentType := ""
	if body != nil {
		contentType = "application/octet-stream"
	}

	// For GCS with AWS signature v2, the canonical resource is just the resource path without query parameters
	var canonicalResource string
	if strings.Contains(path, "?") {
		parts := strings.SplitN(path, "?", 2)
		canonicalResource = parts[0]
	} else {
		canonicalResource = path
	}

	authHeader, err := b.generateSignature(method, canonicalResource, contentType, now)
	if err != nil {
		return nil, fmt.Errorf("error generating signature: %w", err)
	}

	fullURL := b.BaseURL + path
	req, err := http.NewRequest(method, fullURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Date", dateStr)
	req.Header.Set("Host", "storage.googleapis.com")
	req.Header.Set("Authorization", authHeader)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	return client.Do(req)
}

func (b *GCSBucket) GetBranches() ([]string, error) {
	if b.BucketName == "" {
		return nil, errors.New("BucketName not set")
	}

	path := fmt.Sprintf("/%s/?delimiter=/", b.BucketName)
	resp, err := b.makeRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GCS API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result ListBucketResult
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing XML response: %w", err)
	}

	var branches []string
	for _, prefix := range result.CommonPrefixes {
		if prefix.Prefix != "" {
			// Remove trailing slash
			branch := strings.TrimSuffix(prefix.Prefix, "/")
			branches = append(branches, branch)
		}
	}

	return branches, nil
}

func (b *GCSBucket) GetRuntimeVersions(branch string) ([]RuntimeVersionWithStats, error) {
	if b.BucketName == "" {
		return nil, errors.New("BucketName not set")
	}

	path := fmt.Sprintf("/%s/?prefix=%s/&delimiter=/", b.BucketName, url.QueryEscape(branch))
	resp, err := b.makeRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GCS API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result ListBucketResult
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing XML response: %w", err)
	}

	var runtimeVersions []RuntimeVersionWithStats
	for _, prefix := range result.CommonPrefixes {
		if prefix.Prefix != "" {
			// Extract runtime version from prefix like "branch/runtimeVersion/"
			parts := strings.Split(strings.TrimSuffix(prefix.Prefix, "/"), "/")
			if len(parts) >= 2 {
				runtimeVersion := parts[1]

				// Get stats for this runtime version
				updatePath := fmt.Sprintf("/%s/?prefix=%s&delimiter=/", b.BucketName, url.QueryEscape(prefix.Prefix))
				updateResp, err := b.makeRequest("GET", updatePath, nil)
				if err != nil {
					continue
				}

				var updateResult ListBucketResult
				if updateResp.StatusCode == 200 {
					xml.NewDecoder(updateResp.Body).Decode(&updateResult)
				}
				updateResp.Body.Close()

				// Calculate stats
				var lastUpdatedAt time.Time
				var createdAt time.Time
				numberOfUpdates := len(updateResult.CommonPrefixes)

				if numberOfUpdates > 0 {
					// Find the most recent update
					for _, updatePrefix := range updateResult.CommonPrefixes {
						updateParts := strings.Split(strings.TrimSuffix(updatePrefix.Prefix, "/"), "/")
						if len(updateParts) >= 3 {
							if updateId, err := strconv.ParseInt(updateParts[2], 10, 64); err == nil {
								updateTime := time.Duration(updateId) * time.Millisecond
								if lastUpdatedAt.IsZero() || updateTime > time.Duration(lastUpdatedAt.UnixMilli())*time.Millisecond {
									lastUpdatedAt = time.UnixMilli(updateId)
								}
								if createdAt.IsZero() || updateTime < time.Duration(createdAt.UnixMilli())*time.Millisecond {
									createdAt = time.UnixMilli(updateId)
								}
							}
						}
					}
				}

				runtimeVersions = append(runtimeVersions, RuntimeVersionWithStats{
					RuntimeVersion:  runtimeVersion,
					LastUpdatedAt:   lastUpdatedAt.Format(time.RFC3339),
					CreatedAt:       createdAt.Format(time.RFC3339),
					NumberOfUpdates: numberOfUpdates,
				})
			}
		}
	}

	return runtimeVersions, nil
}

func (b *GCSBucket) GetUpdates(branch string, runtimeVersion string) ([]types.Update, error) {
	if b.BucketName == "" {
		return nil, errors.New("BucketName not set")
	}

	prefix := fmt.Sprintf("%s/%s/", branch, runtimeVersion)
	path := fmt.Sprintf("/%s/?prefix=%s&delimiter=/", b.BucketName, url.QueryEscape(prefix))

	resp, err := b.makeRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GCS API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result ListBucketResult
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing XML response: %w", err)
	}

	var updates []types.Update
	for _, commonPrefix := range result.CommonPrefixes {
		if commonPrefix.Prefix != "" {
			// Extract update ID from prefix like "branch/runtimeVersion/updateId/"
			parts := strings.Split(strings.TrimSuffix(commonPrefix.Prefix, "/"), "/")
			if len(parts) >= 3 {
				updateIdStr := parts[2]
				if updateId, err := strconv.ParseInt(updateIdStr, 10, 64); err == nil {
					updates = append(updates, types.Update{
						Branch:         branch,
						RuntimeVersion: runtimeVersion,
						UpdateId:       updateIdStr,
						CreatedAt:      time.Duration(updateId) * time.Millisecond,
					})
				}
			}
		}
	}

	return updates, nil
}

func (b *GCSBucket) GetFile(update types.Update, assetPath string) (*types.BucketFile, error) {
	if b.BucketName == "" {
		return nil, errors.New("BucketName not set")
	}

	key := fmt.Sprintf("%s/%s/%s/%s", update.Branch, update.RuntimeVersion, update.UpdateId, assetPath)
	path := fmt.Sprintf("/%s/%s", b.BucketName, key)

	resp, err := b.makeRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	if resp.StatusCode == 404 {
		resp.Body.Close()
		return nil, nil
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("GCS API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse Last-Modified header
	var lastModified time.Time
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		if parsed, err := time.Parse(time.RFC1123, lm); err == nil {
			lastModified = parsed
		}
	}

	return &types.BucketFile{
		Reader:    resp.Body,
		CreatedAt: lastModified,
	}, nil
}

func (b *GCSBucket) RequestUploadUrlForFileUpdate(branch string, runtimeVersion string, updateId string, fileName string) (string, error) {
	if b.BucketName == "" {
		return "", errors.New("BucketName not set")
	}
	if b.AccessKey == "" || b.SecretKey == "" {
		return "", errors.New("access key and secret key must be set")
	}

	// Generate signed URL for PUT operation
	key := fmt.Sprintf("%s/%s/%s/%s", branch, runtimeVersion, updateId, fileName)
	resource := fmt.Sprintf("/%s/%s", b.BucketName, key)
	
	// Set expiration time (1 hour from now)
	expiration := time.Now().UTC().Add(1 * time.Hour)
	expirationUnix := expiration.Unix()
	
	// Determine content type based on file extension
	contentType := ""
	if strings.HasSuffix(strings.ToLower(fileName), ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(strings.ToLower(fileName), ".jpg") || strings.HasSuffix(strings.ToLower(fileName), ".jpeg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(strings.ToLower(fileName), ".gif") {
		contentType = "image/gif"
	} else if strings.HasSuffix(strings.ToLower(fileName), ".svg") {
		contentType = "image/svg+xml"
	} else if strings.HasSuffix(strings.ToLower(fileName), ".webp") {
		contentType = "image/webp"
	} else if strings.HasSuffix(strings.ToLower(fileName), ".js") {
		contentType = "application/javascript"
	} else if strings.HasSuffix(strings.ToLower(fileName), ".json") {
		contentType = "application/json"
	} else {
		contentType = "application/octet-stream"
	}
	
	// Create string to sign for signed URL - GCS format
	// Format: HTTP-Verb + "\n" + Content-MD5 + "\n" + Content-Type + "\n" + Expiration + "\n" + Canonicalized_Resource
	stringToSign := fmt.Sprintf("PUT\n\n%s\n%d\n%s", contentType, expirationUnix, resource)
	
	// Calculate HMAC-SHA1 signature
	h := hmac.New(sha1.New, []byte(b.SecretKey))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	// Build signed URL
	signedURL := fmt.Sprintf("%s%s?GoogleAccessId=%s&Expires=%d&Signature=%s",
		b.BaseURL,
		resource,
		url.QueryEscape(b.AccessKey),
		expirationUnix,
		url.QueryEscape(signature),
	)
	
	return signedURL, nil
}

func (b *GCSBucket) UploadFileIntoUpdate(update types.Update, fileName string, file io.Reader) error {
	if b.BucketName == "" {
		return errors.New("BucketName not set")
	}

	key := fmt.Sprintf("%s/%s/%s/%s", update.Branch, update.RuntimeVersion, update.UpdateId, fileName)
	path := fmt.Sprintf("/%s/%s", b.BucketName, key)

	resp, err := b.makeRequest("PUT", path, file)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GCS API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (b *GCSBucket) DeleteUpdateFolder(branch string, runtimeVersion string, updateId string) error {
	if b.BucketName == "" {
		return errors.New("BucketName not set")
	}

	prefix := fmt.Sprintf("%s/%s/%s/", branch, runtimeVersion, updateId)
	path := fmt.Sprintf("/%s/?prefix=%s", b.BucketName, url.QueryEscape(prefix))

	resp, err := b.makeRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("error listing objects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GCS API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result ListBucketResult
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("error parsing XML response: %w", err)
	}

	// Delete each object
	for _, obj := range result.Contents {
		if obj.Key != "" {
			objPath := fmt.Sprintf("/%s/%s", b.BucketName, obj.Key)
			delResp, err := b.makeRequest("DELETE", objPath, nil)
			if err != nil {
				return fmt.Errorf("error deleting object %s: %w", obj.Key, err)
			}
			delResp.Body.Close()
		}
	}

	return nil
}

func (b *GCSBucket) CreateUpdateFrom(previousUpdate *types.Update, newUpdateId string) (*types.Update, error) {
	if b.BucketName == "" {
		return nil, errors.New("BucketName not set")
	}
	if previousUpdate == nil {
		return nil, errors.New("previousUpdate is nil")
	}
	if previousUpdate.UpdateId == "" {
		return nil, errors.New("previousUpdate.UpdateId is empty")
	}
	if newUpdateId == "" {
		return nil, errors.New("newUpdateId is empty")
	}

	sourcePrefix := fmt.Sprintf("%s/%s/%s/", previousUpdate.Branch, previousUpdate.RuntimeVersion, previousUpdate.UpdateId)
	targetPrefix := fmt.Sprintf("%s/%s/%s/", previousUpdate.Branch, previousUpdate.RuntimeVersion, newUpdateId)

	// List objects in the source folder
	path := fmt.Sprintf("/%s/?prefix=%s", b.BucketName, url.QueryEscape(sourcePrefix))
	resp, err := b.makeRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing objects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GCS API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result ListBucketResult
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing XML response: %w", err)
	}

	// Copy each object to the new location
	for _, obj := range result.Contents {
		if obj.Key == "" {
			continue
		}

		// Skip metadata files
		relPath := strings.TrimPrefix(obj.Key, sourcePrefix)
		if relPath == "update-metadata.json" || relPath == ".check" {
			continue
		}

		// Get the source object
		srcPath := fmt.Sprintf("/%s/%s", b.BucketName, obj.Key)
		srcResp, err := b.makeRequest("GET", srcPath, nil)
		if err != nil {
			continue // Skip this object on error
		}

		if srcResp.StatusCode == 200 {
			// Upload to new location
			newKey := targetPrefix + relPath
			dstPath := fmt.Sprintf("/%s/%s", b.BucketName, newKey)
			dstResp, err := b.makeRequest("PUT", dstPath, srcResp.Body)
			if err == nil {
				dstResp.Body.Close()
			}
		}
		srcResp.Body.Close()
	}

	updateId, err := strconv.ParseInt(newUpdateId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing update ID: %w", err)
	}
	return &types.Update{
		Branch:         previousUpdate.Branch,
		RuntimeVersion: previousUpdate.RuntimeVersion,
		UpdateId:       newUpdateId,
		CreatedAt:      time.Duration(updateId) * time.Millisecond,
	}, nil
}

func (b *GCSBucket) RetrieveMigrationHistory() ([]string, error) {
	if b.BucketName == "" {
		return nil, errors.New("BucketName not set")
	}

	path := fmt.Sprintf("/%s/.migrationhistory", b.BucketName)
	resp, err := b.makeRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// Migration history file doesn't exist, return empty history
		return nil, nil
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GCS API error (status %d): %s", resp.StatusCode, string(body))
	}

	var migrations []string
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			migrations = append(migrations, line)
		}
	}

	return migrations, nil
}

func (b *GCSBucket) ApplyMigration(migrationId string) error {
	if b.BucketName == "" {
		return errors.New("BucketName not set")
	}

	migrationHistory, err := b.RetrieveMigrationHistory()
	if err != nil {
		return fmt.Errorf("RetrieveMigrationHistory error: %w", err)
	}

	// Check if migration is already applied
	for _, id := range migrationHistory {
		if id == migrationId {
			return nil // Already applied
		}
	}

	// Get current content
	path := fmt.Sprintf("/%s/.migrationhistory", b.BucketName)
	resp, err := b.makeRequest("GET", path, nil)
	var currentContent []byte
	if err == nil && resp.StatusCode == 200 {
		currentContent, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
	} else if resp != nil {
		resp.Body.Close()
	}

	// Append new migration ID
	newContent := append(currentContent, []byte(migrationId+"\n")...)

	// Upload updated content
	uploadResp, err := b.makeRequest("PUT", path, bytes.NewReader(newContent))
	if err != nil {
		return fmt.Errorf("error uploading migration history: %w", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != 200 {
		body, _ := io.ReadAll(uploadResp.Body)
		return fmt.Errorf("GCS API error (status %d): %s", uploadResp.StatusCode, string(body))
	}

	return nil
}

func (b *GCSBucket) RemoveMigrationFromHistory(migrationId string) error {
	if b.BucketName == "" {
		return errors.New("BucketName not set")
	}

	migrationHistory, err := b.RetrieveMigrationHistory()
	if err != nil {
		return fmt.Errorf("RetrieveMigrationHistory error: %w", err)
	}

	// Check if migration exists
	hasMigration := false
	for _, id := range migrationHistory {
		if id == migrationId {
			hasMigration = true
			break
		}
	}
	if !hasMigration {
		return nil // Migration doesn't exist, nothing to remove
	}

	// Build new content without the migration ID
	var newContent []byte
	for _, id := range migrationHistory {
		if id != migrationId {
			newContent = append(newContent, []byte(id+"\n")...)
		}
	}

	// Upload updated content
	path := fmt.Sprintf("/%s/.migrationhistory", b.BucketName)
	resp, err := b.makeRequest("PUT", path, bytes.NewReader(newContent))
	if err != nil {
		return fmt.Errorf("error uploading migration history: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GCS API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
