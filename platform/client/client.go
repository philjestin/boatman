// Package client provides an HTTP client for the Boatman platform API.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	harnessmemory "github.com/philjestin/boatman-ecosystem/harness/memory"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

// Client is an HTTP client for the Boatman platform API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	scope      storage.Scope
}

// New creates a new platform client.
func New(baseURL string, scope storage.Scope) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		scope: scope,
	}
}

// Ping checks if the platform server is reachable.
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned %d", resp.StatusCode)
	}
	return nil
}

// CreateRun creates a new run record on the platform.
func (c *Client) CreateRun(ctx context.Context, run *storage.Run) error {
	return c.post(ctx, "/api/v1/runs", run, nil)
}

// UpdateRun updates an existing run.
func (c *Client) UpdateRun(ctx context.Context, run *storage.Run) error {
	return c.post(ctx, "/api/v1/runs", run, nil)
}

// GetEffectivePolicy returns the merged effective policy for the client's scope.
func (c *Client) GetEffectivePolicy(ctx context.Context) (*storage.Policy, error) {
	var policy storage.Policy
	if err := c.get(ctx, "/api/v1/policies/effective", &policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

// GetMemory retrieves merged memory as a harness Memory struct.
func (c *Client) GetMemory(ctx context.Context) (*harnessmemory.Memory, error) {
	// Get patterns
	var patterns []*storage.Pattern
	if err := c.get(ctx, "/api/v1/memory/patterns", &patterns); err != nil {
		return nil, fmt.Errorf("get patterns: %w", err)
	}

	// Get preferences
	var prefs storage.Preferences
	if err := c.get(ctx, "/api/v1/memory/preferences", &prefs); err != nil {
		return nil, fmt.Errorf("get preferences: %w", err)
	}

	// Convert to harness Memory
	mem := &harnessmemory.Memory{
		ProjectID:    fmt.Sprintf("%s/%s/%s", c.scope.OrgID, c.scope.TeamID, c.scope.RepoID),
		FilePatterns: make(map[string][]string),
		Preferences: harnessmemory.Preferences{
			PreferredTestFramework: prefs.PreferredTestFramework,
			NamingConventions:      prefs.NamingConventions,
			FileOrganization:       prefs.FileOrganization,
			CodeStyle:              prefs.CodeStyle,
			CommitMessageFormat:    prefs.CommitMessageFormat,
			ReviewerThresholds:     prefs.ReviewerThresholds,
		},
	}

	for _, p := range patterns {
		mem.Patterns = append(mem.Patterns, harnessmemory.Pattern{
			ID:          p.ID,
			Type:        p.Type,
			Description: p.Description,
			Example:     p.Example,
			FileMatcher: p.FileMatcher,
			Weight:      p.Weight,
			UsageCount:  p.UsageCount,
			SuccessRate: p.SuccessRate,
		})
	}

	return mem, nil
}

// RecordUsage records a usage record on the platform.
func (c *Client) RecordUsage(ctx context.Context, record *storage.UsageRecord) error {
	return c.post(ctx, "/api/v1/costs/usage", record, nil)
}

// CheckBudget returns the current budget status.
func (c *Client) CheckBudget(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := c.get(ctx, "/api/v1/costs/budget", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// get performs an authenticated GET request.
func (c *Client) get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s: %d %s", path, resp.StatusCode, string(body))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// post performs an authenticated POST/PUT request.
func (c *Client) post(ctx context.Context, path string, body any, result any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("POST %s: %d %s", path, resp.StatusCode, string(respBody))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("X-Boatman-Org", c.scope.OrgID)
	req.Header.Set("X-Boatman-Team", c.scope.TeamID)
	req.Header.Set("X-Boatman-Repo", c.scope.RepoID)
}
