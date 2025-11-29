package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/nobuww/simpel-ktp/internal/router"
	"github.com/nobuww/simpel-ktp/internal/store"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to db: %v\n", err)
	}
	defer dbPool.Close()

	queryStore := store.New(dbPool)

	r := router.New(queryStore)

	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", r)
}
