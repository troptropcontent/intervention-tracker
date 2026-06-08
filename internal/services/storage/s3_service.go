package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/troptropcontent/qr_code_maintenance/internal/utils"
)

const BaseDefaultBucketName = "intervention-tracker"

// S3StorageService implements StorageService using S3-compatible storage
type S3StorageService struct {
	client     *s3.Client
	bucketName string
	baseURL    string
}

// S3Config holds configuration for S3-compatible storage
type S3Config struct {
	// Endpoint is the S3-compatible endpoint URL (e.g., "https://s3.eu-west-1.io.cloud.ovh.net")
	Endpoint string
	// Region is the AWS region or equivalent for your provider
	Region string
	// AccessKeyID is your S3 access key
	AccessKeyID string
	// SecretAccessKey is your S3 secret key
	SecretAccessKey string
	// BucketName is the name of the S3 bucket to use
	BucketName string
	// UsePathStyle determines if path-style URLs should be used (some providers require this)
	UsePathStyle bool
}

// NewS3StorageService creates a new S3-compatible storage service.
//
// Configuration is loaded from environment variables with fallback to cfg values:
//   - S3_ENDPOINT (required): S3-compatible endpoint URL
//   - S3_ACCESS_KEY (required): S3 access key ID
//   - S3_SECRET_KEY (required): S3 secret access key
//   - S3_REGION (required): AWS region or equivalent
//   - S3_USE_PATH_STYLE (optional): Use path-style URLs (some providers require this)
//   - S3_BUCKET_NAME (optional): Bucket name, defaults to "intervention-tracker-{environment}"
//
// The cfg parameter can be nil; all configuration will be read from environment variables.
// If cfg is provided, environment variables take precedence over cfg values.
func NewS3StorageService(cfg *S3Config) (*S3StorageService, error) {
	if cfg == nil {
		cfg = &S3Config{}
	}

	endpoint := utils.GetEnv("S3_ENDPOINT", cfg.Endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("S3 endpoint is required, either set it in S3_ENDPOINT or through the S3Config argument")
	}

	accessKey := utils.GetEnv("S3_ACCESS_KEY", cfg.AccessKeyID)
	if accessKey == "" {
		return nil, fmt.Errorf("S3 access key is required, either set it in S3_ACCESS_KEY or through the S3Config argument")
	}

	secretKey := utils.GetEnv("S3_SECRET_KEY", cfg.SecretAccessKey)
	if secretKey == "" {
		return nil, fmt.Errorf("S3 secret key is required, either set it in S3_SECRET_KEY or through the S3Config argument")
	}

	region := utils.GetEnv("S3_REGION", cfg.Region)
	if region == "" {
		return nil, fmt.Errorf("S3 region is required, either set it in S3_REGION or through the S3Config argument")
	}

	usePathStyle := utils.GetEnvBool("S3_USE_PATH_STYLE", cfg.UsePathStyle)

	bucketName := utils.GetEnv("S3_BUCKET_NAME", cfg.BucketName)
	if bucketName == "" {
		bucketName = findDefaultBucketName()
	}

	// Create custom credentials provider
	credProvider := credentials.NewStaticCredentialsProvider(
		accessKey,
		secretKey,
		"", // session token (optional)
	)

	// Create S3 client with custom endpoint
	client := s3.New(s3.Options{
		Region:       region,
		Credentials:  credProvider,
		BaseEndpoint: aws.String(endpoint),
		UsePathStyle: usePathStyle,
	})

	return &S3StorageService{
		client:     client,
		bucketName: bucketName,
		baseURL:    endpoint,
	}, nil
}

// findDefaultBucketName generates a default bucket name based on the application environment.
// Returns a bucket name in the format: "intervention-tracker-{environment}"
// For example: "intervention-tracker-production" or "intervention-tracker-development"
func findDefaultBucketName() (name string) {
	return fmt.Sprintf("%v-%v", BaseDefaultBucketName, utils.GetAppEnv())
}

// UploadFile uploads a file to S3-compatible storage
func (s *S3StorageService) UploadFile(ctx context.Context, key string, file io.Reader, contentType string) (string, error) {
	// Read all content into memory (for small files)
	// For large files, you might want to use multipart upload
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Return the object URL
	url := fmt.Sprintf("%s/%s/%s", s.baseURL, s.bucketName, key)
	return url, nil
}

// DownloadFile retrieves a file from S3-compatible storage
func (s *S3StorageService) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}

	return result.Body, nil
}

// DeleteFile removes a file from S3-compatible storage
func (s *S3StorageService) DeleteFile(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}

// GetFileURL generates a presigned URL for direct file access
func (s *S3StorageService) GetFileURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	presignResult, err := presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignResult.URL, nil
}

// FileExists checks if a file exists in S3-compatible storage
func (s *S3StorageService) FileExists(ctx context.Context, key string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	_, err := s.client.HeadObject(ctx, input)
	if err != nil {
		// Check if the error is a "not found" error
		var apiErr smithy.APIError
		if ok := errors.As(err, &apiErr); ok {
			if apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "NoSuchKey" {
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}
