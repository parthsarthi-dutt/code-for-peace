package auth

import (
	"context"
	"encoding/json"
	"log"
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

log.Println("Lookup result:", user, "error:", err)

if err != nil {
    log.Println("DB ERROR:", err)
    http.Error(w, "Database error", http.StatusInternalServerError)
    return
}

if user == nil {
    log.Println("User not found, creating user...")
}

	// if user doesn't exist → create
if user == nil {

    log.Println("Creating user with OAuthID:", oauthID)

    newUser := models.User{
        OauthProvider: "google",
        OauthID:       oauthID,
        Email:         email,
        Username:      name,
        AvatarURL:     picture,
    }

    user, err = repository.CreateUser(newUser)

    if err != nil {
        log.Println("CreateUser error:", err)
        http.Error(w, "User creation failed", http.StatusInternalServerError)
        return
    }

    log.Println("User created with ID:", user.ID)
}

	// generate JWT

	jwtToken, err := GenerateJWT(user.ID, user.Username)
	if err != nil {
		http.Error(w, "JWT generation failed", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": jwtToken,
		"user":  user,
	})
}