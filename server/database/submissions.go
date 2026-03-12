package database

import (
	"log"
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

	log.Println("Submission Recorded")

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
		log.Println("DB insert error:", err)
		return
	}

	log.Println("Submission stored in DB")

	producer.PushSubmission(RDB, submissionID, priority)

	log.Println("Submission pushed to queue")
}


func GetInfo(submissionID string) []string {

	data, err := repository.GetSubmission(submissionID)
	if err!=nil {
		panic("submission not found")
	}

	ans:=make([]string,2)

	ans[0]=data.ProblemID
	ans[1]=data.Code
	log.Println("Submission Info Returned")
	return ans
}

func Result(
	submissionID string,
	verdict string,
	executionTime int64,
	memoryUsed int64,
	message string,
) {

	err := repository.UpdateSubmissionResult(
		submissionID,
		verdict,
		executionTime,
		memoryUsed,
		message,
	)

	if err != nil {
		log.Println("DB update error:", err)
		return
	}

	log.Println("Submission result updated")
}
func GetSubmission(submissionID string) (models.Submission, error){

	submission, err := repository.GetSubmission(submissionID)
	if err!=nil {
		panic("submission not found")
	}

	return submission, nil
}
func GetAllSubmissions(userID string) ([]models.Submission, error){

	submission, err := repository.GetAllSubmissions(userID)
	if err!=nil {
		panic("submission not found")
	}

	return submission, nil
}
