package bucket

import (
	"context"
	"errors"
	"expo-open-ota/internal/services"
	"expo-open-ota/internal/types"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
)

type S3Bucket struct {
	BucketName string
}

func (b *S3Bucket) DeleteUpdateFolder(branch, runtimeVersion, updateId string) error {
	if b.BucketName == "" {
		return errors.New("BucketName not set")
	}

	s3Client, err := services.GetS3Client()
	if err != nil {
		return fmt.Errorf("error getting S3 client: %w", err)
	}

	prefix := fmt.Sprintf("%s/%s/%s/", branch, runtimeVersion, updateId)

	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(b.BucketName),
		Prefix: aws.String(prefix),
	}

	var objects []s3types.ObjectIdentifier

	paginator := s3.NewListObjectsV2Paginator(s3Client, listInput)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			objects = append(objects, s3types.ObjectIdentifier{
				Key: obj.Key,
			})
		}
	}

	if len(objects) == 0 {
		return nil
	}

	const batchSize = 1000
	for i := 0; i < len(objects); i += batchSize {
		end := i + batchSize
		if end > len(objects) {
			end = len(objects)
		}

		deleteInput := &s3.DeleteObjectsInput{
			Bucket: aws.String(b.BucketName),
			Delete: &s3types.Delete{
				Objects: objects[i:end],
				Quiet:   aws.Bool(true),
			},
		}

		_, err := s3Client.DeleteObjects(context.TODO(), deleteInput)
		if err != nil {
			return fmt.Errorf("failed to delete objects: %w", err)
		}
	}

	return nil
}

func (b *S3Bucket) GetRuntimeVersions(branch string) ([]RuntimeVersionWithStats, error) {
	if b.BucketName == "" {
		return nil, errors.New("BucketName not set")
	}
	s3Client, errS3 := services.GetS3Client()
	if errS3 != nil {
		return nil, errS3
	}

	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(b.BucketName),
		Prefix:    aws.String(branch + "/"),
		Delimiter: aws.String("/"),
	}
	resp, err := s3Client.ListObjectsV2(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("ListObjectsV2 error: %w", err)
	}

	var runtimeVersions []RuntimeVersionWithStats
	prefixLen := len(branch) + 1

	for _, commonPrefix := range resp.CommonPrefixes {
		runtimeVersion := (*commonPrefix.Prefix)[prefixLen : len(*commonPrefix.Prefix)-1]
		updatesPath := *commonPrefix.Prefix
		updateInput := &s3.ListObjectsV2Input{
			Bucket:    aws.String(b.BucketName),
			Prefix:    aws.String(updatesPath),
			Delimiter: aws.String("/"),
		}
		updateResp, err := s3Client.ListObjectsV2(context.TODO(), updateInput)
		if err != nil {
			return nil, fmt.Errorf("ListObjectsV2 error in updates: %w", err)
		}

		var updateTimestamps []int64
		for _, commonPrefix := range updateResp.CommonPrefixes {
			updateID := strings.TrimSuffix((*commonPrefix.Prefix)[len(updatesPath):], "/")
			timestamp, err := strconv.ParseInt(updateID, 10, 64)
			if err != nil {
				continue
			}
			updateTimestamps = append(updateTimestamps, timestamp)
		}

		if len(updateTimestamps) == 0 {
			continue
		}

		sort.Slice(updateTimestamps, func(i, j int) bool { return updateTimestamps[i] < updateTimestamps[j] })

		runtimeVersions = append(runtimeVersions, RuntimeVersionWithStats{
			RuntimeVersion:  runtimeVersion,
			CreatedAt:       time.UnixMilli(updateTimestamps[0]).UTC().Format(time.RFC3339),
			LastUpdatedAt:   time.UnixMilli(updateTimestamps[len(updateTimestamps)-1]).UTC().Format(time.RFC3339),
			NumberOfUpdates: len(updateTimestamps),
		})
	}

	return runtimeVersions, nil
}

func (b *S3Bucket) GetBranches() ([]string, error) {
	if b.BucketName == "" {
		return nil, errors.New("BucketName not set")
	}
	s3Client, errS3 := services.GetS3Client()
	if errS3 != nil {
		return nil, errS3
	}
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(b.BucketName),
		Delimiter: aws.String("/"),
	}
	resp, err := s3Client.ListObjectsV2(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("ListObjectsV2 error: %w", err)
	}
	var branches []string
	for _, commonPrefix := range resp.CommonPrefixes {
		prefix := *commonPrefix.Prefix
		branches = append(branches, prefix[:len(prefix)-1])
	}
	return branches, nil
}

func (b *S3Bucket) GetUpdates(branch string, runtimeVersion string) ([]types.Update, error) {
	if b.BucketName == "" {
		return nil, errors.New("BucketName not set")
	}
	s3Client, errS3 := services.GetS3Client()
	if errS3 != nil {
		return nil, errS3
	}
	prefix := branch + "/" + runtimeVersion + "/"
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(b.BucketName),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	}
	resp, err := s3Client.ListObjectsV2(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("ListObjectsV2 error: %w", err)
	}
	var updates []types.Update
	for _, commonPrefix := range resp.CommonPrefixes {
		var updateId int64
		if _, err := fmt.Sscanf(*commonPrefix.Prefix, prefix+"%d/", &updateId); err == nil {
			updates = append(updates, types.Update{
				Branch:         branch,
				RuntimeVersion: runtimeVersion,
				UpdateId:       strconv.FormatInt(updateId, 10),
				CreatedAt:      time.Duration(updateId) * time.Millisecond,
			})
		}
	}
	return updates, nil
}

func (b *S3Bucket) GetFile(update types.Update, assetPath string) (types.BucketFile, error) {
	if b.BucketName == "" {
		return types.BucketFile{}, errors.New("BucketName not set")
	}
	filePath := update.Branch + "/" + update.RuntimeVersion + "/" + update.UpdateId + "/" + assetPath
	s3Client, errS3 := services.GetS3Client()
	if errS3 != nil {
		return types.BucketFile{}, errS3
	}
	input := &s3.GetObjectInput{
		Bucket: aws.String(b.BucketName),
		Key:    aws.String(filePath),
	}
	resp, err := s3Client.GetObject(context.TODO(), input)
	if err != nil {

		return types.BucketFile{}, fmt.Errorf("GetObject error: %w", err)
	}
	return types.BucketFile{
		Reader:    resp.Body,
		CreatedAt: *resp.LastModified,
	}, nil
}

func (b *S3Bucket) RequestUploadUrlForFileUpdate(branch string, runtimeVersion string, updateId string, fileName string) (string, error) {
	if b.BucketName == "" {
		return "", errors.New("BucketName not set")
	}

	s3Client, err := services.GetS3Client()
	if err != nil {
		return "", fmt.Errorf("error getting S3 client: %w", err)
	}

	presignClient := s3.NewPresignClient(s3Client)

	key := fmt.Sprintf("%s/%s/%s/%s", branch, runtimeVersion, updateId, fileName)

	input := &s3.PutObjectInput{
		Bucket: aws.String(b.BucketName),
		Key:    aws.String(key),
	}

	presignResult, err := presignClient.PresignPutObject(context.TODO(), input, func(opt *s3.PresignOptions) {
		opt.Expires = 15 * time.Minute
	})
	if err != nil {
		return "", fmt.Errorf("error presigning URL: %w", err)
	}

	return presignResult.URL, nil
}

func (b *S3Bucket) UploadFileIntoUpdate(update types.Update, fileName string, file io.Reader) error {
	if b.BucketName == "" {
		return errors.New("BucketName not set")
	}
	s3Client, err := services.GetS3Client()
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s/%s/%s/%s", update.Branch, update.RuntimeVersion, update.UpdateId, fileName)
	input := &s3.PutObjectInput{
		Bucket: aws.String(b.BucketName),
		Key:    aws.String(key),
		Body:   file,
	}
	_, err = s3Client.PutObject(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("PutObject error: %w", err)
	}
	return nil
}
