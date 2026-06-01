package database

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/parthsarthi-dutt/online-judge/server/queue"
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func ConnectRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Create consumer groups for each stream on startup.
	// This MUST happen before any worker calls XREADGROUP.
	for _, stream := range queue.AllStreams {
		createStreamGroup(stream, queue.ConsumerGroup)
	}
	slog.Info("Redis Streams and consumer groups initialized")
}

// createStreamGroup creates a consumer group for a stream.
// MKSTREAM flag auto-creates the stream if it doesn't exist yet.
// "0" means the group starts reading from the beginning of the stream.
func createStreamGroup(stream, group string) {
	ctx := context.Background()
	err := RDB.XGroupCreateMkStream(ctx, stream, group, "0").Err()
	if err != nil {
		// "BUSYGROUP" means the group already exists — that's perfectly fine.
		if strings.Contains(err.Error(), "BUSYGROUP") {
			slog.Debug("Consumer group already exists", slog.String("group", group), slog.String("stream", stream))
			return
		}
		slog.Error("Failed to create consumer group", slog.String("group", group), slog.String("stream", stream), slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("Created consumer group", slog.String("group", group), slog.String("stream", stream))
}

func CheckRateLimit(key string, limit int, windowSeconds int) (bool, error) {
	ctx := context.Background()
	count, err := RDB.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		RDB.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)
	}
	if count > int64(limit) {
		return false, nil // Rate limited
	}
	return true, nil // Allowed
}

