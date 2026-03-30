package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	oauthConfigOnce sync.Once
	oauthConfig     *oauth2.Config
)

// getOAuthConfig lazily initializes and returns the OAuth config
func getOAuthConfig() *oauth2.Config {
	oauthConfigOnce.Do(func() {
		oauthConfig = &oauth2.Config{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			Scopes:       []string{"user:email", "read:user"},
			Endpoint:     github.Endpoint,
			RedirectURL:  os.Getenv("GITHUB_CALLBACK_URL"),
		}
	})
	return oauthConfig
}

// GetLoginURL generates the OAuth authorization URL
func GetLoginURL(state string) string {
	return getOAuthConfig().AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeToken exchanges the authorization code for an access token
func ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return getOAuthConfig().Exchange(ctx, code)
}

// GetHTTPClient creates an HTTP client with the OAuth token
func GetHTTPClient(ctx context.Context, token *oauth2.Token) *http.Client {
	return getOAuthConfig().Client(ctx, token)
}

// GitHubUser represents a GitHub user response
type GitHubUser struct {
	ID        json.Number `json:"id"`
	Name      string      `json:"name"`
	Login     string      `json:"login"`
	AvatarURL string      `json:"avatar_url"`
	Email     string      `json:"email"`
}

// GitHubEmail represents a GitHub email response
type GitHubEmail struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

// FetchUser fetches the authenticated user's profile from GitHub
func FetchUser(ctx context.Context, client *http.Client) (*GitHubUser, error) {
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}

// FetchEmails fetches the authenticated user's email addresses from GitHub
func FetchEmails(ctx context.Context, client *http.Client) ([]GitHubEmail, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return nil, fmt.Errorf("failed to decode emails: %w", err)
	}

	return emails, nil
}

// FindPrimaryEmail finds the primary verified email from a list of emails
func FindPrimaryEmail(emails []GitHubEmail) string {
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email
		}
	}
	return ""
}

// IsValidStudentEmail checks if the email ends with the student domain (case-insensitive)
func IsValidStudentEmail(email string) bool {
	return strings.HasSuffix(strings.ToLower(email), "@student.unand.ac.id")
}

// ExtractNIM extracts the NIM from a student email
func ExtractNIM(email string) string {
	// email format: NIM@student.unand.ac.id
	for i, c := range email {
		if c == '@' {
			return email[:i]
		}
	}
	return ""
}

// AssignRole assigns a role based on NIM
func AssignRole(nim string) string {
	if len(nim) >= 6 && nim[2:6] == "1152" {
		return "internal"
	}
	return "external"
}

// FetchAvatar downloads the avatar from GitHub URL with timeout and size limit
func FetchAvatar(avatarURL string) ([]byte, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request with context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", avatarURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create avatar request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch avatar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("avatar request returned status %d", resp.StatusCode)
	}

	// Limit response size to 5MB to prevent memory exhaustion
	const maxSize = 5 << 20 // 5MB
	return io.ReadAll(io.LimitReader(resp.Body, maxSize))
}
