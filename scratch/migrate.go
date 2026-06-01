package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://judgeuser:judgepass@127.0.0.1:5435/onlinejudge?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to DB: %v\n", err)
	}
	defer pool.Close()

	// Create tables if they don't exist
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			oauth_provider TEXT,
			oauth_id TEXT UNIQUE,
			email TEXT,
			username TEXT,
			avatar_url TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS submissions (
			submission_id TEXT PRIMARY KEY,
			user_id TEXT,
			problem_id TEXT,
			code TEXT,
			language TEXT,
			verdict TEXT,
			execution_time BIGINT,
			memory_used BIGINT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			message TEXT,
			priority TEXT
		);`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS tokens INT DEFAULT 0;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS current_streak INT DEFAULT 0;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS highest_streak INT DEFAULT 0;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS last_solved_date TIMESTAMP;`,
		`ALTER TABLE submissions ADD COLUMN IF NOT EXISTS tokens_awarded INT DEFAULT 0;`,
		`CREATE TABLE IF NOT EXISTS unlocked_editorials (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			problem_id TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(user_id, problem_id)
		);`,
		`CREATE TABLE IF NOT EXISTS ai_interviews (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			level TEXT NOT NULL,
			duration INT NOT NULL,
			tokens_deducted INT NOT NULL,
			started_at TIMESTAMP DEFAULT NOW(),
			status TEXT DEFAULT 'active',
			history JSONB DEFAULT '[]'::jsonb,
			feedback TEXT DEFAULT ''
		);`,
	}

	for _, q := range queries {
		_, err := pool.Exec(context.Background(), q)
		if err != nil {
			log.Printf("Error executing query: %s\n%v\n", q, err)
		} else {
			fmt.Println("Success:", q)
		}
	}
	fmt.Println("Migration complete.")
}
