package producer

import (
	"context"
	"log/slog"

	"github.com/parthsarthi-dutt/online-judge/server/queue"
	"github.com/redis/go-redis/v9"
)

// PushSubmission adds a submission to the appropriate Redis Stream.
//
// Before (Redis Lists):   LPUSH "practiceSubmission" {json}   → item is just "in the list"
// After  (Redis Streams):  XADD "practiceSubmission" * submission_id abc123
//
// The "*" tells Redis to auto-generate a unique message ID (timestamp-based).
// Unlike LPUSH, the message stays in the stream until a worker ACKs it.
func PushSubmission(rdb *redis.Client, submissionID string, priorityType string) error {

	// Map priority type to the correct stream name
	// Uses constants from database package — single source of truth
	stream := queue.StreamPractice
	switch priorityType {
	case "contest":
		stream = queue.StreamContest
	case "practice":
		stream = queue.StreamPractice
	default:
		stream = queue.StreamMisc
	}

	// XADD appends a new entry to the stream
	// Each entry is a set of field-value pairs (like a mini hash map)
	// We only need submission_id — the worker fetches full details from PostgreSQL
	err := rdb.XAdd(context.Background(), &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{
			"submission_id": submissionID,
		},
	}).Err()

	if err != nil {
		slog.Error("Failed to push submission to stream", slog.String("submission_id", submissionID), slog.String("stream", stream), slog.String("error", err.Error()))
		return err
	}

	slog.Info("Submission pushed to stream", slog.String("submission_id", submissionID), slog.String("stream", stream))
	return nil
}