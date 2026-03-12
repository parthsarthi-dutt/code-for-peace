package repository

import (
	"context"

	"github.com/parthsarthi-dutt/online-judge/server/database/postgres"
	"github.com/parthsarthi-dutt/online-judge/server/models"
)

func CreateSubmission(s models.Submission) error {

	query := `
	INSERT INTO submissions
	(submission_id, problem_id, user_id, language, code, verdict,
	 execution_time, memory_used, message, priority, created_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`

	_, err := postgres.DB.Exec(
		context.Background(),
		query,
		s.SubmissionID,
		s.ProblemID,
		s.UserID,
		s.Language,
		s.Code,
		s.Verdict,
		s.ExecutionTime,
		s.MemoryUsed,
		s.Message,
		s.Priority,
		s.CreatedAt,
	)

	return err
}

func GetSubmission(id string) (models.Submission, error) {

	var s models.Submission

	query := `SELECT code, language, problem_id, user_id, verdict, created_at, execution_time, memory_used, message  FROM submissions WHERE submission_id=$1`

	err := postgres.DB.QueryRow(
		context.Background(),
		query,
		id,
	).Scan(&s.Code, &s.Language, &s.ProblemID, &s.UserID, &s.Verdict, &s.CreatedAt, &s.ExecutionTime, &s.MemoryUsed, &s.Message)

	return s, err
}
func GetAllSubmissions(id string) ([]models.Submission, error) {

	query := `
	SELECT code, language, problem_id, submission_id,
	       verdict, created_at, execution_time,
	       memory_used, message
	FROM submissions
	WHERE user_id=$1
	`

	rows, err := postgres.DB.Query(
		context.Background(),
		query,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []models.Submission

	for rows.Next() {

		var s models.Submission

		err := rows.Scan(
			&s.Code,
			&s.Language,
			&s.ProblemID,
			&s.SubmissionID,
			&s.Verdict,
			&s.CreatedAt,
			&s.ExecutionTime,
			&s.MemoryUsed,
			&s.Message,
		)

		if err != nil {
			return nil, err
		}

		submissions = append(submissions, s)
	}

	return submissions, nil
}

func UpdateSubmissionResult(
	id string,
	verdict string,
	executionTime int64,
	memoryUsed int64,
	message string,
) error {

	query := `
	UPDATE submissions
	SET verdict=$1,
	    execution_time=$2,
	    memory_used=$3,
	    message=$4
	WHERE submission_id=$5
	`

	_, err := postgres.DB.Exec(
		context.Background(),
		query,
		verdict,
		executionTime,
		memoryUsed,
		message,
		id,
	)

	return err
}