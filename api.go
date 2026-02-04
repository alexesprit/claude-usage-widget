package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

// APIClient handles API requests to Claude usage endpoint
type APIClient struct {
	config *Config
}

// NewAPIClient creates a new API client
func NewAPIClient(config *Config) *APIClient {
	return &APIClient{
		config: config,
	}
}

// FetchFullUsageData fetches complete usage data from the API
func (c *APIClient) FetchFullUsageData(ctx context.Context, sessionKey, orgId string) (map[string]interface{}, error) {
	// Create a context with timeout if none provided
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Build URL safely - validate orgId doesn't contain path traversal
	if strings.Contains(orgId, "..") || strings.Contains(orgId, "/") {
		return nil, NewValidationError("orgId", orgId, "organization ID contains invalid characters")
	}

	url := fmt.Sprintf("https://claude.ai/api/organizations/%s/usage", orgId)

	c.config.LogDebug("fetching usage data", "url", url)

	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Safari_16_0),
		tls_client.WithCookieJar(jar),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, NewAPIError("create_client", url, 0, "failed to create HTTP client", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, NewAPIError("create_request", url, 0, "failed to create request", err)
	}

	req.Header.Set("Cookie", fmt.Sprintf("sessionKey=%s", sessionKey))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Sec-Ch-Ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.Canceled {
			return nil, NewAPIError("request", url, 0, "request was canceled", ctx.Err())
		}
		if ctx.Err() == context.DeadlineExceeded {
			return nil, NewAPIError("request", url, 0, "request timed out", ctx.Err())
		}
		return nil, NewAPIError("request", url, 0, "failed to fetch usage data", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Read response body for debugging
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, NewAPIError("read_response", url, resp.StatusCode, "could not read response body", err)
		}
		return nil, NewAPIError("http_request", url, resp.StatusCode, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)), nil)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, NewAPIError("parse_response", url, resp.StatusCode, "invalid JSON response", err)
	}

	c.config.LogDebug("successfully fetched usage data")
	return result, nil
}

// FetchUsageData fetches usage data and extracts specific utilization and resets_at values
// This maintains backward compatibility with the original CLI interface
func (c *APIClient) FetchUsageData(ctx context.Context, sessionKey, orgId string) (utilization int, resetsAt string, err error) {
	data, err := c.FetchFullUsageData(ctx, sessionKey, orgId)
	if err != nil {
		return 0, "", err
	}

	// For backward compatibility, still return five_hour data
	fiveHour, ok := data["five_hour"].(map[string]interface{})
	if !ok {
		return 0, "", NewAPIError("parse_response", "", 200, "missing five_hour data in response", nil)
	}

	utilizationFloat, ok := fiveHour["utilization"].(float64)
	if !ok {
		return 0, "", NewAPIError("parse_response", "", 200, "utilization not found or not a number", nil)
	}
	utilization = int(utilizationFloat)

	if resetsAtVal, ok := fiveHour["resets_at"].(string); ok {
		resetsAt = resetsAtVal
	}

	return utilization, resetsAt, nil
}
