package storage

import (
	"testing"
	"time"
)

// TestLoadConfigDefaults verifies complete storage defaults.
func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("STORAGE_ENDPOINT", "")
	config, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.Endpoint != "127.0.0.1:9000" || config.Bucket != "pixels-camera" || config.UploadTimeout != 10*time.Second || !config.UseSSL || !config.PublicRead {
		t.Fatalf("config=%+v", config)
	}
}
