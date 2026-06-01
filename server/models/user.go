package models

import "time"

type User struct {
	ID            int
	OauthProvider string
	OauthID       string
	Email         string
	Username      string
	AvatarURL     string    `json:"avatar_url"`
	Tokens        int       `json:"tokens"`
	CurrentStreak int       `json:"current_streak"`
	HighestStreak int       `json:"highest_streak"`
	LastSolvedDate *time.Time `json:"last_solved_date"`
}