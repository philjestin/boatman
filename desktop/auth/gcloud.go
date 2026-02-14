package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GCloudAuth handles Google Cloud authentication
type GCloudAuth struct{}

// NewGCloudAuth creates a new GCloud auth handler
func NewGCloudAuth() *GCloudAuth {
	return &GCloudAuth{}
}

// IsInstalled checks if gcloud CLI is installed
func (g *GCloudAuth) IsInstalled() bool {
	_, err := exec.LookPath("gcloud")
	return err == nil
}

// IsAuthenticated checks if user is authenticated with gcloud
func (g *GCloudAuth) IsAuthenticated() (bool, error) {
	if !g.IsInstalled() {
		return false, fmt.Errorf("gcloud CLI not installed")
	}

	// Check if application default credentials exist
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}

	// Check for application_default_credentials.json
	adcPath := filepath.Join(homeDir, ".config", "gcloud", "application_default_credentials.json")
	if _, err := os.Stat(adcPath); err == nil {
		return true, nil
	}

	// Also check legacy path
	legacyPath := filepath.Join(homeDir, ".config", "gcloud", "legacy_credentials")
	if _, err := os.Stat(legacyPath); err == nil {
		return true, nil
	}

	return false, nil
}

// GetAuthInfo returns information about current authentication
func (g *GCloudAuth) GetAuthInfo() (map[string]interface{}, error) {
	if !g.IsInstalled() {
		return nil, fmt.Errorf("gcloud CLI not installed")
	}

	// Get active account
	cmd := exec.Command("gcloud", "config", "get-value", "account")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	account := strings.TrimSpace(string(output))

	// Get active project
	cmd = exec.Command("gcloud", "config", "get-value", "project")
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	project := strings.TrimSpace(string(output))

	// Check if ADC exists
	isAuth, _ := g.IsAuthenticated()

	return map[string]interface{}{
		"account":       account,
		"project":       project,
		"authenticated": isAuth,
		"adcConfigured": isAuth,
	}, nil
}

// Login triggers gcloud auth login flow
func (g *GCloudAuth) Login() error {
	if !g.IsInstalled() {
		return fmt.Errorf("gcloud CLI not installed")
	}

	// Run gcloud auth login
	cmd := exec.Command("gcloud", "auth", "login")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// LoginApplicationDefault triggers gcloud auth application-default login
func (g *GCloudAuth) LoginApplicationDefault() error {
	if !g.IsInstalled() {
		return fmt.Errorf("gcloud CLI not installed")
	}

	// Run gcloud auth application-default login
	cmd := exec.Command("gcloud", "auth", "application-default", "login")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// SetProject sets the active GCP project
func (g *GCloudAuth) SetProject(projectID string) error {
	if !g.IsInstalled() {
		return fmt.Errorf("gcloud CLI not installed")
	}

	cmd := exec.Command("gcloud", "config", "set", "project", projectID)
	return cmd.Run()
}

// GetAvailableProjects returns list of available GCP projects
func (g *GCloudAuth) GetAvailableProjects() ([]string, error) {
	if !g.IsInstalled() {
		return nil, fmt.Errorf("gcloud CLI not installed")
	}

	cmd := exec.Command("gcloud", "projects", "list", "--format=json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var projects []struct {
		ProjectID string `json:"projectId"`
		Name      string `json:"name"`
	}

	if err := json.Unmarshal(output, &projects); err != nil {
		return nil, err
	}

	projectIDs := make([]string, len(projects))
	for i, p := range projects {
		projectIDs[i] = p.ProjectID
	}

	return projectIDs, nil
}

// VerifyVertexAIAccess checks if user has access to Vertex AI
func (g *GCloudAuth) VerifyVertexAIAccess(projectID, region string) error {
	if !g.IsInstalled() {
		return fmt.Errorf("gcloud CLI not installed")
	}

	// Try to list Vertex AI endpoints to verify access
	cmd := exec.Command(
		"gcloud", "ai", "endpoints", "list",
		"--project", projectID,
		"--region", region,
		"--format=json",
	)

	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("failed to verify Vertex AI access: %w", err)
	}

	return nil
}

// Revoke revokes authentication
func (g *GCloudAuth) Revoke() error {
	if !g.IsInstalled() {
		return fmt.Errorf("gcloud CLI not installed")
	}

	// Revoke application default credentials
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	adcPath := filepath.Join(homeDir, ".config", "gcloud", "application_default_credentials.json")
	if err := os.Remove(adcPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
