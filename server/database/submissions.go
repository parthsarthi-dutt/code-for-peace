package database

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/parthsarthi-dutt/online-judge/server/models"
	"github.com/parthsarthi-dutt/online-judge/server/queue/producer"
)
var submissionsDB = make(map[string][]byte)

func UserSubmission(submissionID string,problemID string,code string,userID string,language string,priority string){
	log.Println("Submission Recorded")
	currentTime:=time.Now()
	timeStamp := currentTime.Format("2006-01-02 15:04:05")
	submission:=models.Submission{
		SubmissionID:submissionID,
		ProblemID:problemID,
		UserID:userID,
		Language:language,
		Code:code,
		CreatedAt: timeStamp,
		Verdict: "pending",
		ExecutionTime:0,
		MemoryUsed:0,
		Message: "NA",
		Priority: priority,
	}
	log.Println(submission.UserID)
	log.Println(submission.CreatedAt)
	data,err:=json.Marshal(submission)
	if(err!=nil){
		panic(err)
	}
	submissionsDB[submissionID]=data
	log.Println("Submission Pushed to Queue")


	producer.PushSubmission(RDB,submissionID,priority)
	
}
func GetInfo(submissionID string) []string {

	data, exists := submissionsDB[submissionID]
	if !exists {
		panic("submission not found")
	}

	var submission models.Submission
	ans:=make([]string,2)
	err := json.Unmarshal(data, &submission)
	if err != nil {
		panic(err)
	}
	ans[0]=submission.ProblemID
	ans[1]=submission.Code
	log.Println("Submission Info Returned")
	return ans
}
func Result(submissionID string,verdict string,executionTime int64,memoryUsed int64,message string){
	data, exists := submissionsDB[submissionID]
	if !exists {
		panic("submission not found")
	}
	var submission models.Submission
	err := json.Unmarshal(data, &submission)
		if err != nil {
		panic(err)
	}
	submission.Verdict=verdict
	submission.ExecutionTime=executionTime
	submission.MemoryUsed=memoryUsed
	submission.Message=message
		// Convert back to JSON
	updatedData, err := json.Marshal(submission)
	if err != nil {
		panic(err)
	}

	// Store back in map
	submissionsDB[submissionID] = updatedData
	log.Println(submission.Verdict)
	log.Println(submission.ExecutionTime)
	log.Println(submission.MemoryUsed)
	log.Println(submission.Message)
}
func GetSubmission(submissionID string) (models.Submission, error){

data, exists := submissionsDB[submissionID]
	if !exists {
		return models.Submission{}, fmt.Errorf("submission not found")
	}

	var submission models.Submission

	err := json.Unmarshal(data, &submission)
	if err != nil {
		return models.Submission{}, err
	}

	return submission, nil
}
