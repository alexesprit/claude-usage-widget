package main

import (
	"context"
	"strings"
	"testing"
	"time"
)

// UsageDataFetcher interface for dependency injection and testing
type UsageDataFetcher interface {
	FetchFullUsageData(ctx context.Context, sessionKey, orgId string) (map[string]interface{}, error)
	FetchUsageData(ctx context.Context, sessionKey, orgId string) (utilization int, resetsAt string, err error)
}

// APIClientWrapper wraps APIClient to implement UsageDataFetcher
type APIClientWrapper struct {
	client *APIClient
}

func NewAPIClientWrapper(config *Config) *APIClientWrapper {
	return &APIClientWrapper{
		client: NewAPIClient(config),
	}
}

func (w *APIClientWrapper) FetchFullUsageData(ctx context.Context, sessionKey, orgId string) (map[string]interface{}, error) {
	return w.client.FetchFullUsageData(ctx, sessionKey, orgId)
}

func (w *APIClientWrapper) FetchUsageData(ctx context.Context, sessionKey, orgId string) (utilization int, resetsAt string, err error) {
	return w.client.FetchUsageData(ctx, sessionKey, orgId)
}

// MockUsageDataFetcher is a mock implementation for testing
type MockUsageDataFetcher struct {
	FullData map[string]interface{}
	Usage    int
	ResetsAt string
	Error    error
}

func (m *MockUsageDataFetcher) FetchFullUsageData(ctx context.Context, sessionKey, orgId string) (map[string]interface{}, error) {
	return m.FullData, m.Error
}

func (m *MockUsageDataFetcher) FetchUsageData(ctx context.Context, sessionKey, orgId string) (utilization int, resetsAt string, err error) {
	return m.Usage, m.ResetsAt, m.Error
}

// TestHelper functions

// CreateTestConfig creates a test configuration with reasonable defaults
func CreateTestConfig(t *testing.T) *Config {
	tempDir := t.TempDir()
	return &Config{
		SessionKey:         "test_session_key",
		OrganizationID:     "test_org_id",
		CacheExpirySeconds: 60,
		Debug:              false,
		Verbose:            false,
		CacheFile:          tempDir + "/test_cache.json",
		LockFile:           tempDir + "/test_cache.json.lock",
		Sessions: []Session{
			{
				Key:    "five_hour",
				Name:   "5-hour",
				Format: Format{Template: "{name}: {percent}% {progress} {time} {burn}"},
			},
			{
				Key:    "seven_day",
				Name:   "Weekly",
				Format: Format{Template: "{name}: {percent}% {progress} {time}"},
			},
		},
	}
}

// CreateMockUsageData creates mock usage data for testing
func CreateMockUsageData(sessionKey string, utilization int, hoursFromNow float64) map[string]interface{} {
	return map[string]interface{}{
		sessionKey: map[string]interface{}{
			"utilization": utilization,
			"resets_at":   time.Now().Add(time.Duration(hoursFromNow * float64(time.Hour))).Format(time.RFC3339),
		},
	}
}

// AssertContains is a test helper that checks if a string contains a substring
func AssertContains(t *testing.T, haystack, needle, message string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("%s: expected %q to contain %q", message, haystack, needle)
	}
}

// AssertEqual is a test helper for string equality
func AssertEqual(t *testing.T, expected, actual, message string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %q, got %q", message, expected, actual)
	}
}

// AssertNoError is a test helper that fails if there's an error
func AssertNoError(t *testing.T, err error, message string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", message, err)
	}
}

// AssertError is a test helper that fails if there's no error
func AssertError(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error but got none", message)
	}
}

// BenchmarkHelper is a helper for consistent benchmarking
type BenchmarkHelper struct {
	name string
	fn   func()
}

func NewBenchmarkHelper(name string, fn func()) *BenchmarkHelper {
	return &BenchmarkHelper{name: name, fn: fn}
}

func (h *BenchmarkHelper) Run(b *testing.B) {
	b.Run(h.name, func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			h.fn()
		}
	})
}
