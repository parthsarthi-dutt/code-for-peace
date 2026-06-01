package database

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/parthsarthi-dutt/online-judge/server/models"
	"github.com/parthsarthi-dutt/online-judge/server/queue/producer"
	"github.com/parthsarthi-dutt/online-judge/server/repository"
)


func UserSubmission(
	submissionID string,
	problemID string,
	code string,
	userID string,
	language string,
	priority string,
) {

	slog.Info("Submission Recorded", slog.String("submission_id", submissionID))

	// currentTime := time.Now()
	// timeStamp := currentTime.Format("2006-01-02 15:04:05")
	submission := models.Submission{
		SubmissionID: submissionID,
		ProblemID:    problemID,
		UserID:       userID,
		Language:     language,
		Code:         code,
		CreatedAt:    time.Now(),
		Verdict:      "pending",
		ExecutionTime: 0,
		MemoryUsed:    0,
		Message:       "NA",
		Priority:      priority,
	}

	err := repository.CreateSubmission(submission)
	if err != nil {
		slog.Error("DB insert error", slog.String("error", err.Error()), slog.String("submission_id", submissionID))
		return
	}

	slog.Debug("Submission stored in DB", slog.String("submission_id", submissionID))

	producer.PushSubmission(RDB, submissionID, priority)

	slog.Debug("Submission pushed to queue", slog.String("submission_id", submissionID))
}


func GetInfo(submissionID string) []string {

	data, err := repository.GetSubmission(submissionID)
	if err != nil {
		slog.Error("GetInfo: submission not found", slog.String("submission_id", submissionID))
		return nil
	}

	ans := make([]string, 4)

	ans[0] = data.ProblemID
	ans[1] = data.Code
	ans[2] = data.Language
	ans[3] = data.UserID
	slog.Debug("Submission Info Returned", slog.String("submission_id", submissionID))
	return ans
}

func Result(
	submissionID string,
	verdict string,
	executionTime int64,
	memoryUsed int64,
	message string,
	tokensAwarded int,
) {

	err := repository.UpdateSubmissionResult(
		submissionID,
		verdict,
		executionTime,
		memoryUsed,
		message,
		tokensAwarded,
	)

	if err != nil {
		slog.Error("DB update error", slog.String("error", err.Error()), slog.String("submission_id", submissionID))
		return
	}

	slog.Info("Submission result updated", slog.String("submission_id", submissionID))
}
func GetSubmission(submissionID string) (models.Submission, error) {

	submission, err := repository.GetSubmission(submissionID)
	if err != nil {
		return models.Submission{}, fmt.Errorf("submission not found: %w", err)
	}

	return submission, nil
}
func GetAllSubmissions(userID string) ([]models.Submission, error) {

	submissions, err := repository.GetAllSubmissions(userID)
	if err != nil {
		return nil, fmt.Errorf("submissions not found: %w", err)
	}

	return submissions, nil
}
