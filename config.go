package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	jsonparser "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config represents the application configuration
type Config struct {
	CacheExpirySeconds int       `json:"cache_expiry_seconds" koanf:"cache_expiry_seconds"`
	ConfigFile         string    `koanf:"config_file"`
	CacheFile          string    `json:"cache_file" koanf:"cache_file"`
	LockFile           string    `json:"lock_file" koanf:"lock_file"`
	SessionKey         string    `json:"session_key" koanf:"session_key"`
	OrganizationID     string    `json:"organization_id" koanf:"organization_id"`
	Debug              bool      `json:"debug" koanf:"debug"`
	Verbose            bool      `json:"verbose" koanf:"verbose"`
	Sessions           []Session `json:"sessions" koanf:"sessions"`
	Logger             *slog.Logger
}

type Session struct {
	Key    string `json:"key" koanf:"key"`
	Name   string `json:"name" koanf:"name"`
	Format Format `json:"format" koanf:"format"`
}

type Format struct {
	Template string `json:"template" koanf:"template"`
}

// LoadConfig loads configuration from JSON file and environment variables
func LoadConfig() (*Config, error) {
	k := koanf.New(".")

	// Determine config file paths
	var configPaths []string
	if configDir, err := os.UserConfigDir(); err == nil {
		configPaths = append(configPaths, filepath.Join(configDir, "claude-usage-widget", "config.json"))
	}
	configPaths = append(configPaths, "usage_config.json")

	// Determine cache file paths
	var cacheFilePath, lockFilePath string
	if cacheDir, err := os.UserCacheDir(); err == nil {
		cacheFilePath = filepath.Join(cacheDir, "claude-usage-widget", "cache.json")
		lockFilePath = filepath.Join(cacheDir, "claude-usage-widget", "cache.json.lock")
	} else {
		cacheFilePath = ".usage_cache.json"
		lockFilePath = ".usage_cache.json.lock"
	}

	// Default values
	config := &Config{
		CacheExpirySeconds: 60,
		ConfigFile:         configPaths[0], // Use the first path as default
		CacheFile:          cacheFilePath,
		LockFile:           lockFilePath,
		SessionKey:         "",
		OrganizationID:     "",
		Debug:              false,
		Verbose:            false,
	}

	// Try to load from config files
	for _, path := range configPaths {
		if err := k.Load(file.Provider(path), jsonparser.Parser()); err == nil {
			config.ConfigFile = path // Set to the successfully loaded path
			break
		}
		// If loading fails, continue to next path or use defaults
	}

	// Load from environment variables (these take precedence)
	if err := k.Load(env.Provider("", ".", func(s string) string {
		// Convert env var names to config keys
		switch s {
		case "CLAUDE_CODE_USAGE_SESSION_KEY":
			return "session_key"
		case "CLAUDE_CODE_USAGE_ORGANIZATION_ID":
			return "organization_id"
		case "CLAUDE_CODE_USAGE_CACHE_EXPIRY_SECONDS":
			return "cache_expiry_seconds"
		case "CLAUDE_CODE_USAGE_VERBOSE":
			return "verbose"
		case "CLAUDE_CODE_USAGE_CONFIG_FILE":
			return "config_file"
		case "CLAUDE_CODE_USAGE_CACHE_FILE":
			return "cache_file"
		case "CLAUDE_CODE_USAGE_LOCK_FILE":
			return "lock_file"
		case "CLAUDE_CODE_USAGE_DEBUG":
			return "debug"
		default:
			return ""
		}
	}), nil); err != nil {
		return nil, fmt.Errorf("failed to load env config: %w", err)
	}

	// Unmarshal into config struct
	if err := k.Unmarshal("", config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set default sessions if none loaded
	if len(config.Sessions) == 0 {
		config.Sessions = []Session{
			{
				Key:    "five_hour",
				Name:   "5-hour",
				Format: Format{Template: "{name}: {percent} {progress} {time} {burn}"},
			},
			{
				Key:    "seven_day",
				Name:   "Weekly",
				Format: Format{Template: "{name}: {percent} {progress} {time}"},
			},
		}
	}

	// Save default configuration to file if it doesn't exist
	if _, err := os.Stat(configPaths[0]); os.IsNotExist(err) {
		configDir := filepath.Dir(configPaths[0])
		if err := os.MkdirAll(configDir, 0755); err != nil {
			if config.Verbose {
				fmt.Fprintf(os.Stderr, "[ERROR] Failed to create config directory %s: %v\n", configDir, err)
			}
		} else {
			config.ConfigFile = "" // Don't persist the config file path
			data, err := json.MarshalIndent(config, "", "  ")
			if err != nil {
				if config.Verbose {
					fmt.Fprintf(os.Stderr, "[ERROR] Failed to marshal default config: %v\n", err)
				}
			} else {
				if err := os.WriteFile(configPaths[0], data, 0644); err != nil {
					if config.Verbose {
						fmt.Fprintf(os.Stderr, "[ERROR] Failed to write default config to %s: %v\n", configPaths[0], err)
					}
				}
			}
		}
	}

	// Validate required fields
	if config.SessionKey == "" {
		return nil, NewConfigError("session_key", "session key is required", nil)
	}
	if config.OrganizationID == "" {
		return nil, NewConfigError("organization_id", "organization ID is required", nil)
	}

	// Handle string to int conversion for cache_expiry_seconds
	if expiryStr := os.Getenv("CLAUDE_CODE_USAGE_CACHE_EXPIRY_SECONDS"); expiryStr != "" {
		expiry, err := strconv.Atoi(expiryStr)
		if err != nil {
			return nil, fmt.Errorf("invalid CLAUDE_CODE_USAGE_CACHE_EXPIRY_SECONDS: %w", err)
		}
		config.CacheExpirySeconds = expiry
	}

	// Handle string to bool conversion for debug
	if debugStr := os.Getenv("CLAUDE_CODE_USAGE_DEBUG"); debugStr != "" {
		if debug, err := strconv.ParseBool(debugStr); err == nil {
			config.Debug = debug
		}
	}

	// Log configuration for debugging
	if config.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Config loaded - CacheFile: %s, LockFile: %s, ExpirySeconds: %d\n",
			config.CacheFile, config.LockFile, config.CacheExpirySeconds)
	}

	// Initialize logger
	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if config.Debug {
		handlerOpts.Level = slog.LevelDebug
	}
	handler := slog.NewTextHandler(os.Stderr, handlerOpts)
	config.Logger = slog.New(handler)

	return config, nil
}

// LogDebug logs debug messages if debug is enabled
func (c *Config) LogDebug(message string, args ...any) {
	if c.Logger != nil {
		c.Logger.Debug(message, args...)
	}
}

// LogError logs error messages
func (c *Config) LogError(message string, args ...any) {
	if c.Verbose {
		if c.Logger != nil {
			c.Logger.Error(message, args...)
		} else {
			fmt.Fprintf(os.Stderr, "[ERROR] %s\n", message)
		}
	}
}
