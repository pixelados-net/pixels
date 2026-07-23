package figure

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	// defaultFigureDataTimeout bounds direct Config values without a timeout.
	defaultFigureDataTimeout = 15 * time.Second
	// defaultFigureDataMaxBytes bounds direct Config values without a size limit.
	defaultFigureDataMaxBytes int64 = 16 * 1024 * 1024
)

// loadFigureData reads the configured local override or remote figure-data source.
func loadFigureData(config Config) ([]byte, string, error) {
	config = normalizedConfig(config)
	path := strings.TrimSpace(config.Path)
	if path != "" {
		return loadFigureDataFile(path, config.MaxBytes)
	}
	sourceURL := strings.TrimSpace(config.URL)
	if sourceURL == "" {
		return nil, "", fmt.Errorf("PIXELS_FIGURE_DATA_URL or PIXELS_FIGURE_DATA_PATH is required")
	}
	return loadFigureDataURL(sourceURL, config.Timeout, config.MaxBytes)
}

// normalizedConfig supplies safe limits to direct callers that omit optional fields.
func normalizedConfig(config Config) Config {
	if config.Timeout <= 0 {
		config.Timeout = defaultFigureDataTimeout
	}
	if config.MaxBytes <= 0 {
		config.MaxBytes = defaultFigureDataMaxBytes
	}
	return config
}

// loadFigureDataFile reads one bounded local figure-data document.
func loadFigureDataFile(path string, maxBytes int64) ([]byte, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("open local source: %w", err)
	}
	defer file.Close()
	data, err := readFigureData(file, maxBytes)
	if err != nil {
		return nil, "", fmt.Errorf("read local source: %w", err)
	}
	return data, path, nil
}

// loadFigureDataURL downloads one bounded remote figure-data document.
func loadFigureDataURL(sourceURL string, timeout time.Duration, maxBytes int64) ([]byte, string, error) {
	parsed, err := url.ParseRequestURI(sourceURL)
	if err != nil {
		return nil, "", fmt.Errorf("parse remote source: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, "", fmt.Errorf("unsupported figure data URL scheme %q", parsed.Scheme)
	}
	request, err := http.NewRequest(http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, "", fmt.Errorf("create remote request: %w", err)
	}
	response, err := (&http.Client{Timeout: timeout}).Do(request)
	if err != nil {
		return nil, "", fmt.Errorf("request remote source: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, "", fmt.Errorf("request remote source: unexpected status %s", response.Status)
	}
	data, err := readFigureData(response.Body, maxBytes)
	if err != nil {
		return nil, "", fmt.Errorf("read remote source: %w", err)
	}
	return data, response.Request.URL.Path, nil
}

// readFigureData reads at most maxBytes and rejects a larger document.
func readFigureData(reader io.Reader, maxBytes int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maxBytes))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) < maxBytes {
		return data, nil
	}
	var extra [1]byte
	count, err := reader.Read(extra[:])
	if err != nil && err != io.EOF {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("figure data exceeds %d bytes", maxBytes)
	}
	return data, nil
}
