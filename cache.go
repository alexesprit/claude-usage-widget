package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CacheData represents the structure of cached data
type CacheData struct {
	FullData  string `json:"full_data"`
	ExpiresAt int64  `json:"expires_at"`
}

// Cache represents a file-based cache system
type Cache struct {
	cacheFile      string
	expiryDuration time.Duration
	config         *Config
}

// NewCache creates a new cache instance
func NewCache(cacheFile string, expirySeconds int, config *Config) *Cache {
	return &Cache{
		cacheFile:      cacheFile,
		expiryDuration: time.Duration(expirySeconds) * time.Second,
		config:         config,
	}
}

// IsValid checks if the cache is still valid
func (c *Cache) IsValid() bool {
	// Check if cache file exists
	if _, err := os.Stat(c.cacheFile); os.IsNotExist(err) {
		c.config.LogDebug("cache file does not exist", "file", c.cacheFile)
		return false
	}

	// Read and parse cache file
	file, err := os.Open(c.cacheFile)
	if err != nil {
		c.config.LogDebug("failed to open cache file", "file", c.cacheFile, "error", err)
		return false
	}
	defer file.Close()

	var cacheData CacheData
	if err := json.NewDecoder(file).Decode(&cacheData); err != nil {
		c.config.LogDebug("failed to parse cache file", "file", c.cacheFile, "error", err)
		return false
	}

	// Check if cache has expired
	currentTime := time.Now().Unix()
	if cacheData.ExpiresAt <= currentTime {
		c.config.LogDebug("cache has expired", "file", c.cacheFile)
		return false
	}

	c.config.LogDebug("cache is valid", "file", c.cacheFile)
	return true
}

// Get retrieves data from cache
func (c *Cache) Get() (map[string]interface{}, error) {
	lock, err := AcquireLock(c.config.LockFile)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire cache lock: %w", err)
	}
	defer ReleaseLock(lock)

	if !c.IsValid() {
		return nil, fmt.Errorf("cache is invalid or expired")
	}

	file, err := os.Open(c.cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache file: %w", err)
	}
	defer file.Close()

	var cacheData CacheData
	if err := json.NewDecoder(file).Decode(&cacheData); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %w", err)
	}

	// Parse the JSON string back to map
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(cacheData.FullData), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return result, nil
}

// Set stores data in cache
func (c *Cache) Set(data map[string]interface{}) error {
	lock, err := AcquireLock(c.config.LockFile)
	if err != nil {
		return fmt.Errorf("failed to acquire cache lock for update: %w", err)
	}
	defer ReleaseLock(lock)

	currentTime := time.Now()
	expiresAt := currentTime.Add(c.expiryDuration).Unix()

	// Convert data to JSON string
	dataBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data for cache: %w", err)
	}

	cacheData := CacheData{
		FullData:  string(dataBytes),
		ExpiresAt: expiresAt,
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(filepath.Dir(c.cacheFile), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	file, err := os.Create(c.cacheFile)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cacheData); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	c.config.LogDebug("cache updated successfully", "file", c.cacheFile)
	return nil
}

// Clear removes the cache file
func (c *Cache) Clear() error {
	lock, err := AcquireLock(c.config.LockFile)
	if err != nil {
		return fmt.Errorf("failed to acquire cache lock for clear: %w", err)
	}
	defer ReleaseLock(lock)

	if err := os.Remove(c.cacheFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cache file: %w", err)
	}

	c.config.LogDebug("cache cleared", "file", c.cacheFile)
	return nil
}

// GetExpiryTime returns the cache expiry time
func (c *Cache) GetExpiryTime() (time.Time, error) {
	if !c.IsValid() {
		return time.Time{}, fmt.Errorf("cache is invalid")
	}

	file, err := os.Open(c.cacheFile)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to open cache file: %w", err)
	}
	defer file.Close()

	var cacheData CacheData
	if err := json.NewDecoder(file).Decode(&cacheData); err != nil {
		return time.Time{}, fmt.Errorf("failed to parse cache file: %w", err)
	}

	return time.Unix(cacheData.ExpiresAt, 0), nil
}
