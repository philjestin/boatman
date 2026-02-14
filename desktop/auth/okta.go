package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/browser"
)

// OktaAuth handles Okta OAuth authentication
type OktaAuth struct {
	Domain       string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	tokenCache   *TokenCache
}

// TokenCache stores OAuth tokens
type TokenCache struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope"`
}

// OktaConfig holds Okta configuration
type OktaConfig struct {
	Domain       string `json:"domain"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
}

// NewOktaAuth creates a new Okta OAuth handler
func NewOktaAuth(domain, clientID, clientSecret string) *OktaAuth {
	return &OktaAuth{
		Domain:       domain,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  "http://localhost:8484/callback",
	}
}

// IsAuthenticated checks if we have a valid token
func (o *OktaAuth) IsAuthenticated() bool {
	if o.tokenCache == nil {
		return false
	}
	return time.Now().Before(o.tokenCache.ExpiresAt)
}

// GetAccessToken returns the current access token
func (o *OktaAuth) GetAccessToken() (string, error) {
	if !o.IsAuthenticated() {
		return "", fmt.Errorf("not authenticated or token expired")
	}
	return o.tokenCache.AccessToken, nil
}

// Login initiates the OAuth flow
func (o *OktaAuth) Login(scopes []string) error {
	// Build authorization URL
	state := generateRandomState()
	authURL := fmt.Sprintf(
		"https://%s/oauth2/v1/authorize?client_id=%s&response_type=code&scope=%s&redirect_uri=%s&state=%s",
		o.Domain,
		o.ClientID,
		url.QueryEscape(joinScopes(scopes)),
		url.QueryEscape(o.RedirectURI),
		state,
	)

	// Start local server to receive callback
	resultChan := make(chan authResult)
	server := &http.Server{Addr: ":8484"}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		receivedState := r.URL.Query().Get("state")

		if receivedState != state {
			resultChan <- authResult{err: fmt.Errorf("state mismatch")}
			fmt.Fprintf(w, "Authentication failed: state mismatch")
			return
		}

		if code == "" {
			resultChan <- authResult{err: fmt.Errorf("no code received")}
			fmt.Fprintf(w, "Authentication failed: no code received")
			return
		}

		// Exchange code for token
		token, err := o.exchangeCode(code)
		if err != nil {
			resultChan <- authResult{err: err}
			fmt.Fprintf(w, "Authentication failed: %v", err)
			return
		}

		resultChan <- authResult{token: token}
		fmt.Fprintf(w, "Authentication successful! You can close this window.")
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			resultChan <- authResult{err: err}
		}
	}()

	// Open browser
	if err := browser.OpenURL(authURL); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	// Wait for callback
	result := <-resultChan

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	if result.err != nil {
		return result.err
	}

	o.tokenCache = result.token
	return nil
}

// exchangeCode exchanges authorization code for access token
func (o *OktaAuth) exchangeCode(code string) (*TokenCache, error) {
	tokenURL := fmt.Sprintf("https://%s/oauth2/v1/token", o.Domain)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", o.RedirectURI)
	data.Set("client_id", o.ClientID)
	if o.ClientSecret != "" {
		data.Set("client_secret", o.ClientSecret)
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s - %s", resp.Status, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &TokenCache{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Scope:        tokenResp.Scope,
	}, nil
}

// RefreshToken refreshes the access token using refresh token
func (o *OktaAuth) RefreshToken() error {
	if o.tokenCache == nil || o.tokenCache.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	tokenURL := fmt.Sprintf("https://%s/oauth2/v1/token", o.Domain)

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", o.tokenCache.RefreshToken)
	data.Set("client_id", o.ClientID)
	if o.ClientSecret != "" {
		data.Set("client_secret", o.ClientSecret)
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token refresh failed: %s - %s", resp.Status, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	o.tokenCache = &TokenCache{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Scope:        tokenResp.Scope,
	}

	return nil
}

// Revoke revokes the access token
func (o *OktaAuth) Revoke() error {
	if o.tokenCache == nil {
		return nil
	}

	revokeURL := fmt.Sprintf("https://%s/oauth2/v1/revoke", o.Domain)

	data := url.Values{}
	data.Set("token", o.tokenCache.AccessToken)
	data.Set("token_type_hint", "access_token")
	data.Set("client_id", o.ClientID)
	if o.ClientSecret != "" {
		data.Set("client_secret", o.ClientSecret)
	}

	resp, err := http.PostForm(revokeURL, data)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	defer resp.Body.Close()

	o.tokenCache = nil
	return nil
}

type authResult struct {
	token *TokenCache
	err   error
}

func generateRandomState() string {
	// Simple random state for CSRF protection
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func joinScopes(scopes []string) string {
	if len(scopes) == 0 {
		return "openid profile email"
	}
	result := "openid"
	for _, scope := range scopes {
		result += " " + scope
	}
	return result
}
