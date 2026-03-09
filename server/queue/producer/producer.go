package producer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/parthsarthi-dutt/online-judge/server/models"
	"github.com/redis/go-redis/v9"
)

func PushSubmission(rdb *redis.Client,submissionID string,priorityType string)error{
	job:=models.JudgeResult{
		SubmissionID: submissionID,

	}
	data,err:=json.Marshal(job)
	if(err!=nil){
		return err
	}
	fmt.Println("Submission Pushed to Queue")
	if(priorityType=="practice"){
	return rdb.LPush(context.Background(),"practiceSubmission",data).Err()}else if(priorityType=="contest"){
		return rdb.LPush(context.Background(),"contestSubmission",data).Err()
	}else{
		return rdb.LPush(context.Background(),"miscellaneousSubmission",data).Err()
	}
	
}