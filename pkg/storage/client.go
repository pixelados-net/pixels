package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ErrInvalidConfig reports unusable storage configuration.
var ErrInvalidConfig = errors.New("invalid object storage config")

// objectClient contains the MinIO operations used by Client.
type objectClient interface {
	// BucketExists reports whether a bucket exists.
	BucketExists(context.Context, string) (bool, error)
	// MakeBucket creates a bucket.
	MakeBucket(context.Context, string, minio.MakeBucketOptions) error
	// SetBucketPolicy replaces a bucket policy.
	SetBucketPolicy(context.Context, string, string) error
	// PutObject uploads one object.
	PutObject(context.Context, string, string, io.Reader, int64, minio.PutObjectOptions) (minio.UploadInfo, error)
	// RemoveObject deletes one object.
	RemoveObject(context.Context, string, string, minio.RemoveObjectOptions) error
}

// Client wraps durable S3-compatible object operations.
type Client struct {
	// objects performs provider-specific operations.
	objects objectClient
	// config stores immutable storage policy.
	config Config
	// log records durable object mutations.
	log *zap.Logger
}

// New creates an object client and installs startup validation.
func New(lifecycle fx.Lifecycle, config Config, logger *zap.Logger) (*Client, error) {
	if !config.valid() {
		return nil, ErrInvalidConfig
	}
	objects, err := minio.New(config.Endpoint, &minio.Options{Creds: credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""), Secure: config.UseSSL})
	if err != nil {
		return nil, fmt.Errorf("create object storage client: %w", err)
	}
	client := &Client{objects: objects, config: config, log: logger}
	lifecycle.Append(fx.Hook{OnStart: func(ctx context.Context) error {
		if err := client.ensureBucket(ctx); err != nil {
			return err
		}
		if logger != nil {
			logger.Info("object storage connected", zap.String("endpoint", config.Endpoint), zap.String("bucket", config.Bucket), zap.Bool("public_read", config.PublicRead))
		}
		return nil
	}})
	return client, nil
}

// Put uploads one bounded object and returns its durable public URL.
func (client *Client) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) (string, error) {
	if client == nil || client.objects == nil || !validKey(key) || body == nil || size <= 0 || strings.TrimSpace(contentType) == "" {
		return "", ErrInvalidConfig
	}
	requestCtx, cancel := context.WithTimeout(ctx, client.config.UploadTimeout)
	defer cancel()
	startedAt := time.Now()
	_, err := client.objects.PutObject(requestCtx, client.config.Bucket, key, body, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("put object %q: %w", key, err)
	}
	if client.log != nil {
		client.log.Debug("object storage put", zap.String("bucket", client.config.Bucket), zap.String("key", key), zap.Int64("bytes", size), zap.Duration("duration", time.Since(startedAt)))
	}
	return client.PublicURL(key), nil
}

// Delete removes one object with the configured timeout.
func (client *Client) Delete(ctx context.Context, key string) error {
	if client == nil || client.objects == nil || !validKey(key) {
		return ErrInvalidConfig
	}
	requestCtx, cancel := context.WithTimeout(ctx, client.config.UploadTimeout)
	defer cancel()
	if err := client.objects.RemoveObject(requestCtx, client.config.Bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete object %q: %w", key, err)
	}
	if client.log != nil {
		client.log.Info("object storage delete", zap.String("bucket", client.config.Bucket), zap.String("key", key))
	}
	return nil
}

// PublicURL resolves one permanent public object URL without signing it.
func (client *Client) PublicURL(key string) string {
	if client == nil || !validKey(key) {
		return ""
	}
	base := strings.TrimRight(client.config.PublicBaseURL, "/")
	if base == "" {
		scheme := "http"
		if client.config.UseSSL {
			scheme = "https"
		}
		base = scheme + "://" + client.config.Endpoint + "/" + client.config.Bucket
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return ""
	}
	parsed.Path = path.Join(parsed.Path, key)
	return parsed.String()
}

// ensureBucket verifies connectivity, creates the bucket, and applies public reads.
func (client *Client) ensureBucket(ctx context.Context) error {
	requestCtx, cancel := context.WithTimeout(ctx, client.config.UploadTimeout)
	defer cancel()
	exists, err := client.objects.BucketExists(requestCtx, client.config.Bucket)
	if err != nil {
		return fmt.Errorf("check object storage bucket: %w", err)
	}
	if !exists {
		if err = client.objects.MakeBucket(requestCtx, client.config.Bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create object storage bucket: %w", err)
		}
	}
	if client.config.PublicRead {
		if err = client.objects.SetBucketPolicy(requestCtx, client.config.Bucket, publicReadPolicy(client.config.Bucket)); err != nil {
			return fmt.Errorf("set object storage bucket policy: %w", err)
		}
	}
	return nil
}

// validKey rejects absolute and parent-traversing object names.
func validKey(key string) bool {
	return key != "" && !strings.HasPrefix(key, "/") && path.Clean(key) == key && key != "." && !strings.HasPrefix(key, "../")
}

// publicReadPolicy returns an idempotent S3 read-only bucket policy.
func publicReadPolicy(bucket string) string {
	return fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, bucket)
}
