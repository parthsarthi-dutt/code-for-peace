package repository

import (
	"context"
	"log"

	"github.com/parthsarthi-dutt/online-judge/server/database/postgres"
	"github.com/parthsarthi-dutt/online-judge/server/models"
)

func GetUserByOAuthID(oauthID string) (*models.User, error) {

	query := `
	SELECT id, oauth_provider, oauth_id, email, username, avatar_url
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
	)

	// if err == sql.ErrNoRows {
	// 	return nil, nil
	// }

	if err != nil {
		log.Println(err)
		return nil, nil
	}

	return &user, nil
}

func CreateUser(user models.User) (*models.User, error) {

	query := `
	INSERT INTO users (oauth_provider, oauth_id, email, username, avatar_url)
	VALUES ($1,$2,$3,$4,$5)
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
	).Scan(&user.ID)

	if err != nil {
		return nil, err
	}

	return &user, nil
}