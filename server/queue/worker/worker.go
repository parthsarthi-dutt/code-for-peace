package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/parthsarthi-dutt/online-judge/server/database"
	"github.com/parthsarthi-dutt/online-judge/server/judge"
	"github.com/parthsarthi-dutt/online-judge/server/queue"
	"github.com/parthsarthi-dutt/online-judge/server/repository"
	"github.com/redis/go-redis/v9"
)

// consumerName uniquely identifies THIS worker instance within the consumer group.
// If you run 3 worker processes, each gets a unique name like "DESKTOP-ABC-1234".
// Redis uses this to track which messages each consumer has received.
var consumerName string

func init() {
	hostname, _ := os.Hostname()
	consumerName = fmt.Sprintf("%s-%d", hostname, os.Getpid())
}

func StartWorker(rdb *redis.Client) {
	slog.Info("Worker started", slog.String("consumer", consumerName))

	maxWorkers := 4
	sem := make(chan struct{}, maxWorkers)
	submissionPath := "server/workers/submissions/"
	problemPath := "problems/"

	// ──────────────────────────────────────────────────────────
	// STEP 1: Recover pending messages from a previous crash.
	//
	// WHY: If this worker (or another worker) crashed while processing
	// a submission, that message is stuck in the Pending Entries List (PEL).
	// We reclaim it here so no submission is ever lost.
	//
	// Interview answer: "On startup, I call XPENDING to find orphaned
	// messages and XCLAIM to reassign them to the current worker."
	// ──────────────────────────────────────────────────────────
	recoverPending(rdb, sem, submissionPath, problemPath)

	// ──────────────────────────────────────────────────────────
	// STEP 2: Main loop — read new messages from streams
	//
	// XREADGROUP replaces BRPOP. Key differences:
	//   BRPOP  → deletes message immediately, gone forever
	//   XREADGROUP → assigns message to this consumer, stays in stream
	//
	// The ">" ID means "give me only NEW messages never seen by any
	// consumer in this group."
	// ──────────────────────────────────────────────────────────
	for {
		// Build the Streams arg: [stream1, stream2, stream3, ">", ">", ">"]
		// The ">" after each stream means "only new, undelivered messages"
		streamIDs := make([]string, 0, len(queue.AllStreams)*2)
		streamIDs = append(streamIDs, queue.AllStreams...)
		for range queue.AllStreams {
			streamIDs = append(streamIDs, ">")
		}

		entries, err := rdb.XReadGroup(context.Background(), &redis.XReadGroupArgs{
			Group:    queue.ConsumerGroup,
			Consumer: consumerName,
			Streams:  streamIDs,
			Count:    1,       // Process one at a time for fairness
			Block:    1 * time.Second, // Block up to 1s waiting for messages
		}).Result()

		if err != nil {
			// Timeout (no messages) — just loop again. This is normal.
			continue
		}

		// entries is a list of XStream, each containing messages from one stream
		for _, stream := range entries {
			for _, message := range stream.Messages {
				// Extract submission_id from the stream message
				submissionID, ok := message.Values["submission_id"].(string)
				if !ok {
					slog.Error("Invalid message format, ACKing and skipping", slog.String("message_id", message.ID))
					rdb.XAck(context.Background(), stream.Stream, queue.ConsumerGroup, message.ID)
					continue
				}

				sem <- struct{}{} // Block if all 4 workers are busy
				go processSubmission(
					rdb, sem,
					stream.Stream, message.ID, submissionID,
					submissionPath, problemPath,
				)
			}
		}
	}
}

// processSubmission handles a single submission: compile, run, judge, store result.
// This is the function that runs inside each worker goroutine.
//
// CRITICAL: Notice the two defers at the top:
//   1. Release semaphore slot (so another goroutine can start)
//   2. Recover from panics (so one bad submission doesn't kill the pool)
func processSubmission(
	rdb *redis.Client,
	sem chan struct{},
	streamName string,
	messageID string,
	submissionID string,
	submissionPath string,
	problemPath string,
) {
	// Release the semaphore slot when this goroutine finishes
	defer func() { <-sem }()

	// ──────────────────────────────────────────────────────────
	// PANIC RECOVERY — this is what your current code is missing!
	//
	// Without this: one panic (e.g., nil pointer in judge logic)
	// kills the ENTIRE worker process. All 4 workers die.
	//
	// With this: the panic is caught, logged, the message goes
	// to the dead-letter queue, and the worker keeps running.
	//
	// Interview answer: "I use defer/recover to isolate failures.
	// A single toxic submission cannot bring down the worker pool."
	// ──────────────────────────────────────────────────────────
	defer func() {
		if r := recover(); r != nil {
			slog.Error("RECOVERED from panic processing", slog.String("submission_id", submissionID), slog.Any("panic", r))
			// Move to dead-letter queue for investigation
			rdb.XAdd(context.Background(), &redis.XAddArgs{
				Stream: queue.StreamDLQ,
				Values: map[string]interface{}{
					"submission_id": submissionID,
					"error":         fmt.Sprintf("%v", r),
					"source_stream": streamName,
				},
			})
			// ACK the original so it doesn't retry forever
			rdb.XAck(context.Background(), streamName, queue.ConsumerGroup, messageID)
		}
	}()

	slog.Info("Processing submission", slog.String("submission_id", submissionID))

	info := database.GetInfo(submissionID)
	if info == nil {
		slog.Error("Submission not found in DB", slog.String("submission_id", submissionID))
		rdb.XAck(context.Background(), streamName, queue.ConsumerGroup, messageID)
		return
	}
	constraints := database.GetTimeAndMemory(info[0])

	language := info[2] // DB returns [ProblemID, Code, Language, UserID]
	userID := info[3]
	langConfig := judge.GetLangConfig(language)

	// Cleanup files when done (same as before)
	defer judge.DeleteFile(submissionPath, submissionID)

	judge.DeleteFile(submissionPath, submissionID)
	judge.CreateAndWriteFile(info[1], submissionPath, submissionID, langConfig.FileExtension)
	submissionDir := submissionPath + submissionID

	ok, compileErr := judge.RunFile(submissionDir, langConfig)
	if ok {
		output, err := judge.ExecuteBinary(
			problemPath+info[0],
			submissionDir,
			constraints[0],
			constraints[1],
			langConfig,
		)
		if err != nil {
			slog.Error("Execution failed", slog.String("error", err.Error()), slog.String("submission_id", submissionID))
			rdb.XAck(context.Background(), streamName, queue.ConsumerGroup, messageID)
			return
		}

		tokensAwarded := 0
		// Token Economy: Reward +2 tokens if Accepted
		if output.Verdict == "Accepted" {
			// Ensure we only reward tokens if the user hasn't already solved this problem
			alreadySolved, err := repository.HasUserSolvedProblem(userID, info[0], submissionID)
			if err == nil && !alreadySolved {
				tokensAwarded = 2
				repository.UpdateTokens(userID, 2)
				repository.UpdateUserStreak(userID)

				totalSolved, _ := repository.GetTotalSolvedProblems(userID)
				newTotal := totalSolved + 1 // Add 1 because current submission isn't in DB yet

				bonus := 0
				switch newTotal {
				case 10:
					bonus = 20
				case 50:
					bonus = 50
				case 100:
					bonus = 100
				case 500:
					bonus = 500
				case 1000:
					bonus = 1000
				}

				if bonus > 0 {
					repository.UpdateTokens(userID, bonus)
					tokensAwarded += bonus
					slog.Info("Milestone reached!", slog.Int("milestone", newTotal), slog.Int("bonus", bonus))
				}

				slog.Info("Rewarded tokens and updated streak for first accepted submission", slog.String("user_id", userID), slog.String("problem_id", info[0]))
			}
		}

		database.Result(
			submissionID,
			output.Verdict,
			output.ExecutionTime,
			output.MemoryUsed,
			output.Message,
			tokensAwarded,
		)
		slog.Info("Result Updated", slog.String("submission_id", submissionID))
	} else {
		// Handle Compilation Error
		database.Result(
			submissionID,
			"Compilation Error",
			0,
			0,
			compileErr,
			0,
		)
		slog.Info("Compilation Error Recorded in DB", slog.String("submission_id", submissionID))
	}

	// ──────────────────────────────────────────────────────────
	// XACK — the most important line in the entire upgrade!
	//
	// This tells Redis: "I successfully processed this message.
	// Remove it from the Pending Entries List."
	//
	// Without XACK: Redis thinks we're still working on it.
	// On restart, recoverPending() would reprocess it.
	//
	// Interview answer: "XACK is like a Kafka offset commit.
	// It confirms successful processing and prevents redelivery."
	// ──────────────────────────────────────────────────────────
	rdb.XAck(context.Background(), streamName, queue.ConsumerGroup, messageID)
	slog.Info("Message ACKed", slog.String("message_id", messageID), slog.String("stream", streamName))
}

// recoverPending finds messages that were assigned to workers but never ACKed.
// This happens when a worker crashes between XREADGROUP and XACK.
//
// Uses XPENDING to list orphaned messages, then XCLAIM to take ownership.
// If a message has been retried > 3 times, it goes to the dead-letter queue.
func recoverPending(rdb *redis.Client, sem chan struct{}, submissionPath, problemPath string) {
	ctx := context.Background()

	for _, stream := range queue.AllStreams {
		pending, err := rdb.XPendingExt(ctx, &redis.XPendingExtArgs{
			Stream: stream,
			Group:  queue.ConsumerGroup,
			Start:  "-",
			End:    "+",
			Count:  100,
		}).Result()

		if err != nil || len(pending) == 0 {
			continue
		}

		slog.Info("Found pending messages", slog.Int("count", len(pending)), slog.String("stream", stream))

		for _, p := range pending {
			// If retried too many times → dead-letter queue
			if p.RetryCount > 3 {
				slog.Warn("Message exceeded 3 retries → DLQ", slog.String("message_id", p.ID))
				msgs, _ := rdb.XRange(ctx, stream, p.ID, p.ID).Result()
				if len(msgs) > 0 {
					rdb.XAdd(ctx, &redis.XAddArgs{
						Stream: queue.StreamDLQ,
						Values: map[string]interface{}{
							"submission_id": msgs[0].Values["submission_id"],
							"error":         "exceeded_retry_limit",
							"source_stream": stream,
						},
					})
				}
				rdb.XAck(ctx, stream, queue.ConsumerGroup, p.ID)
				continue
			}

			// XCLAIM: take ownership of this pending message
			claimed, err := rdb.XClaim(ctx, &redis.XClaimArgs{
				Stream:   stream,
				Group:    queue.ConsumerGroup,
				Consumer: consumerName,
				MinIdle:  0,
				Messages: []string{p.ID},
			}).Result()

			if err != nil || len(claimed) == 0 {
				continue
			}

			submissionID, ok := claimed[0].Values["submission_id"].(string)
			if !ok {
				rdb.XAck(ctx, stream, queue.ConsumerGroup, p.ID)
				continue
			}

			slog.Info("Reclaimed pending message", slog.String("message_id", p.ID), slog.String("submission_id", submissionID))
			sem <- struct{}{}
			go processSubmission(rdb, sem, stream, p.ID, submissionID, submissionPath, problemPath)
		}
	}
}