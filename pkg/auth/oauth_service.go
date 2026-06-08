package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/oauth2"
)

// OAuthProvider defines methods to communicate with Google/GitHub
type OAuthProvider interface {
	GetAuthURL(state string) string
	GetUserProfile(code string) (*UserProfile, error)
}

// UserProfile is a normalized structure for data from different providers
type UserProfile struct {
	Email     string
	Name      string
	AvatarURL string
	Provider  string // "google" or "github"
}

// GoogleOAuth handles Google specific logic
type GoogleOAuth struct {
	Config *oauth2.Config
}

func NewGoogleOAuth(clientID, clientSecret, redirectURL string) *GoogleOAuth {
	return &GoogleOAuth{
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
		},
	}
}

func (g *GoogleOAuth) GetAuthURL(state string) string {
	return g.Config.AuthCodeURL(state)
}

func (g *GoogleOAuth) GetUserProfile(code string) (*UserProfile, error) {
	// 1. Exchange code for token
	token, err := g.Config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	// 2. Fetch user info using token
	client := g.Config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	// 3. Parse Google response
	var googleUser struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(data, &googleUser); err != nil {
		return nil, err
	}

	return &UserProfile{
		Email:     googleUser.Email,
		Name:      googleUser.Name,
		AvatarURL: googleUser.Picture,
		Provider:  "google",
	}, nil
}

// GitHubOAuth handles GitHub specific logic
type GitHubOAuth struct {
	Config *oauth2.Config
}

func NewGitHubOAuth(clientID, clientSecret, redirectURL string) *GitHubOAuth {
	return &GitHubOAuth{
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"user:email", "read:user"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
		},
	}
}

func (gh *GitHubOAuth) GetAuthURL(state string) string {
	return gh.Config.AuthCodeURL(state)
}

func (gh *GitHubOAuth) GetUserProfile(code string) (*UserProfile, error) {
	token, err := gh.Config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}

	client := gh.Config.Client(context.Background(), token)

	// Get user profile
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var githubUser struct {
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, err
	}

	// GitHub email might be null if private, we might need to fetch it separately
	if githubUser.Email == "" {
		resp, err = client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer resp.Body.Close()
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&emails); err == nil {
				for _, e := range emails {
					if e.Primary {
						githubUser.Email = e.Email
						break
					}
				}
			}
		}
	}

	return &UserProfile{
		Email:     githubUser.Email,
		Name:      githubUser.Name,
		AvatarURL: githubUser.AvatarURL,
		Provider:  "github",
	}, nil
}
