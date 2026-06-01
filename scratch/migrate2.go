package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := "postgres://judgeuser:judgepass@127.0.0.1:5435/onlinejudge?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to DB: %v\n", err)
	}
	defer pool.Close()

	q := `ALTER TABLE submissions ADD COLUMN IF NOT EXISTS tokens_awarded INT DEFAULT 0;`
	_, err = pool.Exec(context.Background(), q)
	if err != nil {
		log.Fatalf("Error executing query: %s\n%v\n", q, err)
	}
	fmt.Println("Success:", q)
}
