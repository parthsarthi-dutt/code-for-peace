package models

type User struct {
	ID            int
	OauthProvider string
	OauthID       string
	Email         string
	Username      string
	AvatarURL     string
}