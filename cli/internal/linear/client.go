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
	"time"

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

// Comment represents a Linear issue comment.
type Comment struct {
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	UserName  string    `json:"userName"`
}

// FullTicket extends Ticket with additional fields needed for triage.
type FullTicket struct {
	Ticket
	Comments    []Comment `json:"comments"`
	Team        string    `json:"team"`
	ProjectName string    `json:"projectName"`
	Estimate    *float64  `json:"estimate"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ListOptions configures batch ticket fetching.
type ListOptions struct {
	TeamKeys []string // Filter by team key (e.g., "ENG", "FE")
	States   []string // Filter by state type (e.g., "backlog", "triage", "unstarted")
	Limit    int      // Max tickets to fetch (0 = 50)
	Labels   []string // Optional label filter
}

// ListTickets fetches multiple tickets matching the given filters with pagination.
func (c *Client) ListTickets(ctx context.Context, opts ListOptions) ([]FullTicket, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}

	// Build the filter object dynamically
	filter := map[string]interface{}{}

	if len(opts.TeamKeys) > 0 {
		filter["team"] = map[string]interface{}{
			"key": map[string]interface{}{
				"in": opts.TeamKeys,
			},
		}
	}

	if len(opts.States) > 0 {
		filter["state"] = map[string]interface{}{
			"type": map[string]interface{}{
				"in": opts.States,
			},
		}
	}

	if len(opts.Labels) > 0 {
		filter["labels"] = map[string]interface{}{
			"name": map[string]interface{}{
				"in": opts.Labels,
			},
		}
	}

	var allTickets []FullTicket
	var cursor *string
	pageSize := 50
	if limit < pageSize {
		pageSize = limit
	}

	for {
		remaining := limit - len(allTickets)
		if remaining <= 0 {
			break
		}
		fetchCount := pageSize
		if remaining < fetchCount {
			fetchCount = remaining
		}

		tickets, nextCursor, err := c.listTicketsPage(ctx, filter, fetchCount, cursor)
		if err != nil {
			return allTickets, err
		}

		allTickets = append(allTickets, tickets...)

		if nextCursor == nil || len(tickets) < fetchCount {
			break
		}
		cursor = nextCursor
	}

	return allTickets, nil
}

// listTicketsPage fetches a single page of tickets.
func (c *Client) listTicketsPage(ctx context.Context, filter map[string]interface{}, first int, after *string) ([]FullTicket, *string, error) {
	query := `
		query ListIssues($filter: IssueFilter, $first: Int!, $after: String) {
			issues(filter: $filter, first: $first, after: $after, orderBy: updatedAt) {
				nodes {
					id
					identifier
					title
					description
					branchName
					priority
					estimate
					updatedAt
					createdAt
					state {
						name
						type
					}
					labels {
						nodes {
							name
						}
					}
					team {
						key
					}
					project {
						name
					}
					comments(first: 5) {
						nodes {
							body
							createdAt
							user {
								name
							}
						}
					}
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	`

	variables := map[string]interface{}{
		"filter": filter,
		"first":  first,
	}
	if after != nil {
		variables["after"] = *after
	}

	resp, err := c.execute(ctx, query, variables)
	if err != nil {
		return nil, nil, err
	}

	var result struct {
		Data struct {
			Issues struct {
				Nodes []struct {
					ID          string   `json:"id"`
					Identifier  string   `json:"identifier"`
					Title       string   `json:"title"`
					Description string   `json:"description"`
					BranchName  string   `json:"branchName"`
					Priority    int      `json:"priority"`
					Estimate    *float64 `json:"estimate"`
					UpdatedAt   string   `json:"updatedAt"`
					CreatedAt   string   `json:"createdAt"`
					State       struct {
						Name string `json:"name"`
						Type string `json:"type"`
					} `json:"state"`
					Labels struct {
						Nodes []struct {
							Name string `json:"name"`
						} `json:"nodes"`
					} `json:"labels"`
					Team struct {
						Key string `json:"key"`
					} `json:"team"`
					Project *struct {
						Name string `json:"name"`
					} `json:"project"`
					Comments struct {
						Nodes []struct {
							Body      string `json:"body"`
							CreatedAt string `json:"createdAt"`
							User      *struct {
								Name string `json:"name"`
							} `json:"user"`
						} `json:"nodes"`
					} `json:"comments"`
				} `json:"nodes"`
				PageInfo struct {
					HasNextPage bool    `json:"hasNextPage"`
					EndCursor   *string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"issues"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, nil, fmt.Errorf("linear API error: %s", result.Errors[0].Message)
	}

	tickets := make([]FullTicket, 0, len(result.Data.Issues.Nodes))
	for _, issue := range result.Data.Issues.Nodes {
		labels := make([]string, len(issue.Labels.Nodes))
		for i, l := range issue.Labels.Nodes {
			labels[i] = l.Name
		}

		comments := make([]Comment, 0, len(issue.Comments.Nodes))
		for _, c := range issue.Comments.Nodes {
			userName := ""
			if c.User != nil {
				userName = c.User.Name
			}
			createdAt, _ := time.Parse(time.RFC3339, c.CreatedAt)
			comments = append(comments, Comment{
				Body:      c.Body,
				CreatedAt: createdAt,
				UserName:  userName,
			})
		}

		projectName := ""
		if issue.Project != nil {
			projectName = issue.Project.Name
		}

		updatedAt, _ := time.Parse(time.RFC3339, issue.UpdatedAt)
		createdAt, _ := time.Parse(time.RFC3339, issue.CreatedAt)

		tickets = append(tickets, FullTicket{
			Ticket: Ticket{
				ID:          issue.ID,
				Identifier:  issue.Identifier,
				Title:       issue.Title,
				Description: issue.Description,
				State:       issue.State.Name,
				Priority:    issue.Priority,
				Labels:      labels,
				BranchName:  issue.BranchName,
			},
			Comments:    comments,
			Team:        issue.Team.Key,
			ProjectName: projectName,
			Estimate:    issue.Estimate,
			UpdatedAt:   updatedAt,
			CreatedAt:   createdAt,
		})
	}

	var nextCursor *string
	if result.Data.Issues.PageInfo.HasNextPage {
		nextCursor = result.Data.Issues.PageInfo.EndCursor
	}

	return tickets, nextCursor, nil
}

// AddComment posts a comment to a Linear issue.
// issueID must be the UUID (Ticket.ID), not the human-readable identifier.
func (c *Client) AddComment(ctx context.Context, issueID string, body string) error {
	query := `
		mutation CreateComment($issueId: String!, $body: String!) {
			commentCreate(input: { issueId: $issueId, body: $body }) {
				success
				comment {
					id
				}
			}
		}
	`

	variables := map[string]interface{}{
		"issueId": issueID,
		"body":    body,
	}

	resp, err := c.execute(ctx, query, variables)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	var result struct {
		Data struct {
			CommentCreate struct {
				Success bool `json:"success"`
			} `json:"commentCreate"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("failed to parse comment response: %w", err)
	}

	if len(result.Errors) > 0 {
		return fmt.Errorf("linear API error: %s", result.Errors[0].Message)
	}

	if !result.Data.CommentCreate.Success {
		return fmt.Errorf("comment creation failed")
	}

	return nil
}

// GetFullTicket fetches a single ticket with all triage-relevant fields.
func (c *Client) GetFullTicket(ctx context.Context, identifier string) (*FullTicket, error) {
	teamKey, number, ok := parseIdentifier(identifier)
	if !ok {
		return nil, fmt.Errorf("invalid identifier format: %s (expected TEAM-123)", identifier)
	}

	filter := map[string]interface{}{
		"team": map[string]interface{}{
			"key": map[string]interface{}{"eq": teamKey},
		},
		"number": map[string]interface{}{"eq": float64(number)},
	}

	tickets, _, err := c.listTicketsPage(ctx, filter, 1, nil)
	if err != nil {
		return nil, err
	}

	if len(tickets) == 0 {
		return nil, fmt.Errorf("issue not found: %s", identifier)
	}

	return &tickets[0], nil
}
