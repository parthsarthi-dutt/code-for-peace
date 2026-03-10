package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/parthsarthi-dutt/online-judge/server/database"
	"github.com/parthsarthi-dutt/online-judge/server/judge"
	"github.com/parthsarthi-dutt/online-judge/server/models"
	"github.com/redis/go-redis/v9"
)

func StartWorker(rdb *redis.Client) {
	log.Println("Worker Started")
	maxWorkers := 4
	sem := make(chan struct{}, maxWorkers)
	submissionPath:="server/workers/submissions/"
	problemPath:="problems/"
	for{
	result, err := rdb.BRPop(context.Background(), 1*time.Second,
		"contestSubmission",
		"practiceSubmission",
		"miscellaneousSubmission",
	).Result()
		
		if(err!=nil){continue}
		log.Println("Submission Poped")
		jobData:=result[1]
		var job models.JudgeResult

	
		json.Unmarshal([]byte(jobData),&job)
				sem <- struct{}{} 
		go func(j models.JudgeResult) {
    defer func() { <-sem }()

    log.Println("Processing submission:", j.SubmissionID)
    info := database.GetInfo(j.SubmissionID)
    constraints := database.GetTimeAndMemory(info[0])

    // GUARANTEE CLEANUP: This will always run when the goroutine finishes!
    defer judge.DeleteFile(submissionPath, j.SubmissionID)

    // Make sure we start fresh, just in case
    judge.DeleteFile(submissionPath, j.SubmissionID)
    judge.CreateAndWriteFile(info[1], submissionPath, j.SubmissionID)
    
    submissionDir := submissionPath + j.SubmissionID
          
    
    ok:=judge.RunFile(submissionDir)
	if ok{
					output, err := judge.ExecuteBinary(
				problemPath+info[0],
				submissionDir,
				constraints[0],
				constraints[1],
			)
			// judge.DeleteFile(submissionPath,j.SubmissionID)
			if err != nil {
				log.Println(err)
				return
			}

			database.Result(
				j.SubmissionID,
				output.Verdict,
				output.ExecutionTime,
				output.MemoryUsed,
				output.Message,
			)

			log.Println("Result Updated")
			log.Println()
	}
    // if !flag { 
    //     return // DeleteFile will still trigger automatically because of defer!
    // }


		}(job)

	}
	
}