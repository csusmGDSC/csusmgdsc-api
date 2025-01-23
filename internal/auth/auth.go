package auth

import (
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

func GetGitHubConfig() *oauth2.Config {
	return githubConfig
}

func GetGoogleConfig() *oauth2.Config {
	return googleConfig
}
