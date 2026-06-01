package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/parthsarthi-dutt/online-judge/server/models"
	"github.com/parthsarthi-dutt/online-judge/server/repository"
)

func GoogleLogin(w http.ResponseWriter, r *http.Request) {

	url := GoogleConfig.AuthCodeURL("state-token")

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}





func GoogleCallback(w http.ResponseWriter, r *http.Request) {

	code := r.URL.Query().Get("code")

	token, err := GoogleConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "OAuth exchange failed", http.StatusBadRequest)
		return
	}

	client := GoogleConfig.Client(context.Background(), token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	var userInfo map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&userInfo)

	oauthID := userInfo["id"].(string)
	email := userInfo["email"].(string)
	name := userInfo["name"].(string)
	picture := userInfo["picture"].(string)

	// check if user exists

user, err := repository.GetUserByOAuthID(oauthID)

slog.Debug("Lookup result", slog.Any("user", user), slog.Any("error", err))

if err != nil {
    slog.Error("Database error during OAuth lookup", slog.String("error", err.Error()), slog.String("oauth_id", oauthID))
    http.Error(w, "Database error", http.StatusInternalServerError)
    return
}

if user == nil {
    slog.Info("User not found, creating new user", slog.String("oauth_id", oauthID))
}

	// if user doesn't exist → create
if user == nil {

    newUser := models.User{
        OauthProvider: "google",
        OauthID:       oauthID,
        Email:         email,
        Username:      name,
        AvatarURL:     picture,
        Tokens:        20,
    }

    user, err = repository.CreateUser(newUser)

    if err != nil {
        slog.Error("Failed to create user", slog.String("error", err.Error()), slog.String("oauth_id", oauthID))
        http.Error(w, "User creation failed", http.StatusInternalServerError)
        return
    }

    slog.Info("User created successfully", slog.Int("user_id", user.ID))
}

	// generate JWT

	jwtToken, err := GenerateJWT(user.ID, user.Username)
	if err != nil {
		http.Error(w, "JWT generation failed", http.StatusInternalServerError)
		return
	}

	// Determine the frontend origin to redirect back to
	frontendOrigin := os.Getenv("FRONTEND_URL")
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:5173"
		if referer := r.Header.Get("Referer"); referer != "" {
			for _, candidate := range []string{"http://localhost:5174", "http://localhost:5173", "http://localhost:3000"} {
				if len(referer) >= len(candidate) && referer[:len(candidate)] == candidate {
					frontendOrigin = candidate
					break
				}
			}
		}
	}

	redirectURL := fmt.Sprintf(
		"%s/auth/callback?token=%s&user_id=%d&username=%s&avatar=%s&email=%s&tokens=%d",
		frontendOrigin,
		jwtToken,
		user.ID,
		user.Username,
		user.AvatarURL,
		user.Email,
		user.Tokens,
	)

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}