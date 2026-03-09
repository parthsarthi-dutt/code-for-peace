package database

import (
	// "context"
	// "fmt"

	"github.com/redis/go-redis/v9"

)


var RDB *redis.Client

func ConnectRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

