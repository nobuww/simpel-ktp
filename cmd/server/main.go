package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/nobuww/simpel-ktp/internal/router"
	"github.com/nobuww/simpel-ktp/internal/session"
	"github.com/nobuww/simpel-ktp/internal/store"
	"github.com/nobuww/simpel-ktp/internal/vite"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	if err := vite.Init(os.Getenv("GO_ENV")); err != nil {
		log.Fatalf("Failed to initialize vite: %v", err)
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

	// Verify database connection
	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	queryStore := store.New(dbPool)

	// Initialize session manager
	sessionMgr := session.New(os.Getenv("SESSION_SECRET"))

	r := router.New(queryStore, sessionMgr)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
