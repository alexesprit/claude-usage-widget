package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	// Define flags
	flag5h := flag.Bool("5h", false, "Display 5-hour session usage")
	flagWeekly := flag.Bool("weekly", false, "Display weekly session usage")
	flagConfig := flag.Bool("c", false, "Show configuration file path")
	flag.BoolVar(flagConfig, "config", false, "Show configuration file path")
	flagVerbose := flag.Bool("v", false, "Enable verbose output")
	flag.BoolVar(flagVerbose, "verbose", false, "Enable verbose output")
	flag.Parse()

	// Handle config path display command
	if *flagConfig {
		config, err := LoadConfig()
		if err != nil {
			fmt.Printf("ERROR:Failed to load configuration: %s\n", UserFriendlyError(err))
			os.Exit(1)
		}
		fmt.Println(config.ConfigFile)
		os.Exit(0)
	}

	// Load verbose from env or flag
	verbose := *flagVerbose
	if v := os.Getenv("CLAUDE_CODE_USAGE_VERBOSE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			verbose = verbose || b
		}
	}

	// Validate flags: must specify exactly one
	if *flag5h && *flagWeekly {
		fmt.Println(ErrMultipleSessions.Error())
		fmt.Print("\033[33mUsage: ~\033[0m\n")
		os.Exit(1)
	}
	if !*flag5h && !*flagWeekly {
		fmt.Println(ErrNoSessionSpecified.Error())
		fmt.Print("\033[33mUsage: ~\033[0m\n")
		os.Exit(1)
	}

	// Determine session key and window hours based on flag
	var sessionKey string
	var windowHours int
	if *flag5h {
		sessionKey = "five_hour"
		windowHours = 5
	} else {
		sessionKey = "seven_day"
		windowHours = 168
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		if verbose {
			fmt.Printf("ERROR:Failed to load configuration: %s\n", UserFriendlyError(err))
		} else {
			fmt.Print("\033[33mUsage: ~\033[0m\n")
		}
		os.Exit(1)
	}

	config.LogDebug("configuration loaded successfully", "config_file", config.ConfigFile)
	config.LogDebug("session determined", "session_key", sessionKey, "window_hours", windowHours)
	config.LogDebug("session determined", "session_key", sessionKey, "window_hours", windowHours)

	// Override verbose from flag/env
	config.Verbose = verbose

	// Set up signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize components
	apiClient := NewAPIClient(config)
	cache := NewCache(config.CacheFile, config.CacheExpirySeconds, config)

	// Create context with timeout for API operations
	apiCtx, apiCancel := context.WithTimeout(ctx, 45*time.Second)
	defer apiCancel()

	// Get usage data (cached or fresh)
	var data map[string]interface{}
	if cachedData, err := cache.Get(); err == nil {
		data = cachedData
		config.LogDebug("using cached data")
	} else {
		// Fetch fresh data
		freshData, err := apiClient.FetchFullUsageData(apiCtx, config.SessionKey, config.OrganizationID)
		if err != nil {
			fmt.Printf("ERROR:%s\n", UserFriendlyError(err))
			fmt.Print("\033[33mUsage: ~\033[0m\n")
			os.Exit(1)
		}
		data = freshData

		// Update cache
		if cacheErr := cache.Set(data); cacheErr != nil {
			config.LogError("failed to update cache", "error", cacheErr)
		}
	}

	// Extract session data
	sessionDataRaw, ok := data[sessionKey]
	if !ok {
		fmt.Printf("ERROR:No data available for session '%s'\n", sessionKey)
		fmt.Print("\033[33mUsage: ~\033[0m\n")
		os.Exit(1)
	}

	// Convert to map for extraction
	sessionDataMap, ok := sessionDataRaw.(map[string]interface{})
	if !ok {
		fmt.Printf("ERROR:Invalid data format for session '%s'\n", sessionKey)
		fmt.Print("\033[33mUsage: ~\033[0m\n")
		os.Exit(1)
	}

	utilizationFloat, ok := sessionDataMap["utilization"].(float64)
	if !ok {
		fmt.Printf("ERROR:Invalid utilization data for session '%s'\n", sessionKey)
		fmt.Print("\033[33mUsage: ~\033[0m\n")
		os.Exit(1)
	}
	utilization := int(utilizationFloat)

	resetsAtStr, _ := sessionDataMap["resets_at"].(string)

	config.LogDebug("session data extracted", "session_key", sessionKey, "utilization", utilization, "resets_at", resetsAtStr)

	// Calculate burn prediction if resets_at is available
	burnPrediction := ""
	if resetsAtStr != "" {
		resetsAt, err := time.Parse(time.RFC3339, resetsAtStr)
		if err == nil {
			if prediction, calcErr := CalculateBurnPrediction(utilization, resetsAt, windowHours); calcErr == nil {
				burnPrediction = prediction
				config.LogDebug("burn prediction calculated", "prediction", prediction)
			} else {
				config.LogDebug("failed to calculate burn prediction", "error", calcErr)
			}
		} else {
			config.LogDebug("failed to parse resets_at timestamp", "resets_at", resetsAtStr, "error", err)
		}
	} else {
		config.LogDebug("no resets_at available for burn prediction")
	}

	// Load display configuration
	displayConfig, err := LoadDisplayConfig(config, sessionKey)
	if err != nil {
		fmt.Printf("ERROR:Failed to load display configuration: %s\n", UserFriendlyError(err))
		fmt.Print("\033[33mUsage: ~\033[0m\n")
		os.Exit(1)
	}

	// Process the session and display output
	output, err := ProcessSession(sessionKey, data, burnPrediction, displayConfig)
	if err != nil {
		fmt.Printf("ERROR:Failed to process session data: %s\n", UserFriendlyError(err))
		fmt.Print("\033[33mUsage: ~\033[0m\n")
		os.Exit(1)
	}

	fmt.Println(output)
}
