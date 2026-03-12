package postgres

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func ConnectDB() {
	dsn := "postgres://judgeuser:judgepass@127.0.0.1:5435/onlinejudge?sslmode=disable"

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Unable to connect to DB:", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatal("DB ping failed:", err)
	}

	DB = pool
	log.Println("PostgreSQL connected")
}