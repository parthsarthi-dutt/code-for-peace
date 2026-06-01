package auth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"os"
)

var GoogleConfig *oauth2.Config

func InitGoogleOAuth() {
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/auth/google/callback"
	}
	GoogleConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  redirectURL,

		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},

		Endpoint: google.Endpoint,
	}
}