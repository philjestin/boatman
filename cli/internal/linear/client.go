// Package linear provides a client for the Linear API.
package linear

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/philjestin/boatmanmode/internal/retry"
)

const apiURL = "https://api.linear.app/graphql"

// Client is a Linear API client.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// Ticket represents a Linear issue/ticket.
type Ticket struct {
	ID          string   `json:"id"`
	Identifier  string   `json:"identifier"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	State       string   `json:"state"`
	Priority    int      `json:"priority"`
	Labels      []string `json:"labels"`
	BranchName  string   `json:"branchName"`
}

// New creates a new Linear client.
func New(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// GetTicket fetches a ticket by its identifier (e.g., "ENG-123") or UUID.
// Human-readable identifiers like "ENG-123" are looked up via team key + issue
// number filters. UUID strings fall back to the issue(id:) query.
func (c *Client) GetTicket(ctx context.Context, identifier string) (*Ticket, error) {
	if teamKey, number, ok := parseIdentifier(identifier); ok {
		return c.getTicketByFilter(ctx, teamKey, number)
	}
	return c.getTicketByID(ctx, identifier)
}

// parseIdentifier splits a Linear identifier like "ENG-123" into team key and number.
func parseIdentifier(identifier string) (string, int, bool) {
	idx := strings.LastIndex(identifier, "-")
	if idx <= 0 || idx >= len(identifier)-1 {
		return "", 0, false
	}
	teamKey := identifier[:idx]
	numStr := identifier[idx+1:]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return "", 0, false
	}
	return teamKey, num, true
}

// getTicketByFilter fetches a ticket using team key and issue number filters.
// This is the correct way to look up issues by human-readable identifiers like "ENG-123",
// since the issue(id:) query expects a UUID.
func (c *Client) getTicketByFilter(ctx context.Context, teamKey string, number int) (*Ticket, error) {
	query := `
		query GetIssueByIdentifier($teamKey: String!, $number: Float!) {
			issues(filter: {
				team: { key: { eq: $teamKey } }
				number: { eq: $number }
			}, first: 1) {
				nodes {
					id
					identifier
					title
					description
					branchName
					priority
					state {
						name
					}
					labels {
						nodes {
							name
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"teamKey": teamKey,
		"number":  float64(number),
	}

	resp, err := c.execute(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Issues struct {
				Nodes []struct {
					ID          string `json:"id"`
					Identifier  string `json:"identifier"`
					Title       string `json:"title"`
					Description string `json:"description"`
					BranchName  string `json:"branchName"`
					Priority    int    `json:"priority"`
					State       struct {
						Name string `json:"name"`
					} `json:"state"`
					Labels struct {
						Nodes []struct {
							Name string `json:"name"`
						} `json:"nodes"`
					} `json:"labels"`
				} `json:"nodes"`
			} `json:"issues"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("linear API error: %s", result.Errors[0].Message)
	}

	if len(result.Data.Issues.Nodes) == 0 {
		return nil, fmt.Errorf("issue not found: %s-%d", teamKey, number)
	}

	issue := result.Data.Issues.Nodes[0]
	labels := make([]string, len(issue.Labels.Nodes))
	for i, l := range issue.Labels.Nodes {
		labels[i] = l.Name
	}

	return &Ticket{
		ID:          issue.ID,
		Identifier:  issue.Identifier,
		Title:       issue.Title,
		Description: issue.Description,
		State:       issue.State.Name,
		Priority:    issue.Priority,
		Labels:      labels,
		BranchName:  issue.BranchName,
	}, nil
}

// getTicketByID fetches a ticket by its UUID using the issue(id:) query.
func (c *Client) getTicketByID(ctx context.Context, id string) (*Ticket, error) {
	query := `
		query GetIssue($id: String!) {
			issue(id: $id) {
				id
				identifier
				title
				description
				branchName
				priority
				state {
					name
				}
				labels {
					nodes {
						name
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"id": id,
	}

	resp, err := c.execute(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Issue struct {
				ID          string `json:"id"`
				Identifier  string `json:"identifier"`
				Title       string `json:"title"`
				Description string `json:"description"`
				BranchName  string `json:"branchName"`
				Priority    int    `json:"priority"`
				State       struct {
					Name string `json:"name"`
				} `json:"state"`
				Labels struct {
					Nodes []struct {
						Name string `json:"name"`
					} `json:"nodes"`
				} `json:"labels"`
			} `json:"issue"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("linear API error: %s", result.Errors[0].Message)
	}

	issue := result.Data.Issue
	labels := make([]string, len(issue.Labels.Nodes))
	for i, l := range issue.Labels.Nodes {
		labels[i] = l.Name
	}

	return &Ticket{
		ID:          issue.ID,
		Identifier:  issue.Identifier,
		Title:       issue.Title,
		Description: issue.Description,
		State:       issue.State.Name,
		Priority:    issue.Priority,
		Labels:      labels,
		BranchName:  issue.BranchName,
	}, nil
}

// execute performs a GraphQL request to Linear with retry logic.
func (c *Client) execute(ctx context.Context, query string, variables map[string]interface{}) ([]byte, error) {
	body := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var result []byte

	err = retry.Do(ctx, retry.APIConfig(), "Linear API request", func() error {
		req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(jsonBody))
		if err != nil {
			return retry.Permanent(fmt.Errorf("failed to create request: %w", err))
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", c.apiKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err) // Retryable
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		// 4xx errors are permanent (client errors)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return retry.Permanent(fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody)))
		}

		// 5xx errors are retryable (server errors)
		if resp.StatusCode >= 500 {
			return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
		}

		if resp.StatusCode != http.StatusOK {
			return retry.Permanent(fmt.Errorf("API returned unexpected status %d: %s", resp.StatusCode, string(respBody)))
		}

		result = respBody
		return nil
	})

	return result, err
}

// isRetryableError checks if an error message indicates a retryable condition.
func isRetryableError(msg string) bool {
	retryablePatterns := []string{
		"rate limit",
		"timeout",
		"temporarily unavailable",
		"service unavailable",
		"internal server error",
	}
	lower := strings.ToLower(msg)
	for _, pattern := range retryablePatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}
