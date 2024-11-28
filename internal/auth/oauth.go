package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var githubConfig = &oauth2.Config{
	ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
	ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	Endpoint:     github.Endpoint,
	RedirectURL:  "http://localhost:8080/auth/github/callback",
	Scopes:       []string{"user:email"},
}

var googleConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Endpoint:     google.Endpoint,
	RedirectURL:  "http://localhost:8080/auth/google/callback",
	Scopes:       []string{"email", "profile"},
}

// GetOAuthURL generates an OAuth URL for the provider
func GetOAuthURL(provider string) (string, error) {
	switch provider {
	case "github":
		return githubConfig.AuthCodeURL("state"), nil
	case "google":
		return googleConfig.AuthCodeURL("state"), nil
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

// HandleOAuthCallback handles OAuth2 callback and returns the user ID
func HandleOAuthCallback(provider string, code string) (string, error) {
	var config *oauth2.Config
	var userInfoURL string

	switch provider {
	case "github":
		config = githubConfig
		userInfoURL = "https://api.github.com/user"
	case "google":
		config = googleConfig
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return "", err
	}

	client := config.Client(context.Background(), token)
	resp, err := client.Get(userInfoURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var userData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		return "", err
	}

	var userID string
	switch provider {
	case "github":
		userID = fmt.Sprintf("github_%v", userData["id"])
	case "google":
		userID = fmt.Sprintf("google_%v", userData["id"])
	}

	return userID, nil
}
