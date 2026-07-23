package storage

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
)

// storageObjects captures provider operations for tests.
type storageObjects struct {
	// exists reports whether the bucket exists.
	exists bool
	// made reports whether the bucket was created.
	made bool
	// policy stores the applied bucket policy.
	policy string
	// key stores the last object key.
	key string
	// body stores the last uploaded bytes.
	body []byte
	// deadline reports whether a request timeout was installed.
	deadline bool
}

// BucketExists returns the configured bucket state.
func (objects *storageObjects) BucketExists(context.Context, string) (bool, error) {
	return objects.exists, nil
}

// MakeBucket records bucket creation.
func (objects *storageObjects) MakeBucket(context.Context, string, minio.MakeBucketOptions) error {
	objects.made = true
	return nil
}

// SetBucketPolicy records public policy application.
func (objects *storageObjects) SetBucketPolicy(_ context.Context, _ string, policy string) error {
	objects.policy = policy
	return nil
}

// PutObject records one upload.
func (objects *storageObjects) PutObject(ctx context.Context, _ string, key string, body io.Reader, _ int64, _ minio.PutObjectOptions) (minio.UploadInfo, error) {
	_, objects.deadline = ctx.Deadline()
	objects.key = key
	objects.body, _ = io.ReadAll(body)
	return minio.UploadInfo{}, nil
}

// RemoveObject records one deletion.
func (objects *storageObjects) RemoveObject(ctx context.Context, _ string, key string, _ minio.RemoveObjectOptions) error {
	_, objects.deadline = ctx.Deadline()
	objects.key = key
	return nil
}

// TestClientLifecycleAndObjects verifies bucket, URL, upload, and delete behavior.
func TestClientLifecycleAndObjects(t *testing.T) {
	objects := &storageObjects{}
	client := &Client{objects: objects, config: Config{Endpoint: "storage.local", PublicBaseURL: "https://cdn.local/camera", Bucket: "camera", UseSSL: true, PublicRead: true, UploadTimeout: time.Second}}
	if err := client.ensureBucket(context.Background()); err != nil || !objects.made || !strings.Contains(objects.policy, "arn:aws:s3:::camera/*") {
		t.Fatalf("made=%v policy=%q err=%v", objects.made, objects.policy, err)
	}
	url, err := client.Put(context.Background(), "photos/1/a.png", bytes.NewReader([]byte("png")), 3, "image/png")
	if err != nil || url != "https://cdn.local/camera/photos/1/a.png" || !objects.deadline || !bytes.Equal(objects.body, []byte("png")) {
		t.Fatalf("url=%q body=%q deadline=%v err=%v", url, objects.body, objects.deadline, err)
	}
	if err = client.Delete(context.Background(), "photos/1/a.png"); err != nil || objects.key != "photos/1/a.png" {
		t.Fatalf("key=%q err=%v", objects.key, err)
	}
}

// TestClientRejectsInvalidKeys verifies traversal never reaches a provider.
func TestClientRejectsInvalidKeys(t *testing.T) {
	client := &Client{objects: &storageObjects{}, config: Config{Bucket: "camera", UploadTimeout: time.Second}}
	if _, err := client.Put(context.Background(), "../secret", bytes.NewReader([]byte{1}), 1, "image/png"); err == nil {
		t.Fatal("expected traversal rejection")
	}
}
