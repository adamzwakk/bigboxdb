package bbdbigdb // Named this way so it doesnt conflict with lib one

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
	"os"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type Client struct {
	clientID     string
	clientSecret string

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func NewClient() *Client {
	return &Client{}
}

// GetToken returns a cached access token, or fetches a new one if expired.
func (c *Client) GetToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	clientID := os.Getenv("TWITCH_CLIENT")
	clientSecret := os.Getenv("TWITCH_SECRET")

	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("TWITCH_CLIENT or TWITCH_SECRET env vars not set")
	}

	if c.token != "" && time.Now().Before(c.expiresAt) {
		return c.token, nil
	}

	url := fmt.Sprintf(
		"https://id.twitch.tv/oauth2/token?client_id=%s&client_secret=%s&grant_type=client_credentials",
		clientID,
		clientSecret,
	)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	c.token = tokenResp.AccessToken
	// Expire a minute early to avoid edge cases
	c.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn)*time.Second - time.Minute)

	return c.token, nil
}

// ClientID returns the Twitch client ID.
func (c *Client) ClientID() string {
	return os.Getenv("TWITCH_CLIENT")
}