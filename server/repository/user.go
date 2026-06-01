package repository

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/parthsarthi-dutt/online-judge/server/database/postgres"
	"github.com/parthsarthi-dutt/online-judge/server/models"
)

func GetUserByOAuthID(oauthID string) (*models.User, error) {

	query := `
	SELECT id, oauth_provider, oauth_id, email, username, avatar_url, tokens, current_streak, highest_streak, last_solved_date
	FROM users
	WHERE oauth_id = $1
	`

	row := postgres.DB.QueryRow(context.Background(),query, oauthID)

	var user models.User

	err := row.Scan(
		&user.ID,
		&user.OauthProvider,
		&user.OauthID,
		&user.Email,
		&user.Username,
		&user.AvatarURL,
		&user.Tokens,
		&user.CurrentStreak,
		&user.HighestStreak,
		&user.LastSolvedDate,
	)

	// if err == sql.ErrNoRows {
	// 	return nil, nil
	// }

	if err != nil {
		slog.Error("Database constraint or internal error", slog.String("error", err.Error()))
		return nil, nil
	}

	return &user, nil
}

func CreateUser(user models.User) (*models.User, error) {

	query := `
	INSERT INTO users (oauth_provider, oauth_id, email, username, avatar_url, tokens, current_streak, highest_streak)
	VALUES ($1,$2,$3,$4,$5,$6,0,0)
	RETURNING id
	`

	err := postgres.DB.QueryRow(
		context.Background(),
		query,
		user.OauthProvider,
		user.OauthID,
		user.Email,
		user.Username,
		user.AvatarURL,
		user.Tokens,
	).Scan(&user.ID)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UpdateTokens(userID string, amount int) error {
	id, err := strconv.Atoi(userID)
	if err != nil {
		return err
	}
	query := `
	UPDATE users
	SET tokens = tokens + $1
	WHERE id = $2
	`
	_, err = postgres.DB.Exec(context.Background(), query, amount, id)
	return err
}

func GetUserByID(userID string) (*models.User, error) {
	id, err := strconv.Atoi(userID)
	if err != nil {
		return nil, err
	}
	query := `
	SELECT id, oauth_provider, oauth_id, email, username, avatar_url, tokens, current_streak, highest_streak, last_solved_date
	FROM users
	WHERE id = $1
	`
	row := postgres.DB.QueryRow(context.Background(), query, id)
	var user models.User
	err = row.Scan(
		&user.ID,
		&user.OauthProvider,
		&user.OauthID,
		&user.Email,
		&user.Username,
		&user.AvatarURL,
		&user.Tokens,
		&user.CurrentStreak,
		&user.HighestStreak,
		&user.LastSolvedDate,
	)
	if err != nil {
		slog.Error("Failed to get user by ID", slog.String("error", err.Error()))
		return nil, err
	}
	return &user, nil
}

func UpdateUserStreak(userID string) error {
	id, err := strconv.Atoi(userID)
	if err != nil {
		return err
	}
	query := `
	UPDATE users
	SET 
		current_streak = CASE 
			WHEN last_solved_date IS NULL OR last_solved_date < CURRENT_DATE - INTERVAL '1 day' THEN 1
			WHEN last_solved_date >= CURRENT_DATE - INTERVAL '1 day' AND last_solved_date < CURRENT_DATE THEN current_streak + 1
			ELSE current_streak 
		END,
		highest_streak = GREATEST(highest_streak, 
			CASE 
				WHEN last_solved_date IS NULL OR last_solved_date < CURRENT_DATE - INTERVAL '1 day' THEN 1
				WHEN last_solved_date >= CURRENT_DATE - INTERVAL '1 day' AND last_solved_date < CURRENT_DATE THEN current_streak + 1
				ELSE current_streak 
			END
		),
		last_solved_date = CURRENT_TIMESTAMP
	WHERE id = $1
	`
	_, err = postgres.DB.Exec(context.Background(), query, id)
	if err != nil {
		slog.Error("Failed to update user streak", slog.String("error", err.Error()))
	}
	return err
}

func UpdateUserProfile(userID string, username string, avatarURL string) error {
	id, err := strconv.Atoi(userID)
	if err != nil {
		return err
	}
	query := `
	UPDATE users
	SET username = $1, avatar_url = $2
	WHERE id = $3
	`
	_, err = postgres.DB.Exec(context.Background(), query, username, avatarURL, id)
	if err != nil {
		slog.Error("Failed to update user profile", slog.String("error", err.Error()))
	}
	return err
}