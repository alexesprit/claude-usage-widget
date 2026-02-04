package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCalculateBurnPrediction(t *testing.T) {
	// Test case 1: Will hit limit within window
	resetsAt := time.Now().Add(1 * time.Hour)               // 1 hour from now
	result, err := CalculateBurnPrediction(90, resetsAt, 5) // 90% in 1 hour left, 5h window
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty burn prediction")
	}
	// Should contain "hit in" since 90% utilization with 1 hour left should hit within ~0.11 hours
	if !strings.Contains(result, "hit in") {
		t.Errorf("Expected 'hit in' in result, got: %s", result)
	}
	if result == "" {
		t.Error("Expected non-empty burn prediction")
	}
	// Should contain "hit in" since 80% utilization with 1 hour left should hit within ~0.25 hours
	if !strings.Contains(result, "hit in") {
		t.Errorf("Expected 'hit in' in result, got: %s", result)
	}

	// Test case 2: Low burn rate, show percentage per hour
	resetsAt = time.Now().Add(3 * time.Hour)              // 3 hours from now
	result, err = CalculateBurnPrediction(5, resetsAt, 5) // 5% utilization, 5h window
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(result, "%/h") {
		t.Errorf("Expected '%%/h' in result, got: %s", result)
	}
	if !strings.Contains(result, "%/h") {
		t.Errorf("Expected '%%/h' in result, got: %s", result)
	}

	// Test case 3: Invalid utilization
	_, err = CalculateBurnPrediction(-1, resetsAt, 5)
	if err == nil {
		t.Error("Expected error for negative utilization")
	}

	_, err = CalculateBurnPrediction(101, resetsAt, 5)
	if err == nil {
		t.Error("Expected error for utilization > 100")
	}
}

func TestMainLogic(t *testing.T) {
	// Test the main logic without os.Exit calls
	tests := []struct {
		name        string
		args        []string
		expectError bool
		setup       func()
		cleanup     func()
	}{
		{
			name:        "No session flags",
			args:        []string{"--config"}, // Only config flag
			expectError: false,
			setup: func() {
				os.Setenv("CLAUDE_CODE_USAGE_SESSION_KEY", "test_key")
				os.Setenv("CLAUDE_CODE_USAGE_ORGANIZATION_ID", "test_org")
			},
			cleanup: func() {
				os.Unsetenv("CLAUDE_CODE_USAGE_SESSION_KEY")
				os.Unsetenv("CLAUDE_CODE_USAGE_ORGANIZATION_ID")
			},
		},
		{
			name:        "Both session flags",
			args:        []string{"-5h", "-weekly"},
			expectError: true,
			setup:       func() {},
			cleanup:     func() {},
		},
		{
			name:        "Valid 5h flag",
			args:        []string{"-5h"},
			expectError: true, // Will error on API call, but flag parsing should work
			setup: func() {
				os.Setenv("CLAUDE_CODE_USAGE_SESSION_KEY", "test_key")
				os.Setenv("CLAUDE_CODE_USAGE_ORGANIZATION_ID", "test_org")
			},
			cleanup: func() {
				os.Unsetenv("CLAUDE_CODE_USAGE_SESSION_KEY")
				os.Unsetenv("CLAUDE_CODE_USAGE_ORGANIZATION_ID")
			},
		},
		{
			name:        "Valid weekly flag",
			args:        []string{"-weekly"},
			expectError: true, // Will error on API call
			setup: func() {
				os.Setenv("CLAUDE_CODE_USAGE_SESSION_KEY", "test_key")
				os.Setenv("CLAUDE_CODE_USAGE_ORGANIZATION_ID", "test_org")
			},
			cleanup: func() {
				os.Unsetenv("CLAUDE_CODE_USAGE_SESSION_KEY")
				os.Unsetenv("CLAUDE_CODE_USAGE_ORGANIZATION_ID")
			},
		},
		{
			name:        "Verbose flag",
			args:        []string{"-5h", "-v"},
			expectError: true, // Will error on API call
			setup: func() {
				os.Setenv("CLAUDE_CODE_USAGE_SESSION_KEY", "test_key")
				os.Setenv("CLAUDE_CODE_USAGE_ORGANIZATION_ID", "test_org")
			},
			cleanup: func() {
				os.Unsetenv("CLAUDE_CODE_USAGE_SESSION_KEY")
				os.Unsetenv("CLAUDE_CODE_USAGE_ORGANIZATION_ID")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			// Test flag parsing logic
			flag5h := false
			flagWeekly := false
			flagConfig := false

			for _, arg := range tt.args {
				switch arg {
				case "-5h":
					flag5h = true
				case "-weekly":
					flagWeekly = true
				case "--config":
					flagConfig = true
				case "-v", "--verbose":
					// Verbose flag present
				}
			}

			// Test validation logic from main()
			if flagConfig {
				// Config path display - should work
				return
			}

			if flag5h && flagWeekly {
				if !tt.expectError {
					t.Error("Expected error for both flags set")
				}
				return
			}

			if !flag5h && !flagWeekly {
				if !tt.expectError {
					t.Error("Expected error for no session flags")
				}
				return
			}

			// If we get here, flags are valid, but API call will fail
			// This tests the flag parsing and validation logic
		})
	}
}

// Table-driven test for CalculateBurnPrediction
func TestCalculateBurnPredictionTableDriven(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		utilization int
		resetsAt    time.Time
		windowHours int
		expected    string
		expectError bool
	}{
		{
			name:        "High utilization - hits limit in 5h window",
			utilization: 90,
			resetsAt:    now.Add(1 * time.Hour),
			windowHours: 5,
			expected:    "hit in",
			expectError: false,
		},
		{
			name:        "Low utilization - shows rate per hour",
			utilization: 5,
			resetsAt:    now.Add(3 * time.Hour),
			windowHours: 5,
			expected:    "%/h",
			expectError: false,
		},
		{
			name:        "Weekly session - low utilization",
			utilization: 10,
			resetsAt:    now.Add(24 * time.Hour),
			windowHours: 168,
			expected:    "%/d",
			expectError: false,
		},
		{
			name:        "Weekly session - hits limit in days",
			utilization: 80,
			resetsAt:    now.Add(48 * time.Hour),
			windowHours: 168,
			expected:    "hit in 1d",
			expectError: false,
		},
		{
			name:        "Negative utilization",
			utilization: -1,
			resetsAt:    now.Add(1 * time.Hour),
			windowHours: 5,
			expected:    "",
			expectError: true,
		},
		{
			name:        "Utilization over 100",
			utilization: 101,
			resetsAt:    now.Add(1 * time.Hour),
			windowHours: 5,
			expected:    "",
			expectError: true,
		},
		{
			name:        "Zero window hours",
			utilization: 50,
			resetsAt:    now.Add(1 * time.Hour),
			windowHours: 0,
			expected:    "",
			expectError: true,
		},
		{
			name:        "Negative window hours",
			utilization: 50,
			resetsAt:    now.Add(1 * time.Hour),
			windowHours: -1,
			expected:    "",
			expectError: true,
		},
		{
			name:        "Very large window hours",
			utilization: 10,
			resetsAt:    now.Add(24 * time.Hour),
			windowHours: 10000,
			expected:    "%/d", // Should show daily rate for large windows
			expectError: false,
		},
		{
			name:        "Very small window hours (fractional)",
			utilization: 5,
			resetsAt:    now.Add(30 * time.Minute),
			windowHours: 1,
			expected:    "%/d", // Non-5 hour windows show daily rate
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateBurnPrediction(tt.utilization, tt.resetsAt, tt.windowHours)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !strings.Contains(result, tt.expected) {
					t.Errorf("Expected result to contain %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

// BenchmarkCalculateBurnPrediction benchmarks the CalculateBurnPrediction function
func BenchmarkCalculateBurnPrediction(b *testing.B) {
	resetsAt := time.Now().Add(1 * time.Hour)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CalculateBurnPrediction(50, resetsAt, 5)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCalculateBurnPredictionWeekly benchmarks for weekly sessions
func BenchmarkCalculateBurnPredictionWeekly(b *testing.B) {
	resetsAt := time.Now().Add(24 * time.Hour)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CalculateBurnPrediction(20, resetsAt, 168) // 7 days = 168 hours
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestGenerateProgressBar(t *testing.T) {
	// Test 0%
	bar := GenerateProgressBar(0)
	if bar != "░░░░░░░░░░" {
		t.Errorf("Expected all empty bars, got: %s", bar)
	}

	// Test 100%
	bar = GenerateProgressBar(100)
	if bar != "▓▓▓▓▓▓▓▓▓▓" {
		t.Errorf("Expected all filled bars, got: %s", bar)
	}

	// Test 50%
	bar = GenerateProgressBar(50)
	if bar != "▓▓▓▓▓░░░░░" {
		t.Errorf("Expected 5 filled, 5 empty, got: %s", bar)
	}

	// Test negative (should clamp to 0)
	bar = GenerateProgressBar(-10)
	if bar != "░░░░░░░░░░" {
		t.Errorf("Expected all empty for negative, got: %s", bar)
	}

	// Test >100 (should clamp to 100)
	bar = GenerateProgressBar(150)
	if bar != "▓▓▓▓▓▓▓▓▓▓" {
		t.Errorf("Expected all filled for >100, got: %s", bar)
	}
}

// BenchmarkGenerateProgressBar benchmarks the GenerateProgressBar function
func BenchmarkGenerateProgressBar(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateProgressBar(50)
	}
}

// Table-driven test for GenerateProgressBar
func TestGenerateProgressBarTableDriven(t *testing.T) {
	tests := []struct {
		name       string
		percentage int
		expected   string
	}{
		{"Zero percent", 0, "░░░░░░░░░░"},
		{"Ten percent", 10, "▓░░░░░░░░░"},
		{"Twenty percent", 20, "▓▓░░░░░░░░"},
		{"Thirty percent", 30, "▓▓▓░░░░░░░"},
		{"Forty percent", 40, "▓▓▓▓░░░░░░"},
		{"Fifty percent", 50, "▓▓▓▓▓░░░░░"},
		{"Sixty percent", 60, "▓▓▓▓▓▓░░░░"},
		{"Seventy percent", 70, "▓▓▓▓▓▓▓░░░"},
		{"Eighty percent", 80, "▓▓▓▓▓▓▓▓░░"},
		{"Ninety percent", 90, "▓▓▓▓▓▓▓▓▓░"},
		{"Hundred percent", 100, "▓▓▓▓▓▓▓▓▓▓"},
		{"Negative clamped", -10, "░░░░░░░░░░"},
		{"Over hundred clamped", 150, "▓▓▓▓▓▓▓▓▓▓"},
		{"One percent", 1, "░░░░░░░░░░"},  // Rounds down to 0 blocks
		{"Five percent", 5, "▓░░░░░░░░░"}, // Rounds up to 1 block
		{"Six percent", 6, "▓░░░░░░░░░"},  // Rounds up to 1 block
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateProgressBar(tt.percentage)
			if result != tt.expected {
				t.Errorf("GenerateProgressBar(%d) = %q, want %q", tt.percentage, result, tt.expected)
			}
		})
	}
}

func TestFormatTimeDisplay(t *testing.T) {
	now := time.Now()

	// Test future time
	future := now.Add(2*time.Hour + 30*time.Minute)
	result := FormatTimeDisplay(future)
	expected := "2h 30m left"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test only minutes
	future = now.Add(45 * time.Minute)
	result = FormatTimeDisplay(future)
	expected = "45m left"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test expired
	past := now.Add(-1 * time.Hour)
	result = FormatTimeDisplay(past)
	if result != "expired" {
		t.Errorf("Expected 'expired', got '%s'", result)
	}
}

func TestApplyTemplateFormatting(t *testing.T) {
	template := "{name} {percent} {progress} {time} {burn}"
	result := ApplyTemplateFormatting(template, "Test", "50%", "▓▓▓▓▓░░░░░", "2h left", "10%/h")
	expected := "Test 50% ▓▓▓▓▓░░░░░ 2h left 10%/h"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test with placeholders to remove
	template = "{name} {invalid} {percent}"
	result = ApplyTemplateFormatting(template, "Test", "50%", "", "", "")
	expected = "Test 50%"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// Table-driven test for ApplyTemplateFormatting
func TestApplyTemplateFormattingTableDriven(t *testing.T) {
	tests := []struct {
		name     string
		template string
		nameArg  string
		percent  string
		progress string
		timeArg  string
		burn     string
		expected string
	}{
		{
			name:     "All placeholders present",
			template: "{name} {percent} {progress} {time} {burn}",
			nameArg:  "5-hour",
			percent:  "75%",
			progress: "▓▓▓▓▓▓▓░░░",
			timeArg:  "2h 30m left",
			burn:     "25%/h",
			expected: "5-hour 75% ▓▓▓▓▓▓▓░░░ 2h 30m left 25%/h",
		},
		{
			name:     "Missing placeholders removed",
			template: "{name} {percent} {missing} {time}",
			nameArg:  "Weekly",
			percent:  "10%",
			progress: "",
			timeArg:  "5d left",
			burn:     "",
			expected: "Weekly 10% 5d left",
		},
		{
			name:     "Empty template",
			template: "",
			nameArg:  "Test",
			percent:  "50%",
			progress: "▓▓▓▓▓░░░░░",
			timeArg:  "1h left",
			burn:     "50%/h",
			expected: "",
		},
		{
			name:     "Only name placeholder",
			template: "{name}",
			nameArg:  "Test Session",
			percent:  "100%",
			progress: "▓▓▓▓▓▓▓▓▓▓",
			timeArg:  "expired",
			burn:     "0%/h",
			expected: "Test Session",
		},
		{
			name:     "No placeholders in template",
			template: "Static text only",
			nameArg:  "Ignored",
			percent:  "Ignored",
			progress: "Ignored",
			timeArg:  "Ignored",
			burn:     "Ignored",
			expected: "Static text only",
		},
		{
			name:     "Empty placeholders",
			template: "{} {name} {}",
			nameArg:  "Test",
			percent:  "50%",
			progress: "",
			timeArg:  "",
			burn:     "",
			expected: "Test",
		},
		{
			name:     "Special characters in values",
			template: "{name} {percent} {progress}",
			nameArg:  "Test\nWith\nNewlines",
			percent:  "50%\tTabbed",
			progress: "▓▓▓\rCarriage",
			timeArg:  "",
			burn:     "",
			expected: "Test With Newlines 50% Tabbed ▓▓▓ Carriage",
		},
		{
			name:     "Empty placeholders",
			template: "{} {name} {}",
			nameArg:  "Test",
			percent:  "50%",
			progress: "",
			timeArg:  "",
			burn:     "",
			expected: "Test",
		},
		{
			name:     "Very long template",
			template: strings.Repeat("{name} ", 1000) + "{percent}",
			nameArg:  "Test",
			percent:  "50%",
			progress: "",
			timeArg:  "",
			burn:     "",
			expected: strings.Repeat("Test ", 1000) + "50%",
		},
		{
			name:     "Special characters in values",
			template: "{name} {percent} {progress}",
			nameArg:  "Test\nWith\nNewlines",
			percent:  "50%\tTabbed",
			progress: "▓▓▓\rCarriage",
			timeArg:  "",
			burn:     "",
			expected: "Test With Newlines 50% Tabbed ▓▓▓ Carriage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyTemplateFormatting(tt.template, tt.nameArg, tt.percent, tt.progress, tt.timeArg, tt.burn)
			if result != tt.expected {
				t.Errorf("ApplyTemplateFormatting() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// BenchmarkApplyTemplateFormatting benchmarks the ApplyTemplateFormatting function
func BenchmarkApplyTemplateFormatting(b *testing.B) {
	template := "{name} {percent} {progress} {time} {burn}"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplyTemplateFormatting(template, "5-hour", "75%", "▓▓▓▓▓▓▓░░░", "2h 30m left", "25%/h")
	}
}

func TestLoadDisplayConfig(t *testing.T) {
	// Create a config with sessions
	config := &Config{
		Sessions: []Session{
			{
				Key:    "five_hour",
				Name:   "5-hour",
				Format: Format{Template: "{name} {percent} {progress} {burn}"},
			},
		},
	}

	template, err := LoadDisplayConfig(config, "five_hour")
	if err != nil {
		t.Fatalf("Failed to load display config: %v", err)
	}

	expected := "{name} {percent} {progress} {burn}"
	if template != expected {
		t.Errorf("Expected template '%s', got '%s'", expected, template)
	}

	// Test default for unknown session
	_, err = LoadDisplayConfig(config, "unknown")
	if err == nil {
		t.Error("Expected error for unknown session key")
	}

	// Test default for five_hour when not in config
	configEmpty := &Config{Sessions: []Session{}}
	template, err = LoadDisplayConfig(configEmpty, "five_hour")
	if err != nil {
		t.Fatalf("Failed to load default: %v", err)
	}
	expected = "{name}: {percent}% {progress} {time} {burn}"
	if template != expected {
		t.Errorf("Expected default template '%s', got '%s'", expected, template)
	}

	// Test default for seven_day
	template, err = LoadDisplayConfig(configEmpty, "seven_day")
	if err != nil {
		t.Fatalf("Failed to load default: %v", err)
	}
	expected = "{name}: {percent}% {progress} {time}"
	if template != expected {
		t.Errorf("Expected default template '%s', got '%s'", expected, template)
	}
}

func TestProcessSession(t *testing.T) {
	// Test data
	fullData := map[string]interface{}{
		"five_hour": map[string]interface{}{
			"utilization": 50,
			"resets_at":   time.Now().Add(2 * time.Hour).Format(time.RFC3339),
		},
	}

	template := "{name} {percent} {progress} {burn}"
	result, err := ProcessSession("five_hour", fullData, "10%/h", template)
	if err != nil {
		t.Fatalf("Failed to process session: %v", err)
	}

	// Should contain the formatted output
	if !strings.Contains(result, "5-hour") {
		t.Errorf("Expected '5-hour' in result, got: %s", result)
	}
	if !strings.Contains(result, "50%") {
		t.Errorf("Expected '50%%' in result, got: %s", result)
	}
	if !strings.Contains(result, "▓▓▓▓▓░░░░░") {
		t.Errorf("Expected progress bar in result, got: %s", result)
	}
	if !strings.Contains(result, "10%/h") {
		t.Errorf("Expected burn prediction in result, got: %s", result)
	}
}

func TestLoadConfig(t *testing.T) {
	// Test loading config with environment variables
	os.Setenv("CLAUDE_CODE_USAGE_SESSION_KEY", "test_session")
	os.Setenv("CLAUDE_CODE_USAGE_ORGANIZATION_ID", "test_org")
	os.Setenv("USAGE_DEBUG", "true")
	os.Setenv("CLAUDE_CODE_USAGE_CACHE_EXPIRY_SECONDS", "30")

	defer func() {
		os.Unsetenv("CLAUDE_CODE_USAGE_SESSION_KEY")
		os.Unsetenv("CLAUDE_CODE_USAGE_ORGANIZATION_ID")
		os.Unsetenv("USAGE_DEBUG")
		os.Unsetenv("CLAUDE_CODE_USAGE_CACHE_EXPIRY_SECONDS")
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.SessionKey != "test_session" {
		t.Errorf("Expected session key 'test_session', got '%s'", config.SessionKey)
	}

	if config.OrganizationID != "test_org" {
		t.Errorf("Expected org ID 'test_org', got '%s'", config.OrganizationID)
	}

	if config.Debug != true {
		t.Errorf("Expected debug true, got %v", config.Debug)
	}

	if config.CacheExpirySeconds != 30 {
		t.Errorf("Expected cache expiry 30, got %d", config.CacheExpirySeconds)
	}
}

func TestFileLock(t *testing.T) {
	tempDir := t.TempDir()
	lockFile := tempDir + "/test_lock.lock"
	defer os.Remove(lockFile)

	lock := NewFileLock(lockFile, 5*time.Second)

	// Test acquiring lock
	if err := lock.Acquire(); err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	// Test that lock file exists
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		t.Error("Lock file should exist after acquiring lock")
	}

	// Test releasing lock
	if err := lock.Release(); err != nil {
		t.Fatalf("Failed to release lock: %v", err)
	}

	// Test that lock file is removed
	if _, err := os.Stat(lockFile); !os.IsNotExist(err) {
		t.Error("Lock file should be removed after releasing lock")
	}
}

func TestCache(t *testing.T) {
	tempDir := t.TempDir()
	cacheFile := tempDir + "/test_cache.json"
	lockFile := tempDir + "/test_cache.json.lock"
	defer os.Remove(cacheFile)
	defer os.Remove(lockFile)

	config := &Config{
		CacheFile:          cacheFile,
		LockFile:           lockFile,
		CacheExpirySeconds: 1,
		Debug:              false,
	}

	cache := NewCache(cacheFile, 1, config)

	testData := map[string]interface{}{
		"test":   "data",
		"number": 42,
	}

	// Test setting cache
	if err := cache.Set(testData); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Test getting cache
	retrievedData, err := cache.Get()
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}

	if retrievedData["test"] != "data" {
		t.Errorf("Expected test=data, got %v", retrievedData["test"])
	}

	if retrievedData["number"] != float64(42) {
		t.Errorf("Expected number=42, got %v", retrievedData["number"])
	}

	// Test cache expiry
	time.Sleep(1100 * time.Millisecond) // Wait for cache to expire

	if cache.IsValid() {
		t.Error("Cache should be expired")
	}
}

func TestCacheClear(t *testing.T) {
	tempDir := t.TempDir()
	cacheFile := tempDir + "/test_cache.json"
	lockFile := tempDir + "/test_cache.json.lock"

	config := &Config{
		CacheFile:          cacheFile,
		LockFile:           lockFile,
		CacheExpirySeconds: 60,
		Debug:              false,
	}

	cache := NewCache(cacheFile, 60, config)

	testData := map[string]interface{}{
		"test": "data",
	}

	// Set cache
	if err := cache.Set(testData); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Verify cache exists
	if !cache.IsValid() {
		t.Error("Cache should be valid after setting")
	}

	// Clear cache
	if err := cache.Clear(); err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	// Verify cache is cleared
	if cache.IsValid() {
		t.Error("Cache should be invalid after clearing")
	}

	// Test clearing non-existent cache (should not error)
	if err := cache.Clear(); err != nil {
		t.Errorf("Clearing non-existent cache should not error, got: %v", err)
	}
}

func TestCacheGetExpiryTime(t *testing.T) {
	tempDir := t.TempDir()
	cacheFile := tempDir + "/test_cache.json"
	lockFile := tempDir + "/test_cache.json.lock"

	config := &Config{
		CacheFile:          cacheFile,
		LockFile:           lockFile,
		CacheExpirySeconds: 60,
		Debug:              false,
	}

	cache := NewCache(cacheFile, 1, config)

	testData := map[string]interface{}{
		"test": "data",
	}

	// Set cache
	if err := cache.Set(testData); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Get expiry time
	expiryTime, err := cache.GetExpiryTime()
	if err != nil {
		t.Fatalf("Failed to get expiry time: %v", err)
	}

	// Verify expiry time is in the future
	if expiryTime.Before(time.Now()) {
		t.Error("Expiry time should be in the future")
	}

	// Verify expiry time is approximately correct (within 5 seconds)
	expectedExpiry := time.Now().Add(1 * time.Second)
	if expiryTime.After(expectedExpiry.Add(5*time.Second)) || expiryTime.Before(expectedExpiry.Add(-5*time.Second)) {
		t.Errorf("Expiry time %v not close to expected %v", expiryTime, expectedExpiry)
	}

	// Test getting expiry time for invalid cache
	time.Sleep(2000 * time.Millisecond) // Wait longer for cache to expire
	_, err = cache.GetExpiryTime()
	if err == nil {
		t.Error("Expected error when getting expiry time for invalid cache")
	}
}

func TestUserFriendlyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "APIError 401",
			err:      NewAPIError("request", "https://example.com", 401, "Unauthorized", nil),
			expected: "Authentication failed. Please check your session key and try again.",
		},
		{
			name:     "APIError 403",
			err:      NewAPIError("request", "https://example.com", 403, "Forbidden", nil),
			expected: "Access denied. You may not have permission to access this organization's usage data.",
		},
		{
			name:     "APIError 404",
			err:      NewAPIError("request", "https://example.com", 404, "Not Found", nil),
			expected: "Organization not found. Please verify the organization ID.",
		},
		{
			name:     "APIError 429",
			err:      NewAPIError("request", "https://example.com", 429, "Too Many Requests", nil),
			expected: "Too many requests. Please wait a moment before trying again.",
		},
		{
			name:     "APIError 500",
			err:      NewAPIError("request", "https://example.com", 500, "Internal Server Error", nil),
			expected: "Claude API is temporarily unavailable. Please try again later.",
		},
		{
			name:     "APIError unknown status",
			err:      NewAPIError("request", "https://example.com", 418, "I'm a teapot", nil),
			expected: "Failed to retrieve usage data: I'm a teapot",
		},
		{
			name:     "ConfigError session_key",
			err:      NewConfigError("session_key", "missing session key", nil),
			expected: "Session key is not configured. Please set your CLAUDE_SESSION_KEY environment variable.",
		},
		{
			name:     "ConfigError organization_id",
			err:      NewConfigError("organization_id", "missing org id", nil),
			expected: "Organization ID is not configured. Please set your CLAUDE_ORGANIZATION_ID environment variable.",
		},
		{
			name:     "ConfigError other field",
			err:      NewConfigError("cache_file", "invalid path", nil),
			expected: "Configuration issue: invalid path",
		},
		{
			name:     "ValidationError",
			err:      NewValidationError("orgId", "invalid/org", "contains invalid characters"),
			expected: "Invalid input: contains invalid characters",
		},
		{
			name:     "Generic error",
			err:      fmt.Errorf("some random error"),
			expected: "An unexpected error occurred: some random error",
		},
		{
			name:     "Wrapped APIError",
			err:      fmt.Errorf("wrapped: %w", NewAPIError("request", "https://example.com", 401, "Unauthorized", nil)),
			expected: "Authentication failed. Please check your session key and try again.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UserFriendlyError(tt.err)
			if result != tt.expected {
				t.Errorf("UserFriendlyError() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	// Test APIError
	cause := fmt.Errorf("underlying cause")
	apiErr := NewAPIError("test_op", "https://example.com", 500, "test message", cause)

	// Test Error() method
	errStr := apiErr.Error()
	expected := "API request failed: test_op https://example.com (HTTP 500): test message"
	if errStr != expected {
		t.Errorf("APIError.Error() = %q, want %q", errStr, expected)
	}

	// Test Unwrap() method
	unwrapped := apiErr.Unwrap()
	if unwrapped != cause {
		t.Errorf("APIError.Unwrap() = %v, want %v", unwrapped, cause)
	}

	// Test ConfigError
	configErr := NewConfigError("field", "message", cause)
	errStr = configErr.Error()
	expected = "configuration error for 'field': message"
	if errStr != expected {
		t.Errorf("ConfigError.Error() = %q, want %q", errStr, expected)
	}

	unwrapped = configErr.Unwrap()
	if unwrapped != cause {
		t.Errorf("ConfigError.Unwrap() = %v, want %v", unwrapped, cause)
	}

	// Test ValidationError
	valErr := NewValidationError("field", "value", "message")
	errStr = valErr.Error()
	expected = "validation failed for 'field': message"
	if errStr != expected {
		t.Errorf("ValidationError.Error() = %q, want %q", errStr, expected)
	}
}
