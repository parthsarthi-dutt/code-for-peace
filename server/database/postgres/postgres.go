package postgres

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func ConnectDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://judgeuser:judgepass@127.0.0.1:5435/onlinejudge?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		slog.Error("Unable to connect to DB", slog.String("error", err.Error()))
		os.Exit(1)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		slog.Error("DB ping failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	DB = pool
	slog.Info("PostgreSQL connected")
}