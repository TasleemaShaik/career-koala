package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	ckdb "career-koala/db"
	"career-koala/migrations"
)

func main() {
	action := "up"
	if len(os.Args) > 1 {
		action = strings.ToLower(strings.TrimSpace(os.Args[1]))
	}

	dsn := ckdb.DSNFromEnv()

	dbConn, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer dbConn.Close()
	if err := dbConn.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	switch action {
	case "up":
		if err := migrations.Up(ctx, dbConn); err != nil {
			log.Fatalf("migrations up: %v", err)
		}
		log.Println("migrations up: done")
	case "down":
		if err := migrations.Down(ctx, dbConn); err != nil {
			log.Fatalf("migrations down: %v", err)
		}
		log.Println("migrations down: done")
	default:
		log.Fatalf("unknown action %q (use up|down)", action)
	}
}
