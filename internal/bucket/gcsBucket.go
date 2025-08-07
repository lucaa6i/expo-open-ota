package bucket

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/xml"
	"errors"
	"expo-open-ota/config"
	"expo-open-ota/internal/types"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
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
	XMLName     xml.Name `xml:"ListBucketResult"`
	Contents    []Object `xml:"Contents"`
	CommonPrefixes []CommonPrefix `xml:"CommonPrefixes"`
	IsTruncated bool     `xml:"IsTruncated"`
	NextMarker  string   `xml:"NextMarker"`
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

// generateSignature creates an AWS v4 signature for GCS compatibility
func (b *GCSBucket) generateSignature(method, path string, headers map[string]string, payload string) (map[string]string, error) {
	if b.AccessKey == "" || b.SecretKey == "" {
		return nil, errors.New("access key and secret key must be set")
	}

	now := time.Now().UTC()
	date := now.Format("20060102")
	datetime := now.Format("20060102T150405Z")
	
	// Create canonical request
	canonicalHeaders := ""
	signedHeaders := ""
	headerKeys := make([]string, 0, len(headers))
	
	// Add required headers
	headers["host"] = "storage.googleapis.com"
	headers["x-amz-date"] = datetime
	headers["x-amz-content-sha256"] = b.sha256Hash(payload)
	
	for k := range headers {
		headerKeys = append(headerKeys, strings.ToLower(k))
	}
	sort.Strings(headerKeys)
	
	for _, k := range headerKeys {
		canonicalHeaders += k + ":" + strings.TrimSpace(headers[k]) + "\n"
		if signedHeaders != "" {
			signedHeaders += ";"
		}
		signedHeaders += k
	}
	
	canonicalRequest := method + "\n" + path + "\n" + "\n" + canonicalHeaders + "\n" + signedHeaders + "\n" + b.sha256Hash(payload)
	
	// Create string to sign
	credentialScope := date + "/auto/s3/aws4_request"
	stringToSign := "AWS4-HMAC-SHA256\n" + datetime + "\n" + credentialScope + "\n" + b.sha256Hash(canonicalRequest)
	
	// Calculate signature
	dateKey := b.hmacSHA256([]byte("AWS4"+b.SecretKey), date)
	dateRegionKey := b.hmacSHA256(dateKey, "auto")
	dateRegionServiceKey := b.hmacSHA256(dateRegionKey, "s3")
	signingKey := b.hmacSHA256(dateRegionServiceKey, "aws4_request")
	signature := b.hmacSHA256(signingKey, stringToSign)
	
	// Create authorization header
	authorization := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%x",
		b.AccessKey, credentialScope, signedHeaders, signature)
	
	result := make(map[string]string)
	for k, v := range headers {
		result[k] = v
	}
	result["Authorization"] = authorization
	
	return result, nil
}

func (b *GCSBucket) sha256Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func (b *GCSBucket) hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
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
	
	headers := make(map[string]string)
	headers["Content-Type"] = "application/octet-stream"
	
	signedHeaders, err := b.generateSignature(method, path, headers, string(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error generating signature: %w", err)
	}
	
	fullURL := b.BaseURL + path
	req, err := http.NewRequest(method, fullURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	for k, v := range signedHeaders {
		req.Header.Set(k, v)
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

// Implement other required methods with stubs for now
func (b *GCSBucket) RequestUploadUrlForFileUpdate(branch string, runtimeVersion string, updateId string, fileName string) (string, error) {
	return "", errors.New("not implemented yet for GCS")
}

func (b *GCSBucket) UploadFileIntoUpdate(update types.Update, fileName string, file io.Reader) error {
	return errors.New("not implemented yet for GCS")
}

func (b *GCSBucket) DeleteUpdateFolder(branch string, runtimeVersion string, updateId string) error {
	return errors.New("not implemented yet for GCS")
}

func (b *GCSBucket) CreateUpdateFrom(previousUpdate *types.Update, newUpdateId string) (*types.Update, error) {
	return nil, errors.New("not implemented yet for GCS")
}

func (b *GCSBucket) RetrieveMigrationHistory() ([]string, error) {
	return nil, errors.New("not implemented yet for GCS")
}

func (b *GCSBucket) ApplyMigration(migrationId string) error {
	return errors.New("not implemented yet for GCS")
}

func (b *GCSBucket) RemoveMigrationFromHistory(migrationId string) error {
	return errors.New("not implemented yet for GCS")
}
