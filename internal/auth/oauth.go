package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/csusmGDSC/csusmgdsc-api/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var (
	once         sync.Once
	githubConfig *oauth2.Config
	googleConfig *oauth2.Config
)

type OAuthUserData struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Email     *string `json:"email,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Provider  string  `json:"provider"`
}

func InitOAuth() {
	once.Do(func() {
		cfg := config.LoadConfig()

		githubConfig = &oauth2.Config{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			Endpoint:     github.Endpoint,
			RedirectURL:  cfg.OAuthRedirectUrl + "/auth/github/callback",
			Scopes:       []string{"user:email"},
		}

		googleConfig = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  cfg.OAuthRedirectUrl + "/auth/google/callback",
			Scopes:       []string{"email", "profile"},
		}
	})
}

// GetOAuthURL generates an OAuth URL for the provider
func GetOAuthURL(provider string) (string, error) {
	switch provider {
	case "github":
		if githubConfig == nil {
			return "", fmt.Errorf("github OAuth not initialized")
		}
		return githubConfig.AuthCodeURL("state"), nil
	case "google":
		if googleConfig == nil {
			return "", fmt.Errorf("google OAuth not initialized")
		}
		return googleConfig.AuthCodeURL("state"), nil
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

// HandleOAuthCallback handles OAuth2 callback and returns user data
func HandleOAuthCallback(provider string, code string) (*OAuthUserData, error) {
	var config *oauth2.Config
	var userInfoURL string

	switch provider {
	case "github":
		if githubConfig == nil {
			return nil, fmt.Errorf("github OAuth not initialized")
		}
		config = githubConfig
		userInfoURL = "https://api.github.com/user"
	case "google":
		if googleConfig == nil {
			return nil, fmt.Errorf("google OAuth not initialized")
		}
		config = googleConfig
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %v", err)
	}
	client := config.Client(context.Background(), token)
	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()

	var rawData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %v", err)
	}

	userData := &OAuthUserData{
		Provider: provider,
	}

	switch provider {
	case "github":
		// Handle required ID field
		if id, ok := rawData["id"].(float64); ok {
			userData.ID = fmt.Sprintf("%.0f", id)
		} else {
			return nil, fmt.Errorf("invalid or missing github user ID")
		}
		if email, ok := rawData["email"].(string); ok && email != "" {
			userData.Email = &email
		}
		if name, ok := rawData["name"].(string); ok && name != "" {
			userData.Name = name
		}
		if avatar, ok := rawData["avatar_url"].(string); ok {
			userData.AvatarURL = &avatar
		}
	case "google":
		if id, ok := rawData["id"].(string); ok && id != "" {
			userData.ID = id
		}
		if email, ok := rawData["email"].(string); ok && email != "" {
			userData.Email = &email
		}
		if name, ok := rawData["name"].(string); ok && name != "" {
			userData.Name = name
		}
		if picture, ok := rawData["picture"].(string); ok {
			userData.AvatarURL = &picture
		}
	}
	return userData, nil
}
